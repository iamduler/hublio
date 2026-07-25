package apperr

import "fmt"

type ErrorCode string

const (
	ErrCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInternal           ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeTooManyRequests    ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeBadGateway         ErrorCode = "BAD_GATEWAY"
	ErrCodeGatewayTimeout     ErrorCode = "GATEWAY_TIMEOUT"
)

type AppError struct {
	Message string
	Code    ErrorCode
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s, Error: %v", e.Code, e.Message, e.Err)
}

func New(message string, code ErrorCode) error {
	return &AppError{
		Message: message,
		Code:    code,
		Err:     nil,
	}
}

func Wrap(err error, message string, code ErrorCode) error {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}
