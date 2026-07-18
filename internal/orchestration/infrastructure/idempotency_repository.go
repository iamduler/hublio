package infrastructure

import (
	"context"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdempotencyRepository persists Intent Idempotency-Key records. Postgres is the source of
// truth (docs/20-database-schema.dbml); Redis is never used to decide idempotency outcomes.
type IdempotencyRepository struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{pool: pool}
}

func (r *IdempotencyRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *IdempotencyRepository) Save(ctx context.Context, rec *domain.IdempotencyKey) error {
	var intentID uuid.UUID
	if rec.IntentID() != nil {
		intentID = *rec.IntentID()
	}
	return mapUnique(r.q(ctx).InsertIdempotencyKey(ctx, sqlc.InsertIdempotencyKeyParams{
		ID:             rec.ID(),
		OrganizationID: rec.OrganizationID(),
		WorkspaceID:    rec.WorkspaceID(),
		IdempotencyKey: rec.Key(),
		IntentID:       uuidPtrToPgtype(&intentID),
		ExpiresAt:      timestamptzPtr(rec.ExpiresAt()),
		CreatedAt:      timestamptz(rec.CreatedAt()),
	}))
}

func (r *IdempotencyRepository) FindByKey(ctx context.Context, organizationID, workspaceID uuid.UUID, key string) (*domain.IdempotencyKey, error) {
	row, err := r.q(ctx).GetIdempotencyKeyByOrgWorkspaceKey(ctx, sqlc.GetIdempotencyKeyByOrgWorkspaceKeyParams{
		OrganizationID: organizationID,
		WorkspaceID:    workspaceID,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, mapNotFound(err)
	}
	return domain.ReconstituteIdempotencyKey(
		row.ID,
		row.OrganizationID,
		row.WorkspaceID,
		row.IdempotencyKey,
		pgtypeToUUIDPtr(row.IntentID),
		timePtrFrom(row.ExpiresAt),
		timeFrom(row.CreatedAt),
	), nil
}
