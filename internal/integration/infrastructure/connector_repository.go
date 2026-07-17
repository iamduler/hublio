package infrastructure

import (
	"context"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectorRepository struct {
	pool *pgxpool.Pool
}

func NewConnectorRepository(pool *pgxpool.Pool) *ConnectorRepository {
	return &ConnectorRepository{pool: pool}
}

func (r *ConnectorRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *ConnectorRepository) Save(ctx context.Context, connector *domain.Connector) error {
	q := r.q(ctx)
	if err := mapUnique(q.InsertConnector(ctx, sqlc.InsertConnectorParams{
		ID:               connector.ID(),
		Code:             connector.Code(),
		Name:             connector.Name(),
		Vendor:           connector.Vendor(),
		Category:         string(connector.Category()),
		Version:          connector.Version(),
		Status:           string(connector.Status()),
		Description:      strPtr(connector.Description()),
		Homepage:         strPtr(connector.Homepage()),
		DocumentationUrl: strPtr(connector.DocumentationURL()),
		CreatedAt:        timestamptz(connector.CreatedAt()),
		UpdatedAt:        timestamptz(connector.UpdatedAt()),
		DeletedAt:        timestamptzPtr(connector.DeletedAt()),
	})); err != nil {
		return err
	}

	for _, capability := range connector.Capabilities() {
		if err := mapUnique(q.InsertConnectorCapability(ctx, sqlc.InsertConnectorCapabilityParams{
			ID:             capability.ID(),
			ConnectorID:    capability.ConnectorID(),
			CapabilityCode: capability.Code(),
			DisplayName:    capability.DisplayName(),
			Status:         string(capability.Status()),
			IsAsync:        capability.IsAsync(),
			CreatedAt:      timestamptz(capability.CreatedAt()),
			UpdatedAt:      timestamptz(capability.UpdatedAt()),
		})); err != nil {
			return err
		}
	}
	return nil
}

func (r *ConnectorRepository) Update(ctx context.Context, connector *domain.Connector) error {
	return mapUnique(r.q(ctx).UpdateConnector(ctx, sqlc.UpdateConnectorParams{
		ID:        connector.ID(),
		Status:    string(connector.Status()),
		UpdatedAt: timestamptz(connector.UpdatedAt()),
		DeletedAt: timestamptzPtr(connector.DeletedAt()),
	}))
}

func (r *ConnectorRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Connector, error) {
	row, err := r.q(ctx).GetConnectorByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return r.hydrate(ctx, row)
}

func (r *ConnectorRepository) FindByCode(ctx context.Context, code string) (*domain.Connector, error) {
	row, err := r.q(ctx).GetConnectorByCode(ctx, code)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return r.hydrate(ctx, row)
}

func (r *ConnectorRepository) List(ctx context.Context) ([]*domain.Connector, error) {
	rows, err := r.q(ctx).ListConnectors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Connector, 0, len(rows))
	for _, row := range rows {
		connector, err := r.hydrate(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, connector)
	}
	return out, nil
}

func (r *ConnectorRepository) hydrate(ctx context.Context, row sqlc.Connector) (*domain.Connector, error) {
	capRows, err := r.q(ctx).ListCapabilitiesByConnector(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	capabilities := make([]*domain.Capability, 0, len(capRows))
	for _, capRow := range capRows {
		capabilities = append(capabilities, domain.ReconstituteCapability(
			capRow.ID,
			capRow.ConnectorID,
			capRow.CapabilityCode,
			capRow.DisplayName,
			domain.CapabilityStatus(capRow.Status),
			capRow.IsAsync,
			timeFrom(capRow.CreatedAt),
			timeFrom(capRow.UpdatedAt),
		))
	}
	return domain.ReconstituteConnector(
		row.ID,
		row.Code,
		row.Name,
		row.Vendor,
		domain.ConnectorCategory(row.Category),
		row.Version,
		domain.ConnectorStatus(row.Status),
		strFromPtr(row.Description),
		strFromPtr(row.Homepage),
		strFromPtr(row.DocumentationUrl),
		capabilities,
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.DeletedAt),
	), nil
}
