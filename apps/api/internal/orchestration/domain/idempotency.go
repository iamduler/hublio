package domain

import (
	"time"

	"github.com/google/uuid"
)

// IdempotencyKey links a Workspace-scoped idempotency key to the Intent it produced.
// Postgres idempotency_keys is the source of truth (docs/20-database-schema.dbml).
type IdempotencyKey struct {
	id             uuid.UUID
	organizationID uuid.UUID
	workspaceID    uuid.UUID
	key            string
	intentID       *uuid.UUID
	expiresAt      *time.Time
	createdAt      time.Time
}

func NewIdempotencyKey(id, organizationID, workspaceID uuid.UUID, key string, intentID uuid.UUID, expiresAt *time.Time, now time.Time) (*IdempotencyKey, error) {
	if id == uuid.Nil || organizationID == uuid.Nil || workspaceID == uuid.Nil || key == "" {
		return nil, ErrInvalidID
	}
	return &IdempotencyKey{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		key:            key,
		intentID:       &intentID,
		expiresAt:      expiresAt,
		createdAt:      now.UTC(),
	}, nil
}

func ReconstituteIdempotencyKey(
	id, organizationID, workspaceID uuid.UUID,
	key string,
	intentID *uuid.UUID,
	expiresAt *time.Time,
	createdAt time.Time,
) *IdempotencyKey {
	return &IdempotencyKey{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		key:            key,
		intentID:       intentID,
		expiresAt:      expiresAt,
		createdAt:      createdAt,
	}
}

func (k *IdempotencyKey) ID() uuid.UUID             { return k.id }
func (k *IdempotencyKey) OrganizationID() uuid.UUID { return k.organizationID }
func (k *IdempotencyKey) WorkspaceID() uuid.UUID    { return k.workspaceID }
func (k *IdempotencyKey) Key() string               { return k.key }
func (k *IdempotencyKey) IntentID() *uuid.UUID      { return k.intentID }
func (k *IdempotencyKey) ExpiresAt() *time.Time     { return k.expiresAt }
func (k *IdempotencyKey) CreatedAt() time.Time      { return k.createdAt }

// IsExpired reports whether the key's TTL (if any) has elapsed as of now.
func (k *IdempotencyKey) IsExpired(now time.Time) bool {
	return k.expiresAt != nil && !k.expiresAt.After(now.UTC())
}
