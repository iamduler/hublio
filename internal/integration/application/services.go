package application

import (
	"context"
	"time"

	"hublio/internal/integration/domain"

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

// SecretEncryptor encrypts/decrypts Credential secrets at the Application/Infrastructure boundary.
// The Domain never sees plaintext and never calls this port directly.
type SecretEncryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// AuditRecorder records best-effort audit facts after a successful commit (F2). Never fails
// the calling use case: implementations (a bridge over the Events BC Auditor, wired in the
// composition root) log their own failures.
type AuditRecorder interface {
	Record(ctx context.Context, rec AuditEvent) error
}

// AuditEvent is the minimal, BC-local shape passed to AuditRecorder. Tenant/request context
// is filled in by the bridge from context, so Integration never needs to know about it.
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

type Services struct {
	Connectors  domain.ConnectorRepository
	Connections domain.ConnectionRepository
	Credentials domain.CredentialRepository
	Runtimes    domain.RuntimeRegistry
	Secrets     SecretEncryptor
	Events      EventPublisher
	Audit       AuditRecorder
	Clock       Clock
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

func (s *Services) PublishAfterCommit(ctx context.Context, events ...domain.Event) {
	if len(events) == 0 {
		return
	}
	_ = s.events().Publish(ctx, events...)
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

// EnrichEvents stamps workspace_id onto every Event's Payload when not already present (e.g.
// ConnectionEnabled/Disabled/Verified only record a minimal payload today). See
// orchestration/application.EnrichEvents for the rationale: call sites own tenant context,
// Domain stays free of it.
func EnrichEvents(events []domain.Event, workspaceID uuid.UUID) []domain.Event {
	for i := range events {
		payload := events[i].Payload
		if payload == nil {
			payload = map[string]any{}
		} else {
			cloned := make(map[string]any, len(payload)+1)
			for k, v := range payload {
				cloned[k] = v
			}
			payload = cloned
		}
		if _, ok := payload["workspace_id"]; !ok && workspaceID != uuid.Nil {
			payload["workspace_id"] = workspaceID.String()
		}
		events[i].Payload = payload
	}
	return events
}
