package errors

import (
	"fmt"
)

// 業務錯誤類型
type BotError struct {
	Code    string
	Message string
	Err     error
}

func (e *BotError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *BotError) Unwrap() error {
	return e.Err
}

// 預定義錯誤代碼
const (
	ErrCodeAPICall           = "API_CALL"
	ErrCodeRateLimit         = "RATE_LIMIT"
	ErrCodeConfig            = "CONFIG"
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeOrderFailed       = "ORDER_FAILED"
	ErrCodeAuthentication    = "AUTH_FAILED"
)

// 創建錯誤的便利函數
func NewAPIError(message string, err error) *BotError {
	return &BotError{Code: ErrCodeAPICall, Message: message, Err: err}
}

func NewConfigError(message string, err error) *BotError {
	return &BotError{Code: ErrCodeConfig, Message: message, Err: err}
}

func NewValidationError(message string) *BotError {
	return &BotError{Code: ErrCodeInvalidInput, Message: message}
}

func NewOrderError(message string, err error) *BotError {
	return &BotError{Code: ErrCodeOrderFailed, Message: message, Err: err}
}
