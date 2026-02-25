package licenseflow

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseURL   string
	APIKey    string
	JWTSecret string
	Timeout   time.Duration
	CacheTTL  time.Duration // TTL for cached verifications (default 5 min)
}

type cacheEntry struct {
	data      map[string]interface{}
	expiresAt time.Time
}

type Client struct {
	config     Config
	httpClient *http.Client
	cache      map[string]cacheEntry
	cacheMu    sync.RWMutex
}

func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache: make(map[string]cacheEntry),
	}
}

func (c *Client) cacheGet(key string) (map[string]interface{}, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	entry, ok := c.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *Client) cacheSet(key string, data map[string]interface{}) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache[key] = cacheEntry{data: data, expiresAt: time.Now().Add(c.config.CacheTTL)}
}

func (c *Client) cacheClear() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]cacheEntry)
}

func (c *Client) GetHardwareID() string {
	hostname, _ := os.Hostname()
	return hostname
}

func (c *Client) Activate(licenseKey string, deviceName string, environmentID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"license_key": licenseKey,
		"device_id":   c.GetHardwareID(),
		"device_name": deviceName,
	}
	if environmentID != "" {
		payload["environment_id"] = environmentID
	}
	return c.post("functions/v1/activate-license", payload)
}

func (c *Client) Verify(licenseKey string, environmentID string) (map[string]interface{}, error) {
	deviceID := c.GetHardwareID()
	envID := environmentID
	if envID == "" {
		envID = "default"
	}
	cacheKey := fmt.Sprintf("verify:%s:%s:%s", licenseKey, deviceID, envID)

	if val, ok := c.cacheGet(cacheKey); ok {
		return val, nil
	}

	payload := map[string]interface{}{
		"licenseKey": licenseKey,
		"deviceId":   deviceID,
	}
	if environmentID != "" {
		payload["environmentId"] = environmentID
	}

	res, err := c.post("functions/v1/verify-license", payload)
	if err == nil && res["valid"] == true {
		c.cacheSet(cacheKey, res)
	}
	return res, err
}

func (c *Client) Deactivate(licenseKey string, environmentID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"license_key": licenseKey,
		"device_id":   c.GetHardwareID(),
	}
	if environmentID != "" {
		payload["environment_id"] = environmentID
	}
	res, err := c.post("functions/v1/deactivate-license", payload)
	if err == nil {
		c.cacheClear()
	}
	return res, err
}

// CheckoutLicense acquires a temporary floating license lease
func (c *Client) CheckoutLicense(licenseKey string, durationSeconds int, requesterID string, requesterType string) (map[string]interface{}, error) {
	if requesterID == "" {
		requesterID = c.GetHardwareID()
	}
	if requesterType == "" {
		requesterType = "ci"
	}
	payload := map[string]interface{}{
		"license_key":      licenseKey,
		"duration_seconds": durationSeconds,
		"requester_id":     requesterID,
		"requester_type":   requesterType,
	}
	return c.post("functions/v1/checkout-license", payload)
}

// CheckinLicense releases a floating license lease early
func (c *Client) CheckinLicense(leaseKey string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"lease_key": leaseKey,
	}
	return c.post("functions/v1/checkin-license", payload)
}

// GetLeaseStatus checks the status and remaining time of a lease
func (c *Client) GetLeaseStatus(leaseKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/functions/v1/lease-status?lease_key=%s", strings.TrimSuffix(c.config.BaseURL, "/"), leaseKey)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: err.Error(), Code: ErrNetwork}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// CheckForUpdates checks for the latest release for a product
func (c *Client) CheckForUpdates(productID, currentVersion, channel string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/functions/v1/release-management/latest?product_id=%s&channel=%s", strings.TrimSuffix(c.config.BaseURL, "/"), productID, channel)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: err.Error(), Code: ErrNetwork}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil
	}

	if ver, ok := result["version"].(string); ok && ver == currentVersion {
		return nil, nil
	}

	return result, nil
}

// DownloadArtifact retrieves the download URL for a release artifact
func (c *Client) DownloadArtifact(licenseKey, releaseID, platform, arch string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"license_key":  licenseKey,
		"release_id":   releaseID,
		"platform":     platform,
		"architecture": arch,
	}
	return c.post("functions/v1/artifact-download", payload)
}

// HasFeature checks if an entitlement is enabled
func (c *Client) HasFeature(verification map[string]interface{}, featureCode string) bool {
	entitlements, ok := verification["entitlements"].(map[string]interface{})
	if !ok {
		return false
	}
	val, exists := entitlements[featureCode]
	if !exists {
		return false
	}
	if boolVal, ok := val.(bool); ok {
		return boolVal
	}
	if mapVal, ok := val.(map[string]interface{}); ok {
		if enabled, ok := mapVal["enabled"].(bool); ok {
			return enabled
		}
		if value, ok := mapVal["value"].(bool); ok {
			return value
		}
	}
	if strVal, ok := val.(string); ok {
		return strVal == "true"
	}
	return false
}

// GetEntitlement returns the raw value of an entitlement
func (c *Client) GetEntitlement(verification map[string]interface{}, featureCode string) interface{} {
	entitlements, ok := verification["entitlements"].(map[string]interface{})
	if !ok {
		return nil
	}
	return entitlements[featureCode]
}

// VerifyOfflineLicense validates a local .lic file
func (c *Client) VerifyOfflineLicense(licenseContent string, publicKeyHex string) (map[string]interface{}, error) {
	var licenseData struct {
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal([]byte(licenseContent), &licenseData); err != nil {
		return nil, fmt.Errorf("invalid license format")
	}

	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid public key hex")
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(licenseData.Signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature base64")
	}

	valid := ed25519.Verify(pubKeyBytes, []byte(licenseData.Payload), sigBytes)
	if !valid {
		return nil, fmt.Errorf("invalid signature")
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(licenseData.Payload)
	if err != nil {
		return nil, fmt.Errorf("invalid payload base64")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload json")
	}
	return payload, nil
}

func (c *Client) post(path string, payload interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.BaseURL, "/"), path)
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: err.Error(), Code: ErrNetwork}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode >= 400 {
		code := ErrUnknown
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			code = ErrRateLimit
		case http.StatusBadRequest, http.StatusNotFound:
			code = ErrInvalid
		}

		msg := "Unknown error"
		if result != nil {
			if m, ok := result["message"].(string); ok && m != "" {
				msg = m
			} else if e, ok := result["error"].(string); ok && e != "" {
				msg = e
			}
		}
		if msg == "Unknown error" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, &Error{Message: msg, Code: code, Status: resp.StatusCode}
	}

	// Handle case where body might be empty but status ok
	if result == nil {
		result = make(map[string]interface{})
	}

	return result, nil
}
