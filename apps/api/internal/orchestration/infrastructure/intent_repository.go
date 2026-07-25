package infrastructure

import (
	"context"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IntentRepository struct {
	pool *pgxpool.Pool
}

func NewIntentRepository(pool *pgxpool.Pool) *IntentRepository {
	return &IntentRepository{pool: pool}
}

func (r *IntentRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *IntentRepository) Save(ctx context.Context, intent *domain.Intent) error {
	payload, err := marshalJSONMap(intent.Payload())
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).InsertIntent(ctx, sqlc.InsertIntentParams{
		ID:             intent.ID(),
		OrganizationID: intent.OrganizationID(),
		WorkspaceID:    intent.WorkspaceID(),
		ConnectionID:   intent.ConnectionID(),
		Capability:     intent.Capability(),
		Payload:        payload,
		Status:         string(intent.Status()),
		CorrelationID:  strPtr(intent.CorrelationID()),
		IdempotencyKey: strPtr(intent.IdempotencyKey()),
		SubmittedAt:    timestamptz(intent.SubmittedAt()),
		CreatedAt:      timestamptz(intent.CreatedAt()),
	}))
}

func (r *IntentRepository) Update(ctx context.Context, intent *domain.Intent) error {
	return mapUnique(r.q(ctx).UpdateIntent(ctx, sqlc.UpdateIntentParams{
		ID:     intent.ID(),
		Status: string(intent.Status()),
	}))
}

func (r *IntentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Intent, error) {
	row, err := r.q(ctx).GetIntentByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapIntent(row)
}

func mapIntent(row sqlc.Intent) (*domain.Intent, error) {
	payload, err := unmarshalJSONMap(row.Payload)
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteIntent(
		row.ID,
		row.OrganizationID,
		row.WorkspaceID,
		row.ConnectionID,
		row.Capability,
		payload,
		domain.IntentStatus(row.Status),
		strFromPtr(row.CorrelationID),
		strFromPtr(row.IdempotencyKey),
		timeFrom(row.SubmittedAt),
		timeFrom(row.CreatedAt),
	), nil
}
