package infrastructure

import (
	"context"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectionRepository struct {
	pool *pgxpool.Pool
}

func NewConnectionRepository(pool *pgxpool.Pool) *ConnectionRepository {
	return &ConnectionRepository{pool: pool}
}

func (r *ConnectionRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *ConnectionRepository) Save(ctx context.Context, conn *domain.Connection) error {
	config, err := marshalJSONMap(conn.Config())
	if err != nil {
		return err
	}
	retryPolicy, err := marshalJSONMap(conn.RetryPolicy())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).InsertConnection(ctx, sqlc.InsertConnectionParams{
		ID:                 conn.ID(),
		ActiveCredentialID: uuidPtrToPgtype(conn.ActiveCredentialID()),
		WorkspaceID:        conn.WorkspaceID(),
		ConnectorID:        conn.ConnectorID(),
		Name:               conn.Name(),
		IsDefault:          conn.IsDefault(),
		Description:        strPtr(conn.Description()),
		Environment:        conn.Environment(),
		Status:             string(conn.Status()),
		Config:             config,
		RetryPolicy:        retryPolicy,
		TimeoutSeconds:     int32(conn.TimeoutSeconds()),
		CreatedAt:          timestamptz(conn.CreatedAt()),
		UpdatedAt:          timestamptz(conn.UpdatedAt()),
		DeletedAt:          timestamptzPtr(conn.DeletedAt()),
	}))
}

func (r *ConnectionRepository) Update(ctx context.Context, conn *domain.Connection) error {
	config, err := marshalJSONMap(conn.Config())
	if err != nil {
		return err
	}
	retryPolicy, err := marshalJSONMap(conn.RetryPolicy())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).UpdateConnection(ctx, sqlc.UpdateConnectionParams{
		ID:                 conn.ID(),
		ActiveCredentialID: uuidPtrToPgtype(conn.ActiveCredentialID()),
		Name:               conn.Name(),
		Description:        strPtr(conn.Description()),
		Status:             string(conn.Status()),
		Config:             config,
		RetryPolicy:        retryPolicy,
		TimeoutSeconds:     int32(conn.TimeoutSeconds()),
		UpdatedAt:          timestamptz(conn.UpdatedAt()),
		DeletedAt:          timestamptzPtr(conn.DeletedAt()),
	}))
}

func (r *ConnectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Connection, error) {
	row, err := r.q(ctx).GetConnectionByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapConnection(row)
}

func (r *ConnectionRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.Connection, error) {
	rows, err := r.q(ctx).ListConnectionsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Connection, 0, len(rows))
	for _, row := range rows {
		conn, err := mapConnection(row)
		if err != nil {
			return nil, err
		}
		out = append(out, conn)
	}
	return out, nil
}

func mapConnection(row sqlc.Connection) (*domain.Connection, error) {
	config, err := unmarshalJSONMap(row.Config)
	if err != nil {
		return nil, err
	}
	retryPolicy, err := unmarshalJSONMap(row.RetryPolicy)
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteConnection(
		row.ID,
		row.WorkspaceID,
		row.ConnectorID,
		row.Name,
		row.IsDefault,
		strFromPtr(row.Description),
		row.Environment,
		domain.ConnectionStatus(row.Status),
		config,
		retryPolicy,
		int(row.TimeoutSeconds),
		pgtypeToUUIDPtr(row.ActiveCredentialID),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.DeletedAt),
	), nil
}
