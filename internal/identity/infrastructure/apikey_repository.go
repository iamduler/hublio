package infrastructure

import (
	"context"
	"time"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

func (r *APIKeyRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *APIKeyRepository) Save(ctx context.Context, key *domain.APIKey) error {
	return mapUnique(r.q(ctx).InsertAPIKey(ctx, sqlc.InsertAPIKeyParams{
		ID:                key.ID(),
		WorkspaceID:       key.WorkspaceID(),
		Name:              key.Name(),
		KeyHash:           key.KeyHash(),
		LastUsedAt:        timestamptzPtr(nil),
		ExpiresAt:         timestamptzPtr(key.ExpiresAt()),
		Status:            string(key.Status()),
		Prefix:            key.Prefix(),
		LastUsedIp:        nil,
		LastUsedUserAgent: nil,
		CreatedAt:         timestamptz(key.CreatedAt()),
		UpdatedAt:         timestamptz(key.UpdatedAt()),
		DeletedAt:         timestamptzPtr(key.DeletedAt()),
	}))
}

func (r *APIKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	return mapUnique(r.q(ctx).UpdateAPIKey(ctx, sqlc.UpdateAPIKeyParams{
		ID:        key.ID(),
		Name:      key.Name(),
		KeyHash:   key.KeyHash(),
		ExpiresAt: timestamptzPtr(key.ExpiresAt()),
		Status:    string(key.Status()),
		Prefix:    key.Prefix(),
		UpdatedAt: timestamptz(key.UpdatedAt()),
		DeletedAt: timestamptzPtr(key.DeletedAt()),
	}))
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	row, err := r.q(ctx).GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapAPIKey(row), nil
}

func (r *APIKeyRepository) FindByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error) {
	row, err := r.q(ctx).GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapAPIKey(row), nil
}

func (r *APIKeyRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.APIKey, error) {
	rows, err := r.q(ctx).ListAPIKeysByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.APIKey, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAPIKey(row))
	}
	return out, nil
}

func (r *APIKeyRepository) TouchLastUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.q(ctx).TouchAPIKeyLastUsed(ctx, sqlc.TouchAPIKeyLastUsedParams{
		ID:         id,
		LastUsedAt: timestamptz(at),
	})
}

func mapAPIKey(row sqlc.ApiKey) *domain.APIKey {
	return domain.ReconstituteAPIKey(
		row.ID,
		row.WorkspaceID,
		row.Name,
		row.KeyHash,
		row.Prefix,
		domain.APIKeyStatus(row.Status),
		timePtrFrom(row.ExpiresAt),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		timePtrFrom(row.DeletedAt),
	)
}
