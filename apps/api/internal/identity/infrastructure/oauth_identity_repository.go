package infrastructure

import (
	"context"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthIdentityRepository struct {
	pool *pgxpool.Pool
}

func NewOAuthIdentityRepository(pool *pgxpool.Pool) *OAuthIdentityRepository {
	return &OAuthIdentityRepository{pool: pool}
}

func (r *OAuthIdentityRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *OAuthIdentityRepository) Save(ctx context.Context, identity *domain.OAuthIdentity) error {
	return mapUnique(r.q(ctx).InsertOAuthIdentity(ctx, sqlc.InsertOAuthIdentityParams{
		ID:              identity.ID(),
		UserID:          identity.UserID(),
		Provider:        string(identity.Provider()),
		ProviderSubject: identity.ProviderSubject(),
		Email:           identity.Email(),
		LinkedAt:        timestamptz(identity.LinkedAt()),
		LastLoginAt:     timestamptzPtr(identity.LastLoginAt()),
		CreatedAt:       timestamptz(identity.CreatedAt()),
		UpdatedAt:       timestamptz(identity.UpdatedAt()),
	}))
}

func (r *OAuthIdentityRepository) Update(ctx context.Context, identity *domain.OAuthIdentity) error {
	return mapUnique(r.q(ctx).UpdateOAuthIdentity(ctx, sqlc.UpdateOAuthIdentityParams{
		ID:          identity.ID(),
		Email:       identity.Email(),
		LastLoginAt: timestamptzPtr(identity.LastLoginAt()),
		UpdatedAt:   timestamptz(identity.UpdatedAt()),
	}))
}

func (r *OAuthIdentityRepository) FindByProviderSubject(
	ctx context.Context,
	provider domain.OAuthProvider,
	subject string,
) (*domain.OAuthIdentity, error) {
	row, err := r.q(ctx).GetOAuthIdentityByProviderSubject(ctx, sqlc.GetOAuthIdentityByProviderSubjectParams{
		Provider:        string(provider),
		ProviderSubject: subject,
	})
	if err != nil {
		return nil, mapNotFound(err)
	}
	return domain.ReconstituteOAuthIdentity(
		row.ID,
		row.UserID,
		domain.OAuthProvider(row.Provider),
		row.ProviderSubject,
		row.Email,
		timeFrom(row.LinkedAt),
		timePtrFrom(row.LastLoginAt),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
	), nil
}
