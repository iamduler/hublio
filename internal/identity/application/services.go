package application

import (
	"context"
	"time"

	"hublio/internal/identity/domain"
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

type Services struct {
	Orgs         domain.OrganizationRepository
	Workspaces   domain.WorkspaceRepository
	Users        domain.UserRepository
	Memberships  domain.MembershipRepository
	APIKeys      domain.APIKeyRepository
	Passwords    domain.PasswordHasher
	Events       EventPublisher
	Clock        Clock
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
