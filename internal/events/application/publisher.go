package application

import (
	"context"

	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

// Publisher is the Events BC port other Bounded Contexts publish through (via a bridge —
// see internal/events/infrastructure). Never use this to schedule Execution work: the
// Event Platform is not a work queue (AGENTS.md).
type Publisher interface {
	Publish(ctx context.Context, events ...PublishInput) error
}

// PublishInput is the Application-level shape of a fact to persist+deliver. OrganizationID/
// WorkspaceID/ExecutionID are nil when unknown (all NULLABLE in the `events` table).
type PublishInput struct {
	OrganizationID *uuid.UUID
	WorkspaceID    *uuid.UUID
	AggregateType  string
	AggregateID    uuid.UUID
	ExecutionID    *uuid.UUID
	Category       string // runtime|business|system
	EventName      string
	CorrelationID  string
	Payload        map[string]any
	Metadata       map[string]any
	PublishedBy    string
}

// Publish persists every input (append-only, `events` table) and only then notifies
// in-process subscribers (at-least-once delivery). Persistence failures abort the whole
// call (callers publish after their own commit, so a persistence failure here is already
// post-commit and must be surfaced); subscriber failures never fail Publish — they are
// reported via s.Metrics/subscriber-owned error handling only, never propagated.
func (s *Services) Publish(ctx context.Context, inputs ...PublishInput) error {
	for _, in := range inputs {
		if err := s.publishOne(ctx, in); err != nil {
			return err
		}
	}
	return nil
}

func (s *Services) publishOne(ctx context.Context, in PublishInput) error {
	eventID, err := id.NewV7()
	if err != nil {
		return apperr.Wrap(err, "failed to generate event id", apperr.ErrCodeInternal)
	}

	event, err := domain.NewPlatformEvent(
		eventID,
		in.OrganizationID,
		in.WorkspaceID,
		domain.AggregateType(in.AggregateType),
		in.AggregateID,
		in.ExecutionID,
		domain.Category(in.Category),
		in.EventName,
		in.CorrelationID,
		in.Payload,
		in.Metadata,
		in.PublishedBy,
		s.clock().Now(),
	)
	if err != nil {
		return apperr.Wrap(err, "invalid platform event", apperr.ErrCodeBadRequest)
	}

	if err := s.eventRepo().Save(ctx, event); err != nil {
		return apperr.Wrap(err, "failed to persist platform event", apperr.ErrCodeInternal)
	}
	s.metricsOrNoop().IncEventsPublished()

	s.notifySubscribers(ctx, event)
	return nil
}

func (s *Services) eventRepo() domain.EventRepository {
	if s.Events != nil {
		return s.Events
	}
	return noopEventRepository{}
}

type noopEventRepository struct{}

func (noopEventRepository) Save(ctx context.Context, event *domain.PlatformEvent) error {
	_, _ = ctx, event
	return nil
}
