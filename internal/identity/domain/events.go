package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventOrganizationCreated   = "OrganizationCreated"
	EventOrganizationUpdated   = "OrganizationUpdated"
	EventOrganizationSuspended = "OrganizationSuspended"
	EventOrganizationActivated = "OrganizationActivated"
	EventOrganizationArchived  = "OrganizationArchived"

	EventWorkspaceCreated  = "WorkspaceCreated"
	EventWorkspaceUpdated  = "WorkspaceUpdated"
	EventWorkspaceEnabled  = "WorkspaceEnabled"
	EventWorkspaceDisabled = "WorkspaceDisabled"

	EventUserCreated = "UserCreated"

	EventMembershipAdded = "MembershipAdded"

	EventAPIKeyCreated  = "ApiKeyCreated"
	EventAPIKeyDisabled = "ApiKeyDisabled"
	EventAPIKeyRotated  = "ApiKeyRotated"
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

func (r *eventRecorder) PullEvents() []Event {
	out := r.events
	r.events = nil
	return out
}
