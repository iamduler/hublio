package nhanh

import (
	"fmt"

	"hublio/internal/integration/domain"
)

var (
	ErrMissingCredentials = domain.ErrRuntimeMissingCredentials
	ErrAuthFailed         = domain.ErrRuntimeAuthFailed
	ErrUnsupportedCap     = domain.ErrRuntimeUnsupportedCapability
	ErrProviderRejected   = domain.ErrRuntimeProviderRejected
	ErrInvalidPayload     = domain.ErrRuntimeInvalidPayload
	ErrNotFound           = domain.ErrRuntimeNotFound
)

func providerError(code string) error {
	if code == "" {
		return ErrProviderRejected
	}
	return fmt.Errorf("%w: %s", ErrProviderRejected, code)
}
