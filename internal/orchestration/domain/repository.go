package domain

import (
	"context"

	"github.com/google/uuid"
)

// IntentRepository persists the Intent aggregate.
type IntentRepository interface {
	Save(ctx context.Context, intent *Intent) error
	Update(ctx context.Context, intent *Intent) error
	FindByID(ctx context.Context, id uuid.UUID) (*Intent, error)
}

// ExecutionRepository persists the Execution aggregate together with its Steps,
// Snapshots and Timeline children.
type ExecutionRepository interface {
	// Save inserts a new Execution and its (Pending) default Steps.
	Save(ctx context.Context, execution *Execution) error
	// Update persists Execution row changes, Step updates, and any newly appended
	// (append-only) Snapshots/Timeline entries.
	Update(ctx context.Context, execution *Execution) error
	FindByID(ctx context.Context, id uuid.UUID) (*Execution, error)
	FindByIntentID(ctx context.Context, intentID uuid.UUID) (*Execution, error)
}

// IdempotencyRepository persists Intent idempotency keys (Postgres is the source of truth;
// Redis may cache lookups but never replaces this store).
type IdempotencyRepository interface {
	Save(ctx context.Context, rec *IdempotencyKey) error
	FindByKey(ctx context.Context, organizationID, workspaceID uuid.UUID, key string) (*IdempotencyKey, error)
}
