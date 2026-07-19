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

// Transformer runs Canonical → Canonical normalization for a Step (Transformation BC). It is
// implemented in internal/orchestration/infrastructure, wrapping the Transformation
// Application; capability is passed only so the adapter can pick which built-in
// normalization spec applies (e.g. invoice-like capabilities) — Orchestration itself never
// inspects Provider DTOs or Transformation internals.
type Transformer interface {
	TransformRequest(ctx context.Context, capability string, doc map[string]any) (map[string]any, error)
	TransformResponse(ctx context.Context, capability string, doc map[string]any) (map[string]any, error)
}

// passthroughTransformer is the safe default when Services.Transformer is not wired (e.g. in
// unit tests): it returns the Document unchanged, exactly like an identity Pipeline.
type passthroughTransformer struct{}

func (passthroughTransformer) TransformRequest(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	_, _ = ctx, capability
	return doc, nil
}

func (passthroughTransformer) TransformResponse(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	_, _ = ctx, capability
	return doc, nil
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

// AuditRecorder records best-effort audit facts after a successful commit (F2). Never fails
// the calling use case: implementations (a bridge over the Events BC Auditor, wired in the
// composition root) log their own failures.
type AuditRecorder interface {
	Record(ctx context.Context, rec AuditEvent) error
}

// AuditEvent is the minimal, BC-local shape passed to AuditRecorder. Tenant/request context
// is filled in by the bridge from context, so Orchestration never needs to know about it.
type AuditEvent struct {
	ActorType    string // "user" | "api_key" | "system"
	ActorID      uuid.UUID
	Action       string
	ResourceType string
	ResourceID   uuid.UUID
	Metadata     map[string]any
}

// NoopAuditor discards audit records (wiring default until Events BC is composed).
type NoopAuditor struct{}

func (NoopAuditor) Record(ctx context.Context, rec AuditEvent) error {
	_, _ = ctx, rec
	return nil
}

// SyncRouteGateway resolves an Enabled SyncRoute for inbound webhook ingress without leaking
// Integration types into Orchestration Domain. Implemented in orchestration/infrastructure.
type SyncRouteGateway interface {
	ResolveWebhook(ctx context.Context, in ResolveWebhookInput) (ResolvedWebhookRoute, error)
}

type ResolveWebhookInput struct {
	SyncRouteID  uuid.UUID
	SecretHeader string
	ResourceType string
	Payload      map[string]any
}

// ResolvedWebhookRoute carries tenant + primary destination for AcceptWebhook → SubmitIntent.
type ResolvedWebhookRoute struct {
	SyncRouteID        uuid.UUID
	OrganizationID     uuid.UUID
	WorkspaceID        uuid.UUID
	ConnectionID       uuid.UUID // primary destination (v1 single Execution)
	Capability         string
	IdempotencyRule    map[string]any
	SourceConnectionID uuid.UUID
}

// Services wires the Orchestration use cases. MaxRetries defaults to 3 when <= 0.
type Services struct {
	Intents     domain.IntentRepository
	Executions  domain.ExecutionRepository
	Idempotency domain.IdempotencyRepository
	Connections ConnectionGateway
	Connectors  ConnectorGateway
	SyncRoutes  SyncRouteGateway
	Transformer Transformer
	Jobs        JobEnqueuer
	Events      EventPublisher
	Audit       AuditRecorder
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

func (s *Services) transformer() Transformer {
	if s.Transformer != nil {
		return s.Transformer
	}
	return passthroughTransformer{}
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

// EnrichEvents stamps organization_id/workspace_id/correlation_id onto every Event's Payload
// when not already present, so the Events BC bridge (internal/events/infrastructure) can
// populate the events table's tenant/correlation columns without Orchestration's Domain
// having to carry that context on every Execution/Intent event itself. Call sites (the
// Interfaces handler and the worker) own this because they are the ones with the
// request/job-scoped tenant context; Domain stays free of it. Never overwrites an existing
// key (e.g. IntentSubmitted already sets organization_id/workspace_id itself).
func EnrichEvents(events []domain.Event, organizationID, workspaceID uuid.UUID, correlationID string) []domain.Event {
	for i := range events {
		payload := events[i].Payload
		if payload == nil {
			payload = map[string]any{}
		} else {
			cloned := make(map[string]any, len(payload)+3)
			for k, v := range payload {
				cloned[k] = v
			}
			payload = cloned
		}
		if _, ok := payload["organization_id"]; !ok && organizationID != uuid.Nil {
			payload["organization_id"] = organizationID.String()
		}
		if _, ok := payload["workspace_id"]; !ok && workspaceID != uuid.Nil {
			payload["workspace_id"] = workspaceID.String()
		}
		if _, ok := payload["correlation_id"]; !ok && correlationID != "" {
			payload["correlation_id"] = correlationID
		}
		events[i].Payload = payload
	}
	return events
}

func (s *Services) audit() AuditRecorder {
	if s.Audit != nil {
		return s.Audit
	}
	return NoopAuditor{}
}

// RecordAudit is best-effort: it never fails the caller (see AuditRecorder).
func (s *Services) RecordAudit(ctx context.Context, rec AuditEvent) {
	_ = s.audit().Record(ctx, rec)
}
