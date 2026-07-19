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
	return s.seedConnector(ctx, RegisterConnectorInput{
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
}

// SeedMISAConnector registers the MISA meInvoice destination Connector (idempotent).
func (s *Services) SeedMISAConnector(ctx context.Context) (*domain.Connector, error) {
	return s.seedConnector(ctx, RegisterConnectorInput{
		Code:             "misa",
		Name:             "MISA meInvoice",
		Vendor:           "MISA",
		Category:         domain.ConnectorCategoryDestination,
		Version:          "1.0.0",
		Description:      "Electronic invoice destination via MISA meInvoice Open API.",
		Homepage:         "https://www.misa.vn/",
		DocumentationURL: "https://www.misa.vn/154989/tai-lieu-open-api-tich-hop-hoa-don-dien-tu-misa-meinvoice-dau-ra/",
		Capabilities: []RegisterCapabilityInput{
			{Code: "invoice.create", DisplayName: "Create / publish e-invoice", IsAsync: false},
		},
	})
}

// SeedNhanhConnector registers the Nhanh.vn origin Connector (idempotent).
func (s *Services) SeedNhanhConnector(ctx context.Context) (*domain.Connector, error) {
	return s.seedConnector(ctx, RegisterConnectorInput{
		Code:             "nhanh",
		Name:             "Nhanh.vn",
		Vendor:           "Nhanh.vn",
		Category:         domain.ConnectorCategorySource,
		Version:          "1.0.0",
		Description:      "Nhanh.vn POS origin (retail bills) and reverse status update.",
		Homepage:         "https://nhanh.vn/",
		DocumentationURL: "https://apidocs.nhanh.vn/",
		Capabilities: []RegisterCapabilityInput{
			{Code: "invoice.get", DisplayName: "Get retail bill as Canonical Invoice", IsAsync: false},
			{Code: "invoice.update_status", DisplayName: "Update order status (reverse)", IsAsync: false},
		},
	})
}

// SeedBuiltInConnectors registers Fake + MISA + Nhanh (idempotent). Used on API boot.
func (s *Services) SeedBuiltInConnectors(ctx context.Context) error {
	if _, err := s.SeedFakeConnector(ctx); err != nil {
		return err
	}
	if _, err := s.SeedMISAConnector(ctx); err != nil {
		return err
	}
	if _, err := s.SeedNhanhConnector(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Services) seedConnector(ctx context.Context, in RegisterConnectorInput) (*domain.Connector, error) {
	existing, err := s.Connectors.FindByCode(ctx, in.Code)
	if err == nil {
		return existing, nil
	}
	if err != domain.ErrNotFound {
		return nil, mapRepoErr(err)
	}

	connector, err := s.RegisterConnector(ctx, in)
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
