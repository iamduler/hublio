package infrastructure

import (
	"context"

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
		return orchestrationapp.InvokeResponse{}, apperr.Wrap(err, "connector invoke failed", apperr.ErrCodeBadGateway)
	}
	return orchestrationapp.InvokeResponse{Payload: out.Payload, Metadata: out.Metadata}, nil
}
