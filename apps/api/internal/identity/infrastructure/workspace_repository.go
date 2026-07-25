package infrastructure

import (
	"context"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceRepository struct {
	pool *pgxpool.Pool
}

func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

func (r *WorkspaceRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *WorkspaceRepository) Save(ctx context.Context, ws *domain.Workspace) error {
	return mapUnique(r.q(ctx).InsertWorkspace(ctx, sqlc.InsertWorkspaceParams{
		ID:             ws.ID(),
		OrganizationID: ws.OrganizationID(),
		Name:           ws.Name(),
		Environment:    ws.Environment(),
		Status:         string(ws.Status()),
		CreatedAt:      timestamptz(ws.CreatedAt()),
		UpdatedAt:      timestamptz(ws.UpdatedAt()),
		DeletedAt:      timestamptzPtr(ws.DeletedAt()),
	}))
}

func (r *WorkspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	return mapUnique(r.q(ctx).UpdateWorkspace(ctx, sqlc.UpdateWorkspaceParams{
		ID:          ws.ID(),
		Name:        ws.Name(),
		Environment: ws.Environment(),
		Status:      string(ws.Status()),
		UpdatedAt:   timestamptz(ws.UpdatedAt()),
		DeletedAt:   timestamptzPtr(ws.DeletedAt()),
	}))
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	row, err := r.q(ctx).GetWorkspaceByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapWorkspace(row), nil
}

func (r *WorkspaceRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.Workspace, error) {
	rows, err := r.q(ctx).ListWorkspacesByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Workspace, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapWorkspace(row))
	}
	return out, nil
}

func mapWorkspace(row sqlc.Workspace) *domain.Workspace {
	return domain.ReconstituteWorkspace(
		row.ID,
		row.OrganizationID,
		row.Name,
		row.Environment,
		domain.WorkspaceStatus(row.Status),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.DeletedAt),
	)
}
