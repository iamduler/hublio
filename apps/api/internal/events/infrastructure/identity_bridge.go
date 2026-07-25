package infrastructure

import (
	"context"
	"strings"

	eventsapp "hublio/internal/events/application"
	eventsdomain "hublio/internal/events/domain"
	identityapp "hublio/internal/identity/application"
	identitydomain "hublio/internal/identity/domain"

	"github.com/google/uuid"
)

// IdentityEventBridge adapts the Identity BC's own EventPublisher port to the Events BC
// Publisher. See OrchestrationEventBridge for the general pattern.
//
// The frozen `aggregate_type` enum has no "user" / "api_key" / "membership" member, so
// User/Membership/ApiKey events are attributed to their closest owning Aggregate
// (Organization or Workspace) using ids already present on the domain Event's Payload.
type IdentityEventBridge struct {
	publisher eventsapp.Publisher
}

func NewIdentityEventBridge(publisher eventsapp.Publisher) *IdentityEventBridge {
	return &IdentityEventBridge{publisher: publisher}
}

var _ identityapp.EventPublisher = (*IdentityEventBridge)(nil)

func (b *IdentityEventBridge) Publish(ctx context.Context, events ...identitydomain.Event) error {
	if len(events) == 0 {
		return nil
	}
	inputs := make([]eventsapp.PublishInput, 0, len(events))
	for _, e := range events {
		inputs = append(inputs, mapIdentityEvent(e))
	}
	return b.publisher.Publish(ctx, inputs...)
}

func mapIdentityEvent(e identitydomain.Event) eventsapp.PublishInput {
	var aggregateType eventsdomain.AggregateType
	var aggregateID uuid.UUID

	switch {
	case strings.HasPrefix(e.Name, "Organization"):
		aggregateType, aggregateID = eventsdomain.AggregateTypeOrganization, e.AggregateID
	case strings.HasPrefix(e.Name, "Workspace"):
		aggregateType, aggregateID = eventsdomain.AggregateTypeWorkspace, e.AggregateID
	case strings.HasPrefix(e.Name, "Membership"):
		aggregateType = eventsdomain.AggregateTypeWorkspace
		aggregateID = uuidPtrOrSelf(uuidFromPayload(e.Payload, "workspace_id"), e.AggregateID)
	case strings.HasPrefix(e.Name, "ApiKey"):
		aggregateType = eventsdomain.AggregateTypeWorkspace
		aggregateID = uuidPtrOrSelf(uuidFromPayload(e.Payload, "workspace_id"), e.AggregateID)
	default: // User* and any future Identity event
		aggregateType = eventsdomain.AggregateTypeOrganization
		aggregateID = uuidPtrOrSelf(uuidFromPayload(e.Payload, "organization_id"), e.AggregateID)
	}

	return eventsapp.PublishInput{
		OrganizationID: uuidFromPayload(e.Payload, "organization_id"),
		WorkspaceID:    uuidFromPayload(e.Payload, "workspace_id"),
		AggregateType:  string(aggregateType),
		AggregateID:    aggregateID,
		Category:       string(eventsdomain.CategorySystem),
		EventName:      e.Name,
		Payload:        e.Payload,
		PublishedBy:    "identity",
	}
}

// IdentityAuditBridge adapts Identity's AuditRecorder port to the Events BC Auditor.
type IdentityAuditBridge struct {
	auditor eventsapp.Auditor
}

func NewIdentityAuditBridge(auditor eventsapp.Auditor) *IdentityAuditBridge {
	return &IdentityAuditBridge{auditor: auditor}
}

var _ identityapp.AuditRecorder = (*IdentityAuditBridge)(nil)

func (b *IdentityAuditBridge) Record(ctx context.Context, rec identityapp.AuditEvent) error {
	return b.auditor.Record(ctx, auditRecordFromContext(ctx, rec.ActorType, rec.ActorID, rec.Action, rec.ResourceType, rec.ResourceID, rec.Metadata))
}
