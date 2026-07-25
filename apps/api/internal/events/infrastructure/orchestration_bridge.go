package infrastructure

import (
	"context"
	"strings"

	eventsapp "hublio/internal/events/application"
	eventsdomain "hublio/internal/events/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	orchestrationdomain "hublio/internal/orchestration/domain"

	"github.com/google/uuid"
)

// OrchestrationEventBridge adapts the Orchestration BC's own EventPublisher port to the
// Events BC Publisher, translating orchestrationdomain.Event (Name/AggregateID/Payload) into
// eventsapp.PublishInput. This keeps Orchestration's Domain/Application free of any import
// on the Events BC (AGENTS.md package boundaries); only this Infrastructure-level bridge
// knows about both sides, mirroring the existing ConnectionGateway/ConnectorGateway
// composition pattern in internal/orchestration/infrastructure.
type OrchestrationEventBridge struct {
	publisher eventsapp.Publisher
}

func NewOrchestrationEventBridge(publisher eventsapp.Publisher) *OrchestrationEventBridge {
	return &OrchestrationEventBridge{publisher: publisher}
}

var _ orchestrationapp.EventPublisher = (*OrchestrationEventBridge)(nil)

// Publish maps every Orchestration domain Event to a runtime PlatformEvent and publishes
// them together. Intent/Execution events are always "runtime" facts per the Architecture
// Freeze (Intent -> Execution -> Execution Step is the Runtime Model).
func (b *OrchestrationEventBridge) Publish(ctx context.Context, events ...orchestrationdomain.Event) error {
	if len(events) == 0 {
		return nil
	}
	inputs := make([]eventsapp.PublishInput, 0, len(events))
	for _, e := range events {
		inputs = append(inputs, mapOrchestrationEvent(e))
	}
	return b.publisher.Publish(ctx, inputs...)
}

func mapOrchestrationEvent(e orchestrationdomain.Event) eventsapp.PublishInput {
	aggregateType := string(eventsdomain.AggregateTypeExecution)
	var executionID *uuid.UUID
	if strings.HasPrefix(e.Name, "Intent") {
		aggregateType = string(eventsdomain.AggregateTypeIntent)
	} else {
		id := e.AggregateID
		executionID = &id
	}

	return eventsapp.PublishInput{
		OrganizationID: uuidFromPayload(e.Payload, "organization_id"),
		WorkspaceID:    uuidFromPayload(e.Payload, "workspace_id"),
		AggregateType:  aggregateType,
		AggregateID:    e.AggregateID,
		ExecutionID:    executionID,
		Category:       string(eventsdomain.CategoryRuntime),
		EventName:      e.Name,
		CorrelationID:  stringFromPayload(e.Payload, "correlation_id"),
		Payload:        e.Payload,
		PublishedBy:    "orchestration",
	}
}

// OrchestrationAuditBridge adapts Orchestration's AuditRecorder port to the Events BC
// Auditor, filling tenant/request context from requestctx (available on every HTTP request;
// see internal/platform/middleware).
type OrchestrationAuditBridge struct {
	auditor eventsapp.Auditor
}

func NewOrchestrationAuditBridge(auditor eventsapp.Auditor) *OrchestrationAuditBridge {
	return &OrchestrationAuditBridge{auditor: auditor}
}

var _ orchestrationapp.AuditRecorder = (*OrchestrationAuditBridge)(nil)

func (b *OrchestrationAuditBridge) Record(ctx context.Context, rec orchestrationapp.AuditEvent) error {
	return b.auditor.Record(ctx, auditRecordFromContext(ctx, rec.ActorType, rec.ActorID, rec.Action, rec.ResourceType, rec.ResourceID, rec.Metadata))
}
