package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventIntentSubmitted = "IntentSubmitted"
	EventIntentAccepted  = "IntentAccepted"
	EventIntentRejected  = "IntentRejected"
	EventIntentExpired   = "IntentExpired"

	EventExecutionCreated        = "ExecutionCreated"
	EventExecutionQueued         = "ExecutionQueued"
	EventExecutionStarted        = "ExecutionStarted"
	EventExecutionSucceeded      = "ExecutionSucceeded"
	EventExecutionFailed         = "ExecutionFailed"
	EventExecutionCancelled      = "ExecutionCancelled"
	EventExecutionExpired        = "ExecutionExpired"
	EventExecutionRetryScheduled = "ExecutionRetryScheduled"
	EventExecutionDeadLettered   = "ExecutionDeadLettered"
)

// Event is an immutable domain fact recorded on aggregates.
type Event struct {
	Name        string
	AggregateID uuid.UUID
	OccurredAt  time.Time
	Payload     map[string]any
}

type eventRecorder struct {
	events []Event
}

func (r *eventRecorder) record(name string, id uuid.UUID, at time.Time, payload map[string]any) {
	r.events = append(r.events, Event{
		Name:        name,
		AggregateID: id,
		OccurredAt:  at,
		Payload:     payload,
	})
}

// PullEvents drains and returns recorded events; subsequent calls return nil until new events are recorded.
func (r *eventRecorder) PullEvents() []Event {
	out := r.events
	r.events = nil
	return out
}
