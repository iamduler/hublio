package domain

import "errors"

// Runtime errors are returned by Connector Runtimes (internal/integration/connectors/*).
// Application/Infrastructure map them to platform AppErrors; Domain stays free of HTTP/provider
// packages. Vendors should wrap these with fmt.Errorf("%w: …", ErrRuntime…) so errors.Is works.
var (
	ErrRuntimeAuthFailed            = errors.New("integration: connector authentication failed")
	ErrRuntimeMissingCredentials    = errors.New("integration: connector missing credentials")
	ErrRuntimeMissingConfig         = errors.New("integration: connector missing config")
	ErrRuntimeInvalidPayload        = errors.New("integration: connector invalid payload")
	ErrRuntimeProviderRejected      = errors.New("integration: connector provider rejected request")
	ErrRuntimeUnsupportedCapability = errors.New("integration: connector unsupported capability")
	ErrRuntimeNotFound              = errors.New("integration: connector resource not found")
)
