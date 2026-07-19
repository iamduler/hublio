package misa

import (
	"fmt"

	"hublio/internal/integration/domain"
)

// Package-local aliases of Domain Runtime errors so connector code stays readable while
// Orchestration/Integration can errors.Is against domain.ErrRuntime*.
var (
	ErrMissingCredentials = domain.ErrRuntimeMissingCredentials
	ErrMissingConfig      = domain.ErrRuntimeMissingConfig
	ErrAuthFailed         = domain.ErrRuntimeAuthFailed
	ErrUnsupportedCap     = domain.ErrRuntimeUnsupportedCapability
	ErrProviderRejected   = domain.ErrRuntimeProviderRejected
	ErrInvalidPayload     = domain.ErrRuntimeInvalidPayload
)

func authError(providerCode string) error {
	if providerCode == "" {
		return ErrAuthFailed
	}
	return fmt.Errorf("%w: %s", ErrAuthFailed, providerCode)
}

func providerError(providerCode, detail string) error {
	if detail == "" {
		return fmt.Errorf("%w: %s", ErrProviderRejected, providerCode)
	}
	return fmt.Errorf("%w: %s (%s)", ErrProviderRejected, providerCode, detail)
}
