package application

import (
	"context"
	"time"

	"hublio/internal/orchestration/domain"

	"github.com/google/uuid"
)

// Clock abstracts time for use cases (tests can inject fixed clocks).
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// EventPublisher publishes domain events after a successful commit.
type EventPublisher interface {
	Publish(ctx context.Context, events ...domain.Event) error
}

// NoopPublisher discards events (wiring until Events BC is ready).
type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, events ...domain.Event) error {
	_ = ctx
	_ = events
	return nil
}

// ResolvedConnection carries everything RunExecution/SubmitIntent need to call a Connector
// Runtime, without leaking Integration domain/provider types into Orchestration.
type ResolvedConnection struct {
	ConnectionID  uuid.UUID
	WorkspaceID   uuid.UUID
	ConnectorID   uuid.UUID
	ConnectorCode string
	Config        map[string]any
	// Secret is the already-decrypted credential map. Callers must never persist or log it.
	Secret         map[string]any
	TimeoutSeconds int
}

// ConnectionGateway resolves and validates a Workspace-scoped Connection for use by an
// Intent/Execution, without the Orchestration Domain/Application importing Integration.
type ConnectionGateway interface {
	// ResolveForIntent returns an error if the Connection is not found, not Active, or does
	// not belong to workspaceID (also checks Workspace/Organization are able to run Intents).
	ResolveForIntent(ctx context.Context, workspaceID, connectionID uuid.UUID) (ResolvedConnection, error)
}

// InvokeRequest/InvokeResponse use canonical-ish map[string]any payloads only; provider DTOs
// never cross this boundary (they stay inside internal/integration/connectors/<vendor>).
type InvokeRequest struct {
	ConnectionID uuid.UUID
	Capability   string
	Config       map[string]any
	Secret       map[string]any
	Payload      map[string]any
}

type InvokeResponse struct {
	Payload  map[string]any
	Metadata map[string]any
}

// ConnectorGateway invokes a Connector Runtime by code. Implemented in
// internal/orchestration/infrastructure, wrapping the Integration Connector Runtime registry.
type ConnectorGateway interface {
	Invoke(ctx context.Context, connectorCode string, in InvokeRequest) (InvokeResponse, error)
}

// ExecutionJob is the Platform work-queue payload for the orchestration.execution job type.
type ExecutionJob struct {
	ExecutionID    uuid.UUID
	IntentID       uuid.UUID
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	CorrelationID  string
}

// JobEnqueuer schedules asynchronous Execution processing on the Platform work queue.
type JobEnqueuer interface {
	EnqueueExecution(ctx context.Context, job ExecutionJob) error
}

// Services wires the Orchestration use cases. MaxRetries defaults to 3 when <= 0.
type Services struct {
	Intents     domain.IntentRepository
	Executions  domain.ExecutionRepository
	Idempotency domain.IdempotencyRepository
	Connections ConnectionGateway
	Connectors  ConnectorGateway
	Jobs        JobEnqueuer
	Events      EventPublisher
	Clock       Clock
	MaxRetries  int
}

func (s *Services) clock() Clock {
	if s.Clock != nil {
		return s.Clock
	}
	return systemClock{}
}

func (s *Services) events() EventPublisher {
	if s.Events != nil {
		return s.Events
	}
	return NoopPublisher{}
}

func (s *Services) maxRetries() int {
	if s.MaxRetries > 0 {
		return s.MaxRetries
	}
	return 3
}

func (s *Services) PublishAfterCommit(ctx context.Context, events ...domain.Event) {
	if len(events) == 0 {
		return
	}
	_ = s.events().Publish(ctx, events...)
}
