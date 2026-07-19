package infrastructure

import (
	"context"
	"errors"

	integrationdomain "hublio/internal/integration/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"
)

// ConnectorGateway adapts the Integration Connector Runtime registry into the Orchestration
// Application's ConnectorGateway port. Provider DTOs never cross this boundary: both sides
// only exchange canonical-ish map[string]any payloads.
type ConnectorGateway struct {
	registry integrationdomain.RuntimeRegistry
}

func NewConnectorGateway(registry integrationdomain.RuntimeRegistry) *ConnectorGateway {
	return &ConnectorGateway{registry: registry}
}

func (g *ConnectorGateway) Invoke(ctx context.Context, connectorCode string, in orchestrationapp.InvokeRequest) (orchestrationapp.InvokeResponse, error) {
	runtime, err := g.registry.Resolve(connectorCode)
	if err != nil {
		return orchestrationapp.InvokeResponse{}, apperr.Wrap(err, "connector runtime not available", apperr.ErrCodeServiceUnavailable)
	}

	out, err := runtime.Invoke(ctx, integrationdomain.InvokeInput{
		ConnectionID: in.ConnectionID,
		Capability:   in.Capability,
		Config:       in.Config,
		Secret:       in.Secret,
		Payload:      in.Payload,
	})
	if err != nil {
		return orchestrationapp.InvokeResponse{}, mapInvokeErr(err)
	}
	return orchestrationapp.InvokeResponse{Payload: out.Payload, Metadata: out.Metadata}, nil
}

// mapInvokeErr translates Domain Runtime sentinels (wrapped by vendor connectors) into
// platform AppError codes so auth/payload failures are not all BadGateway.
func mapInvokeErr(err error) error {
	switch {
	case errors.Is(err, integrationdomain.ErrRuntimeAuthFailed),
		errors.Is(err, integrationdomain.ErrRuntimeMissingCredentials):
		return apperr.Wrap(err, "connector authentication failed", apperr.ErrCodeUnauthorized)
	case errors.Is(err, integrationdomain.ErrRuntimeInvalidPayload),
		errors.Is(err, integrationdomain.ErrRuntimeMissingConfig),
		errors.Is(err, integrationdomain.ErrRuntimeUnsupportedCapability):
		return apperr.Wrap(err, "connector rejected request", apperr.ErrCodeBadRequest)
	case errors.Is(err, integrationdomain.ErrRuntimeNotFound):
		return apperr.Wrap(err, "connector resource not found", apperr.ErrCodeNotFound)
	case errors.Is(err, integrationdomain.ErrRuntimeProviderRejected):
		return apperr.Wrap(err, "connector provider rejected request", apperr.ErrCodeBadGateway)
	default:
		return apperr.Wrap(err, "connector invoke failed", apperr.ErrCodeBadGateway)
	}
}
