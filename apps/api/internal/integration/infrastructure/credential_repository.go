package infrastructure

import (
	"context"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

func (r *CredentialRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *CredentialRepository) Save(ctx context.Context, cred *domain.Credential) error {
	secret, err := marshalEncryptedSecret(cred.EncryptedSecret())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).InsertCredential(ctx, sqlc.InsertCredentialParams{
		ID:              cred.ID(),
		ConnectionID:    cred.ConnectionID(),
		Type:            string(cred.Type()),
		Status:          string(cred.Status()),
		Version:         int32(cred.Version()),
		EncryptedSecret: secret,
		ExpiresAt:       timestamptzPtr(cred.ExpiresAt()),
		RotatedAt:       timestamptzPtr(cred.RotatedAt()),
		CreatedAt:       timestamptz(cred.CreatedAt()),
		UpdatedAt:       timestamptz(cred.UpdatedAt()),
		CreatedBy:       cred.CreatedBy(),
	}))
}

func (r *CredentialRepository) Update(ctx context.Context, cred *domain.Credential) error {
	return mapUnique(r.q(ctx).UpdateCredential(ctx, sqlc.UpdateCredentialParams{
		ID:        cred.ID(),
		Status:    string(cred.Status()),
		RotatedAt: timestamptzPtr(cred.RotatedAt()),
		UpdatedAt: timestamptz(cred.UpdatedAt()),
	}))
}

func (r *CredentialRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Credential, error) {
	row, err := r.q(ctx).GetCredentialByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapCredential(row)
}

func (r *CredentialRepository) FindActiveByConnection(ctx context.Context, connectionID uuid.UUID) (*domain.Credential, error) {
	row, err := r.q(ctx).GetActiveCredentialByConnection(ctx, connectionID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapCredential(row)
}

func (r *CredentialRepository) ListByConnection(ctx context.Context, connectionID uuid.UUID) ([]*domain.Credential, error) {
	rows, err := r.q(ctx).ListCredentialsByConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Credential, 0, len(rows))
	for _, row := range rows {
		cred, err := mapCredential(row)
		if err != nil {
			return nil, err
		}
		out = append(out, cred)
	}
	return out, nil
}

func mapCredential(row sqlc.Credential) (*domain.Credential, error) {
	secret, err := unmarshalEncryptedSecret(row.EncryptedSecret)
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteCredential(
		row.ID,
		row.ConnectionID,
		domain.CredentialType(row.Type),
		domain.CredentialStatus(row.Status),
		int(row.Version),
		secret,
		timePtrFrom(row.ExpiresAt),
		timePtrFrom(row.RotatedAt),
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
		row.CreatedBy,
	), nil
}
