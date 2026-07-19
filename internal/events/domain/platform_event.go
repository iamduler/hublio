package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Category classifies a PlatformEvent per the Architecture Freeze: runtime facts (Intent/
// Execution lifecycle), business facts, and system facts (Identity/Integration lifecycle).
type Category string

const (
	CategoryRuntime  Category = "runtime"
	CategoryBusiness Category = "business"
	CategorySystem   Category = "system"
)

func (c Category) valid() bool {
	switch c {
	case CategoryRuntime, CategoryBusiness, CategorySystem:
		return true
	default:
		return false
	}
}

// AggregateType matches the frozen `aggregate_type` enum (docs/20-database-schema.dbml).
// It intentionally has no "user" / "api_key" / "credential" member: events for those
// entities are attributed to the closest owning Aggregate (see the events BC bridges).
type AggregateType string

const (
	AggregateTypeOrganization AggregateType = "organization"
	AggregateTypeWorkspace    AggregateType = "workspace"
	AggregateTypeConnector    AggregateType = "connector"
	AggregateTypeConnection   AggregateType = "connection"
	AggregateTypeIntent       AggregateType = "intent"
	AggregateTypeExecution    AggregateType = "execution"
)

func (t AggregateType) valid() bool {
	switch t {
	case AggregateTypeOrganization, AggregateTypeWorkspace, AggregateTypeConnector,
		AggregateTypeConnection, AggregateTypeIntent, AggregateTypeExecution:
		return true
	default:
		return false
	}
}

// PlatformEvent is an immutable, append-only fact recorded in the `events` table. It is
// never mutated after creation: the Repository port only exposes Save (insert).
type PlatformEvent struct {
	id             uuid.UUID
	organizationID *uuid.UUID
	workspaceID    *uuid.UUID
	aggregateType  AggregateType
	aggregateID    uuid.UUID
	executionID    *uuid.UUID
	category       Category
	eventName      string
	correlationID  string
	payload        map[string]any
	metadata       map[string]any
	publishedBy    string
	createdAt      time.Time
}

// NewPlatformEvent validates and constructs an immutable PlatformEvent. id must be an
// application-generated UUID v7 (Domain never generates ids).
func NewPlatformEvent(
	id uuid.UUID,
	organizationID, workspaceID *uuid.UUID,
	aggregateType AggregateType,
	aggregateID uuid.UUID,
	executionID *uuid.UUID,
	category Category,
	eventName, correlationID string,
	payload, metadata map[string]any,
	publishedBy string,
	now time.Time,
) (*PlatformEvent, error) {
	if id == uuid.Nil || aggregateID == uuid.Nil {
		return nil, ErrInvalidID
	}
	if !aggregateType.valid() {
		return nil, ErrInvalidAggregateType
	}
	if !category.valid() {
		return nil, ErrInvalidCategory
	}
	if eventName == "" || len(eventName) > 150 {
		return nil, ErrInvalidEventName
	}

	return &PlatformEvent{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		aggregateType:  aggregateType,
		aggregateID:    aggregateID,
		executionID:    executionID,
		category:       category,
		eventName:      eventName,
		correlationID:  correlationID,
		payload:        payload,
		metadata:       metadata,
		publishedBy:    publishedBy,
		createdAt:      now.UTC(),
	}, nil
}

// ReconstitutePlatformEvent hydrates a PlatformEvent from persistence without re-validating
// (the row is already known-good, immutable data).
func ReconstitutePlatformEvent(
	id uuid.UUID,
	organizationID, workspaceID *uuid.UUID,
	aggregateType AggregateType,
	aggregateID uuid.UUID,
	executionID *uuid.UUID,
	category Category,
	eventName, correlationID string,
	payload, metadata map[string]any,
	publishedBy string,
	createdAt time.Time,
) *PlatformEvent {
	return &PlatformEvent{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		aggregateType:  aggregateType,
		aggregateID:    aggregateID,
		executionID:    executionID,
		category:       category,
		eventName:      eventName,
		correlationID:  correlationID,
		payload:        payload,
		metadata:       metadata,
		publishedBy:    publishedBy,
		createdAt:      createdAt,
	}
}

func (e *PlatformEvent) ID() uuid.UUID                { return e.id }
func (e *PlatformEvent) OrganizationID() *uuid.UUID   { return e.organizationID }
func (e *PlatformEvent) WorkspaceID() *uuid.UUID      { return e.workspaceID }
func (e *PlatformEvent) AggregateType() AggregateType { return e.aggregateType }
func (e *PlatformEvent) AggregateID() uuid.UUID       { return e.aggregateID }
func (e *PlatformEvent) ExecutionID() *uuid.UUID      { return e.executionID }
func (e *PlatformEvent) Category() Category           { return e.category }
func (e *PlatformEvent) EventName() string            { return e.eventName }
func (e *PlatformEvent) CorrelationID() string        { return e.correlationID }
func (e *PlatformEvent) Payload() map[string]any      { return e.payload }
func (e *PlatformEvent) Metadata() map[string]any     { return e.metadata }
func (e *PlatformEvent) PublishedBy() string          { return e.publishedBy }
func (e *PlatformEvent) CreatedAt() time.Time         { return e.createdAt }

// EventRepository is the append-only persistence port for PlatformEvent. Only Save (insert)
// is exposed: events are immutable and never updated or deleted.
type EventRepository interface {
	Save(ctx context.Context, event *PlatformEvent) error
}

// EventReader is a read-only query port backing the Platform Events API
// (GET /api/v1/events). Kept separate from EventRepository so the write path stays
// literally Save-only.
type EventReader interface {
	// ListByWorkspace returns the most recent events for workspaceID, newest first,
	// optionally filtered to one executionID, bounded by limit.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, executionID *uuid.UUID, limit int32) ([]*PlatformEvent, error)
}

// EventHandler is an in-process subscriber invoked after a PlatformEvent has been durably
// persisted (Publish always persists before notifying handlers).
type EventHandler func(ctx context.Context, event *PlatformEvent) error
