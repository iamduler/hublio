package infrastructure

import (
	"context"
	"fmt"

	"hublio/internal/events/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository is the Postgres implementation of domain.AuditRepository. It only ever
// inserts: `audit_logs` is append-only and immutable.
type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *AuditRepository) Save(ctx context.Context, entry *domain.AuditEntry) error {
	metadata, err := marshalJSONMap(entry.Metadata())
	if err != nil {
		return fmt.Errorf("audit repo: marshal metadata: %w", err)
	}

	if err := r.q(ctx).InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		ID:             entry.ID(),
		OrganizationID: uuidPtrToPgtype(entry.OrganizationID()),
		WorkspaceID:    uuidPtrToPgtype(entry.WorkspaceID()),
		ActorType:      string(entry.ActorType()),
		ActorID:        uuidPtrToPgtype(entry.ActorID()),
		Action:         entry.Action(),
		ResourceType:   entry.ResourceType(),
		ResourceID:     uuidPtrToPgtype(entry.ResourceID()),
		RequestID:      strPtr(entry.RequestID()),
		CorrelationID:  strPtr(entry.CorrelationID()),
		Ip:             ipPtr(entry.IP()),
		UserAgent:      strPtr(entry.UserAgent()),
		Metadata:       metadata,
		CreatedAt:      timestamptz(entry.CreatedAt()),
	}); err != nil {
		return fmt.Errorf("audit repo: insert audit log: %w", err)
	}
	return nil
}
