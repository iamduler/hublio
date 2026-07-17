package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventConnectorRegistered      = "ConnectorRegistered"
	EventConnectorEnabled         = "ConnectorEnabled"
	EventConnectorDisabled        = "ConnectorDisabled"
	EventConnectorRemoved         = "ConnectorRemoved"
	EventConnectorCapabilityAdded = "ConnectorCapabilityAdded"

	EventConnectionCreated            = "ConnectionCreated"
	EventConnectionVerifying          = "ConnectionVerifying"
	EventConnectionVerified           = "ConnectionVerified"
	EventConnectionVerificationFailed = "ConnectionVerificationFailed"
	EventConnectionEnabled            = "ConnectionEnabled"
	EventConnectionDisabled           = "ConnectionDisabled"

	EventCredentialCreated = "CredentialCreated"
	EventCredentialRevoked = "CredentialRevoked"
	EventCredentialRotated = "CredentialRotated"
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
