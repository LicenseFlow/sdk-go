package licenseflow

import "fmt"

type Error struct {
	Message string
	Code    string
	Status  int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s (Status: %d)", e.Code, e.Message, e.Status)
}

var (
	ErrNetwork      = "NETWORK_ERROR"
	ErrRateLimit    = "RATE_LIMIT_EXCEEDED"
	ErrInvalid      = "INVALID_LICENSE"
	ErrUnknown      = "UNKNOWN_ERROR"
)
