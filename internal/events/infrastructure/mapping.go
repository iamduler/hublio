package infrastructure

import (
	"context"

	eventsapp "hublio/internal/events/application"
	"hublio/internal/platform/requestctx"

	"github.com/google/uuid"
)

// uuidFromPayload extracts an optional uuid.UUID string field from a domain event's
// (loosely typed) Payload map. Returns nil when the key is missing, not a string, or not a
// valid UUID — bridges must degrade gracefully rather than fail Publish.
func uuidFromPayload(payload map[string]any, key string) *uuid.UUID {
	if payload == nil {
		return nil
	}
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	s, ok := raw.(string)
	if !ok || s == "" {
		return nil
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &parsed
}

// stringFromPayload extracts an optional string field from a domain event's Payload map.
func stringFromPayload(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func uuidPtrOrSelf(preferred *uuid.UUID, fallback uuid.UUID) uuid.UUID {
	if preferred != nil && *preferred != uuid.Nil {
		return *preferred
	}
	return fallback
}

func uuidPtrOrNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func parseUUIDOrNil(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &parsed
}

// auditRecordFromContext builds a full eventsapp.AuditRecord out of a BC-local, minimal
// audit call plus whatever tenant/request context is available (organization_id,
// workspace_id, correlation_id, request_id, ip, user_agent all flow through
// internal/platform/requestctx, set by middleware on every HTTP request).
func auditRecordFromContext(
	ctx context.Context,
	actorType string,
	actorID uuid.UUID,
	action, resourceType string,
	resourceID uuid.UUID,
	metadata map[string]any,
) eventsapp.AuditRecord {
	return eventsapp.AuditRecord{
		OrganizationID: parseUUIDOrNil(requestctx.OrganizationID(ctx)),
		WorkspaceID:    parseUUIDOrNil(requestctx.WorkspaceID(ctx)),
		ActorType:      actorType,
		ActorID:        uuidPtrOrNil(actorID),
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     uuidPtrOrNil(resourceID),
		RequestID:      requestctx.RequestID(ctx),
		CorrelationID:  requestctx.CorrelationID(ctx),
		IP:             requestctx.IP(ctx),
		UserAgent:      requestctx.UserAgent(ctx),
		Metadata:       metadata,
	}
}
