package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusDisabled APIKeyStatus = "disabled"
)

// APIKey is a Workspace-scoped machine credential. Plaintext is never retained.
type APIKey struct {
	eventRecorder

	id          uuid.UUID
	workspaceID uuid.UUID
	name        string
	keyHash     string
	prefix      string
	status      APIKeyStatus
	expiresAt   *time.Time
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

func NewAPIKey(id, workspaceID uuid.UUID, name, keyHash, prefix string, expiresAt *time.Time, now time.Time) (*APIKey, error) {
	name = strings.TrimSpace(name)
	if id == uuid.Nil || workspaceID == uuid.Nil {
		return nil, ErrInvalidName
	}
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}
	if keyHash == "" || prefix == "" || len(prefix) > 10 {
		return nil, ErrInvalidName
	}

	k := &APIKey{
		id:          id,
		workspaceID: workspaceID,
		name:        name,
		keyHash:     keyHash,
		prefix:      prefix,
		status:      APIKeyStatusActive,
		expiresAt:   expiresAt,
		createdAt:   now.UTC(),
		updatedAt:   now.UTC(),
	}
	k.record(EventAPIKeyCreated, id, now.UTC(), map[string]any{
		"workspace_id": workspaceID.String(),
		"name":         name,
		"prefix":       prefix,
	})
	return k, nil
}

func ReconstituteAPIKey(
	id, workspaceID uuid.UUID,
	name, keyHash, prefix string,
	status APIKeyStatus,
	expiresAt *time.Time,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *APIKey {
	return &APIKey{
		id:          id,
		workspaceID: workspaceID,
		name:        name,
		keyHash:     keyHash,
		prefix:      prefix,
		status:      status,
		expiresAt:   expiresAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

func (k *APIKey) ID() uuid.UUID          { return k.id }
func (k *APIKey) WorkspaceID() uuid.UUID { return k.workspaceID }
func (k *APIKey) Name() string           { return k.name }
func (k *APIKey) KeyHash() string        { return k.keyHash }
func (k *APIKey) Prefix() string         { return k.prefix }
func (k *APIKey) Status() APIKeyStatus   { return k.status }
func (k *APIKey) ExpiresAt() *time.Time  { return k.expiresAt }
func (k *APIKey) CreatedAt() time.Time   { return k.createdAt }
func (k *APIKey) UpdatedAt() time.Time   { return k.updatedAt }
func (k *APIKey) DeletedAt() *time.Time  { return k.deletedAt }

func (k *APIKey) IsUsable(now time.Time) error {
	if k.status != APIKeyStatusActive || k.deletedAt != nil {
		return ErrAPIKeyDisabled
	}
	if k.expiresAt != nil && !k.expiresAt.After(now.UTC()) {
		return ErrAPIKeyExpired
	}
	return nil
}

func (k *APIKey) Disable(now time.Time) error {
	if k.status != APIKeyStatusActive {
		return ErrInvalidTransition
	}
	k.status = APIKeyStatusDisabled
	k.updatedAt = now.UTC()
	k.record(EventAPIKeyDisabled, k.id, k.updatedAt, nil)
	return nil
}

// Rotate replaces the stored hash/prefix. Caller must persist and return plaintext once.
func (k *APIKey) Rotate(keyHash, prefix string, now time.Time) error {
	if k.status != APIKeyStatusActive {
		return ErrInvalidTransition
	}
	if keyHash == "" || prefix == "" || len(prefix) > 10 {
		return ErrInvalidName
	}
	k.keyHash = keyHash
	k.prefix = prefix
	k.updatedAt = now.UTC()
	k.record(EventAPIKeyRotated, k.id, k.updatedAt, map[string]any{"prefix": prefix})
	return nil
}
