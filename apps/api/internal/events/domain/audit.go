package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ActorType matches the frozen `actor_type` enum (docs/20-database-schema.dbml).
type ActorType string

const (
	ActorTypeUser   ActorType = "user"
	ActorTypeAPIKey ActorType = "api_key"
	ActorTypeSystem ActorType = "system"
)

func (t ActorType) valid() bool {
	switch t {
	case ActorTypeUser, ActorTypeAPIKey, ActorTypeSystem:
		return true
	default:
		return false
	}
}

// AuditEntry is an immutable, append-only security/compliance record in `audit_logs`. Never
// carries secrets: callers (Application layer) are responsible for redacting Metadata
// before constructing an AuditEntry.
type AuditEntry struct {
	id             uuid.UUID
	organizationID *uuid.UUID
	workspaceID    *uuid.UUID
	actorType      ActorType
	actorID        *uuid.UUID
	action         string
	resourceType   string
	resourceID     *uuid.UUID
	requestID      string
	correlationID  string
	ip             string
	userAgent      string
	metadata       map[string]any
	createdAt      time.Time
}

// NewAuditEntry validates and constructs an immutable AuditEntry. id must be an
// application-generated UUID v7.
func NewAuditEntry(
	id uuid.UUID,
	organizationID, workspaceID *uuid.UUID,
	actorType ActorType,
	actorID *uuid.UUID,
	action, resourceType string,
	resourceID *uuid.UUID,
	requestID, correlationID, ip, userAgent string,
	metadata map[string]any,
	now time.Time,
) (*AuditEntry, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}
	if !actorType.valid() {
		return nil, ErrInvalidActorType
	}
	if action == "" || len(action) > 150 {
		return nil, ErrInvalidAction
	}
	if resourceType == "" || len(resourceType) > 100 {
		return nil, ErrInvalidResourceType
	}

	return &AuditEntry{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		actorType:      actorType,
		actorID:        actorID,
		action:         action,
		resourceType:   resourceType,
		resourceID:     resourceID,
		requestID:      requestID,
		correlationID:  correlationID,
		ip:             ip,
		userAgent:      userAgent,
		metadata:       metadata,
		createdAt:      now.UTC(),
	}, nil
}

func (e *AuditEntry) ID() uuid.UUID              { return e.id }
func (e *AuditEntry) OrganizationID() *uuid.UUID { return e.organizationID }
func (e *AuditEntry) WorkspaceID() *uuid.UUID    { return e.workspaceID }
func (e *AuditEntry) ActorType() ActorType       { return e.actorType }
func (e *AuditEntry) ActorID() *uuid.UUID        { return e.actorID }
func (e *AuditEntry) Action() string             { return e.action }
func (e *AuditEntry) ResourceType() string       { return e.resourceType }
func (e *AuditEntry) ResourceID() *uuid.UUID     { return e.resourceID }
func (e *AuditEntry) RequestID() string          { return e.requestID }
func (e *AuditEntry) CorrelationID() string      { return e.correlationID }
func (e *AuditEntry) IP() string                 { return e.ip }
func (e *AuditEntry) UserAgent() string          { return e.userAgent }
func (e *AuditEntry) Metadata() map[string]any   { return e.metadata }
func (e *AuditEntry) CreatedAt() time.Time       { return e.createdAt }

// AuditRepository is the append-only persistence port for AuditEntry.
type AuditRepository interface {
	Save(ctx context.Context, entry *AuditEntry) error
}
