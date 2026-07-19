package misa

import (
	"errors"
	"fmt"
)

// Sentinel errors for the MISA Connector Runtime. Callers (Integration/Orchestration) map
// these to platform AppErrors; provider ErrorCode strings never leave this package raw in
// success paths — only translated messages.
var (
	ErrMissingCredentials = errors.New("misa: missing credentials (app_id, username, password, tax_code)")
	ErrMissingConfig      = errors.New("misa: missing required config")
	ErrAuthFailed         = errors.New("misa: authentication failed")
	ErrUnsupportedCap     = errors.New("misa: unsupported capability")
	ErrProviderRejected   = errors.New("misa: provider rejected request")
	ErrInvalidPayload     = errors.New("misa: invalid canonical invoice payload")
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
