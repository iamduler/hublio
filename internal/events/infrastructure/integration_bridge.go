package infrastructure

import (
	"context"
	"strings"

	eventsapp "hublio/internal/events/application"
	eventsdomain "hublio/internal/events/domain"
	integrationapp "hublio/internal/integration/application"
	integrationdomain "hublio/internal/integration/domain"

	"github.com/google/uuid"
)

// IntegrationEventBridge adapts the Integration BC's own EventPublisher port to the Events
// BC Publisher. See OrchestrationEventBridge for the general pattern.
//
// The frozen `aggregate_type` enum has no "credential" member, so Credential events are
// attributed to their owning Connection using the connection_id already present on the
// domain Event's Payload.
type IntegrationEventBridge struct {
	publisher eventsapp.Publisher
}

func NewIntegrationEventBridge(publisher eventsapp.Publisher) *IntegrationEventBridge {
	return &IntegrationEventBridge{publisher: publisher}
}

var _ integrationapp.EventPublisher = (*IntegrationEventBridge)(nil)

func (b *IntegrationEventBridge) Publish(ctx context.Context, events ...integrationdomain.Event) error {
	if len(events) == 0 {
		return nil
	}
	inputs := make([]eventsapp.PublishInput, 0, len(events))
	for _, e := range events {
		inputs = append(inputs, mapIntegrationEvent(e))
	}
	return b.publisher.Publish(ctx, inputs...)
}

func mapIntegrationEvent(e integrationdomain.Event) eventsapp.PublishInput {
	var aggregateType eventsdomain.AggregateType
	var aggregateID uuid.UUID

	switch {
	case strings.HasPrefix(e.Name, "Connector"):
		aggregateType, aggregateID = eventsdomain.AggregateTypeConnector, e.AggregateID
	case strings.HasPrefix(e.Name, "Credential"):
		aggregateType = eventsdomain.AggregateTypeConnection
		aggregateID = uuidPtrOrSelf(uuidFromPayload(e.Payload, "connection_id"), e.AggregateID)
	default: // Connection*
		aggregateType, aggregateID = eventsdomain.AggregateTypeConnection, e.AggregateID
	}

	return eventsapp.PublishInput{
		WorkspaceID:   uuidFromPayload(e.Payload, "workspace_id"),
		AggregateType: string(aggregateType),
		AggregateID:   aggregateID,
		Category:      string(eventsdomain.CategorySystem),
		EventName:     e.Name,
		Payload:       e.Payload,
		PublishedBy:   "integration",
	}
}

// IntegrationAuditBridge adapts Integration's AuditRecorder port to the Events BC Auditor.
type IntegrationAuditBridge struct {
	auditor eventsapp.Auditor
}

func NewIntegrationAuditBridge(auditor eventsapp.Auditor) *IntegrationAuditBridge {
	return &IntegrationAuditBridge{auditor: auditor}
}

var _ integrationapp.AuditRecorder = (*IntegrationAuditBridge)(nil)

func (b *IntegrationAuditBridge) Record(ctx context.Context, rec integrationapp.AuditEvent) error {
	return b.auditor.Record(ctx, auditRecordFromContext(ctx, rec.ActorType, rec.ActorID, rec.Action, rec.ResourceType, rec.ResourceID, rec.Metadata))
}
