package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MFARepository struct {
	pool *pgxpool.Pool
}

func NewMFARepository(pool *pgxpool.Pool) *MFARepository {
	return &MFARepository{pool: pool}
}

func (r *MFARepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *MFARepository) Save(ctx context.Context, cfg *domain.MFAConfig) error {
	hashes, err := marshalRecoveryCodeHashes(cfg.RecoveryCodeHashes())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).UpsertUserMFA(ctx, sqlc.UpsertUserMFAParams{
		UserID:              cfg.UserID(),
		TotpSecretEncrypted: cfg.TOTPSecretEncrypted(),
		EnabledAt:           timestamptzPtr(cfg.EnabledAt()),
		RecoveryCodesHash:   hashes,
		CreatedAt:           timestamptz(cfg.CreatedAt()),
		UpdatedAt:           timestamptz(cfg.UpdatedAt()),
	}))
}

func (r *MFARepository) Update(ctx context.Context, cfg *domain.MFAConfig) error {
	hashes, err := marshalRecoveryCodeHashes(cfg.RecoveryCodeHashes())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).UpdateUserMFA(ctx, sqlc.UpdateUserMFAParams{
		UserID:              cfg.UserID(),
		TotpSecretEncrypted: cfg.TOTPSecretEncrypted(),
		EnabledAt:           timestamptzPtr(cfg.EnabledAt()),
		RecoveryCodesHash:   hashes,
		UpdatedAt:           timestamptz(cfg.UpdatedAt()),
	}))
}

func (r *MFARepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.MFAConfig, error) {
	row, err := r.q(ctx).GetUserMFA(ctx, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	hashes, err := unmarshalRecoveryCodeHashes(row.RecoveryCodesHash)
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteMFAConfig(
		row.UserID,
		row.TotpSecretEncrypted,
		timePtrFrom(row.EnabledAt),
		hashes,
		timeFrom(row.CreatedAt),
		timeFrom(row.UpdatedAt),
	), nil
}

func (r *MFARepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return mapUnique(r.q(ctx).DeleteUserMFA(ctx, userID))
}

func marshalRecoveryCodeHashes(hashes []string) ([]byte, error) {
	raw, err := json.Marshal(hashes)
	if err != nil {
		return nil, fmt.Errorf("identity: encode recovery code hashes: %w", err)
	}
	return raw, nil
}

func unmarshalRecoveryCodeHashes(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var hashes []string
	if err := json.Unmarshal(raw, &hashes); err != nil {
		return nil, fmt.Errorf("identity: decode recovery code hashes: %w", err)
	}
	return hashes, nil
}
