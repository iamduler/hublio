package application

import (
	"context"
	"time"

	"hublio/internal/identity/domain"

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

// NoopPublisher discards events (Phase B wiring until Events BC is ready).
type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, events ...domain.Event) error {
	_ = ctx
	_ = events
	return nil
}

// AuditRecorder records best-effort audit facts after a successful commit (F2). Never fails
// the calling use case: implementations (a bridge over the Events BC Auditor, wired in the
// composition root) log their own failures.
type AuditRecorder interface {
	Record(ctx context.Context, rec AuditEvent) error
}

// AuditEvent is the minimal, BC-local shape passed to AuditRecorder. Tenant/request context
// (organization_id, workspace_id, request_id, correlation_id, ip, user_agent) is filled in by
// the bridge from context, so Identity never needs to know about those fields.
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
	Orgs        domain.OrganizationRepository
	Workspaces  domain.WorkspaceRepository
	Users       domain.UserRepository
	Memberships domain.MembershipRepository
	APIKeys     domain.APIKeyRepository
	Passwords   domain.PasswordHasher
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

// EnrichEvents stamps organization_id/workspace_id onto every Event's Payload when not
// already present (e.g. ApiKeyDisabled/Rotated only record a minimal payload today). See
// orchestration/application.EnrichEvents for the rationale: call sites own tenant context,
// Domain stays free of it.
func EnrichEvents(events []domain.Event, organizationID, workspaceID uuid.UUID) []domain.Event {
	for i := range events {
		payload := events[i].Payload
		if payload == nil {
			payload = map[string]any{}
		} else {
			cloned := make(map[string]any, len(payload)+2)
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
		events[i].Payload = payload
	}
	return events
}
