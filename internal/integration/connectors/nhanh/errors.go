package nhanh

import (
	"errors"
	"fmt"
)

var (
	ErrMissingCredentials = errors.New("nhanh: missing credentials (access_token) or config (app_id, business_id)")
	ErrAuthFailed         = errors.New("nhanh: authentication failed")
	ErrUnsupportedCap     = errors.New("nhanh: unsupported capability")
	ErrProviderRejected   = errors.New("nhanh: provider rejected request")
	ErrInvalidPayload     = errors.New("nhanh: invalid canonical payload")
	ErrNotFound           = errors.New("nhanh: resource not found")
)

func providerError(code string) error {
	if code == "" {
		return ErrProviderRejected
	}
	return fmt.Errorf("%w: %s", ErrProviderRejected, code)
}
