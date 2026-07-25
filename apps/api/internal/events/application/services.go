package application

import (
	"time"

	"hublio/internal/events/domain"
	"hublio/internal/platform/metrics"
)

// Clock abstracts time for use cases (tests can inject fixed clocks).
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// Services wires the Events BC use cases: Publish (PlatformEvent, F1) and Record (AuditEntry,
// F2). It is the single implementation other Bounded Contexts talk to via thin bridges
// (internal/events/infrastructure) so Identity/Integration/Orchestration never import this
// package's types directly.
type Services struct {
	Events  domain.EventRepository
	Reader  domain.EventReader
	Audit   domain.AuditRepository
	Metrics *metrics.Counters
	Clock   Clock

	// OnSubscriberError is an optional hook invoked when an in-process subscriber returns an
	// error (best-effort delivery: the error never fails Publish). Composition roots wire
	// this to structured logging without Application importing the logging package.
	OnSubscriberError func(event *domain.PlatformEvent, err error)

	subscriptions []subscription
}

func (s *Services) clock() Clock {
	if s.Clock != nil {
		return s.Clock
	}
	return systemClock{}
}

func (s *Services) metricsOrNoop() *metrics.Counters {
	if s.Metrics != nil {
		return s.Metrics
	}
	return metrics.New()
}
