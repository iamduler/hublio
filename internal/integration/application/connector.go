package application

import (
	"context"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

type RegisterCapabilityInput struct {
	Code        string
	DisplayName string
	IsAsync     bool
}

type RegisterConnectorInput struct {
	Code             string
	Name             string
	Vendor           string
	Category         domain.ConnectorCategory
	Version          string
	Description      string
	Homepage         string
	DocumentationURL string
	Capabilities     []RegisterCapabilityInput
}

// RegisterConnector installs a new Connector in Registered status.
// This is a platform-admin / seed operation for Phase C (no real provider yet).
func (s *Services) RegisterConnector(ctx context.Context, in RegisterConnectorInput) (*domain.Connector, error) {
	now := s.clock().Now()

	connectorID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate connector id", apperr.ErrCodeInternal)
	}

	connector, err := domain.NewConnector(connectorID, in.Code, in.Name, in.Vendor, in.Category, in.Version, in.Description, in.Homepage, in.DocumentationURL, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	for _, capIn := range in.Capabilities {
		capID, err := id.NewV7()
		if err != nil {
			return nil, apperr.Wrap(err, "failed to generate capability id", apperr.ErrCodeInternal)
		}
		if _, err := connector.AddCapability(capID, capIn.Code, capIn.DisplayName, capIn.IsAsync, now); err != nil {
			return nil, mapDomainErr(err)
		}
	}

	if err := s.Connectors.Save(ctx, connector); err != nil {
		return nil, mapRepoErr(err)
	}
	return connector, nil
}

func (s *Services) EnableConnector(ctx context.Context, connectorID uuid.UUID) (*domain.Connector, error) {
	return s.changeConnector(ctx, connectorID, func(c *domain.Connector) error {
		return c.Enable(s.clock().Now())
	})
}

func (s *Services) DisableConnector(ctx context.Context, connectorID uuid.UUID) (*domain.Connector, error) {
	return s.changeConnector(ctx, connectorID, func(c *domain.Connector) error {
		return c.Disable(s.clock().Now())
	})
}

func (s *Services) RemoveConnector(ctx context.Context, connectorID uuid.UUID) (*domain.Connector, error) {
	return s.changeConnector(ctx, connectorID, func(c *domain.Connector) error {
		return c.Remove(s.clock().Now())
	})
}

func (s *Services) changeConnector(ctx context.Context, connectorID uuid.UUID, fn func(*domain.Connector) error) (*domain.Connector, error) {
	connector, err := s.Connectors.FindByID(ctx, connectorID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := fn(connector); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connectors.Update(ctx, connector); err != nil {
		return nil, mapRepoErr(err)
	}
	return connector, nil
}

func (s *Services) GetConnector(ctx context.Context, connectorID uuid.UUID) (*domain.Connector, error) {
	connector, err := s.Connectors.FindByID(ctx, connectorID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return connector, nil
}

func (s *Services) ListConnectors(ctx context.Context) ([]*domain.Connector, error) {
	list, err := s.Connectors.List(ctx)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

// SeedFakeConnector registers the built-in "fake" Connector (idempotent) used by
// Orchestration tests before any real provider Connector exists.
func (s *Services) SeedFakeConnector(ctx context.Context) (*domain.Connector, error) {
	existing, err := s.Connectors.FindByCode(ctx, "fake")
	if err == nil {
		return existing, nil
	}
	if err != domain.ErrNotFound {
		return nil, mapRepoErr(err)
	}

	connector, err := s.RegisterConnector(ctx, RegisterConnectorInput{
		Code:        "fake",
		Name:        "Fake Connector",
		Vendor:      "hublio",
		Category:    domain.ConnectorCategoryUtility,
		Version:     "1.0.0",
		Description: "Noop connector for local development and Orchestration tests.",
		Capabilities: []RegisterCapabilityInput{
			{Code: "echo", DisplayName: "Echo", IsAsync: false},
		},
	})
	if err != nil {
		return nil, err
	}
	if err := connector.Enable(s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Connectors.Update(ctx, connector); err != nil {
		return nil, mapRepoErr(err)
	}
	return connector, nil
}
