package domain

import "errors"

var (
	ErrInvalidCode           = errors.New("integration: invalid code")
	ErrInvalidName           = errors.New("integration: invalid name")
	ErrInvalidVendor         = errors.New("integration: invalid vendor")
	ErrInvalidCategory       = errors.New("integration: invalid category")
	ErrInvalidVersion        = errors.New("integration: invalid version")
	ErrInvalidEnvironment    = errors.New("integration: invalid environment")
	ErrInvalidCredentialType = errors.New("integration: invalid credential type")
	ErrInvalidTransition     = errors.New("integration: invalid status transition")
	ErrConnectorRemoved      = errors.New("integration: connector is removed")
	ErrConnectorNotUsable    = errors.New("integration: connector is not enabled")
	ErrCapabilityNotFound    = errors.New("integration: capability not found")
	ErrCredentialNotActive   = errors.New("integration: credential is not active")
	ErrConnectionNotActive   = errors.New("integration: connection is not active")
	ErrEmptySecret           = errors.New("integration: encrypted secret is required")
	ErrNotFound              = errors.New("integration: not found")
	ErrConflict              = errors.New("integration: conflict")
)
