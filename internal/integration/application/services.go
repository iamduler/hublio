package application

import (
	"context"
	"time"

	"hublio/internal/integration/domain"
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

type Services struct {
	Connectors  domain.ConnectorRepository
	Connections domain.ConnectionRepository
	Credentials domain.CredentialRepository
	Runtimes    domain.RuntimeRegistry
	Secrets     SecretEncryptor
	Events      EventPublisher
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
