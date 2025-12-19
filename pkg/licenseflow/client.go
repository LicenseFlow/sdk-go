package licenseflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	BaseURL   string
	APIKey    string
	JWTSecret string
	Timeout   time.Duration
}

type Client struct {
	config     Config
	httpClient *http.Client
	cache      map[string]interface{}
}

func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache: make(map[string]interface{}),
	}
}

func (c *Client) GetHardwareID() string {
	hostname, _ := os.Hostname()
	return hostname
}

func (c *Client) Activate(licenseKey string, deviceName string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"license_key": licenseKey,
		"device_id":   c.GetHardwareID(),
		"device_name": deviceName,
	}
	return c.post("functions/v1/activate-license", payload)
}

func (c *Client) Verify(licenseKey string) (map[string]interface{}, error) {
	deviceID := c.GetHardwareID()
	cacheKey := fmt.Sprintf("verify:%s:%s", licenseKey, deviceID)

	if val, ok := c.cache[cacheKey]; ok {
		return val.(map[string]interface{}), nil
	}

	payload := map[string]interface{}{
		"license_key": licenseKey,
		"device_id":   deviceID,
	}

	res, err := c.post("functions/v1/verify-license", payload)
	if err == nil && res["valid"] == true {
		c.cache[cacheKey] = res
	}
	return res, err
}

func (c *Client) Deactivate(licenseKey string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"license_key": licenseKey,
		"device_id":   c.GetHardwareID(),
	}
	res, err := c.post("functions/v1/deactivate-license", payload)
	if err == nil {
		c.cache = make(map[string]interface{}) // Clear cache
	}
	return res, err
}

func (c *Client) post(path string, payload interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.BaseURL, "/"), path)
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: err.Error(), Code: ErrNetwork}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusOK {
		code := ErrUnknown
		if resp.StatusCode == http.StatusTooManyRequests {
			code = ErrRateLimit
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
			code = ErrInvalid
		}

		msg, _ := result["message"].(string)
		if msg == "" {
			msg, _ = result["error"].(string)
		}
		return nil, &Error{Message: msg, Code: code, Status: resp.StatusCode}
	}

	return result, nil
}
