package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type IntentStatus string

const (
	IntentStatusSubmitted IntentStatus = "submitted"
	IntentStatusAccepted  IntentStatus = "accepted"
	IntentStatusRejected  IntentStatus = "rejected"
	IntentStatusExpired   IntentStatus = "expired"
)

// Intent is the Business Intent aggregate: a tenant's request to run a capability against
// a Connection. States: Submitted -> Accepted | Rejected | Expired. Accepted is immutable.
type Intent struct {
	eventRecorder

	id             uuid.UUID
	organizationID uuid.UUID
	workspaceID    uuid.UUID
	connectionID   uuid.UUID
	capability     string
	payload        map[string]any
	status         IntentStatus
	correlationID  string
	idempotencyKey string
	submittedAt    time.Time
	createdAt      time.Time
}

// NewIntent creates a Submitted Intent. Structural fields (tenant/connection ids) are
// required; capability/payload business validation happens explicitly via Accept/Reject
// so the outcome is observable on the Intent itself rather than a hard 400.
func NewIntent(
	id, organizationID, workspaceID, connectionID uuid.UUID,
	capability string,
	payload map[string]any,
	correlationID, idempotencyKey string,
	now time.Time,
) (*Intent, error) {
	if id == uuid.Nil || organizationID == uuid.Nil || workspaceID == uuid.Nil || connectionID == uuid.Nil {
		return nil, ErrInvalidID
	}

	intent := &Intent{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		connectionID:   connectionID,
		capability:     strings.TrimSpace(capability),
		payload:        payload,
		status:         IntentStatusSubmitted,
		correlationID:  strings.TrimSpace(correlationID),
		idempotencyKey: strings.TrimSpace(idempotencyKey),
		submittedAt:    now.UTC(),
		createdAt:      now.UTC(),
	}
	intent.record(EventIntentSubmitted, id, now.UTC(), map[string]any{
		"organization_id": organizationID.String(),
		"workspace_id":    workspaceID.String(),
		"connection_id":   connectionID.String(),
		"capability":      intent.capability,
	})
	return intent, nil
}

func ReconstituteIntent(
	id, organizationID, workspaceID, connectionID uuid.UUID,
	capability string,
	payload map[string]any,
	status IntentStatus,
	correlationID, idempotencyKey string,
	submittedAt, createdAt time.Time,
) *Intent {
	return &Intent{
		id:             id,
		organizationID: organizationID,
		workspaceID:    workspaceID,
		connectionID:   connectionID,
		capability:     capability,
		payload:        payload,
		status:         status,
		correlationID:  correlationID,
		idempotencyKey: idempotencyKey,
		submittedAt:    submittedAt,
		createdAt:      createdAt,
	}
}

func (i *Intent) ID() uuid.UUID             { return i.id }
func (i *Intent) OrganizationID() uuid.UUID { return i.organizationID }
func (i *Intent) WorkspaceID() uuid.UUID    { return i.workspaceID }
func (i *Intent) ConnectionID() uuid.UUID   { return i.connectionID }
func (i *Intent) Capability() string        { return i.capability }
func (i *Intent) Payload() map[string]any   { return i.payload }
func (i *Intent) Status() IntentStatus      { return i.status }
func (i *Intent) CorrelationID() string     { return i.correlationID }
func (i *Intent) IdempotencyKey() string    { return i.idempotencyKey }
func (i *Intent) SubmittedAt() time.Time    { return i.submittedAt }
func (i *Intent) CreatedAt() time.Time      { return i.createdAt }

// IsImmutable reports whether the Intent can no longer be transitioned (Accepted).
func (i *Intent) IsImmutable() bool {
	return i.status == IntentStatusAccepted
}

// IsValid reports whether the Intent carries the minimum business data required to be
// Accepted: a non-empty capability and a non-empty payload.
func (i *Intent) IsValid() bool {
	return i.capability != "" && len(i.payload) > 0
}

// Accept transitions a Submitted Intent to Accepted. Terminal; the Intent becomes immutable.
func (i *Intent) Accept(now time.Time) error {
	if i.status != IntentStatusSubmitted {
		return ErrInvalidTransition
	}
	i.status = IntentStatusAccepted
	i.record(EventIntentAccepted, i.id, now.UTC(), nil)
	return nil
}

// Reject transitions a Submitted Intent to Rejected. Terminal.
func (i *Intent) Reject(reason string, now time.Time) error {
	if i.status != IntentStatusSubmitted {
		return ErrInvalidTransition
	}
	i.status = IntentStatusRejected
	i.record(EventIntentRejected, i.id, now.UTC(), map[string]any{"reason": reason})
	return nil
}

// Expire transitions a Submitted Intent to Expired. Terminal.
func (i *Intent) Expire(now time.Time) error {
	if i.status != IntentStatusSubmitted {
		return ErrInvalidTransition
	}
	i.status = IntentStatusExpired
	i.record(EventIntentExpired, i.id, now.UTC(), nil)
	return nil
}
