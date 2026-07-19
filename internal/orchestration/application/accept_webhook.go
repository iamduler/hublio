package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

type AcceptWebhookInput struct {
	SyncRouteID    uuid.UUID
	SecretHeader   string
	ResourceType   string
	Operation      string
	Payload        map[string]any
	IdempotencyKey string
	CorrelationID  string
}

// AcceptWebhook validates an inbound SyncRoute webhook (via SyncRouteGateway) then submits
// a Business Intent. Clients never create Executions directly. Fan-out to multiple Executions
// is deferred — v1 uses the primary activity step only.
func (s *Services) AcceptWebhook(ctx context.Context, in AcceptWebhookInput) (*SubmitIntentResult, error) {
	if s.SyncRoutes == nil {
		return nil, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}
	resourceType := strings.TrimSpace(strings.ToLower(in.ResourceType))
	if resourceType == "" {
		return nil, apperr.New("resource_type is required", apperr.ErrCodeBadRequest)
	}
	if len(in.Payload) == 0 {
		return nil, apperr.New("payload is required", apperr.ErrCodeBadRequest)
	}

	route, err := s.SyncRoutes.ResolveWebhook(ctx, ResolveWebhookInput{
		SyncRouteID:  in.SyncRouteID,
		SecretHeader: in.SecretHeader,
		ResourceType: resourceType,
		Payload:      in.Payload,
	})
	if err != nil {
		return nil, err
	}

	operation := strings.TrimSpace(strings.ToLower(in.Operation))
	if operation == "" {
		operation = "upsert"
	}

	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = deriveWebhookIdempotencyKey(
			route.WorkspaceID,
			route.SyncRouteID,
			resourceType,
			operation,
			in.Payload,
			route.IdempotencyRule,
		)
	}

	correlationID := strings.TrimSpace(in.CorrelationID)
	if correlationID == "" {
		correlationID = fmt.Sprintf("wh-%s", route.SyncRouteID.String())
	}

	return s.SubmitIntent(ctx, SubmitIntentInput{
		OrganizationID: route.OrganizationID,
		WorkspaceID:    route.WorkspaceID,
		ConnectionID:   route.ConnectionID,
		Capability:     route.Capability,
		Payload:        in.Payload,
		CorrelationID:  correlationID,
		IdempotencyKey: idempotencyKey,
	})
}

func deriveWebhookIdempotencyKey(
	workspaceID, syncRouteID uuid.UUID,
	resourceType, operation string,
	payload map[string]any,
	rule map[string]any,
) string {
	businessKey := webhookBusinessKey(payload, rule)
	raw := fmt.Sprintf("%s|%s|%s|%s|%s", workspaceID.String(), syncRouteID.String(), resourceType, operation, businessKey)
	sum := sha256.Sum256([]byte(raw))
	return "wh_" + hex.EncodeToString(sum[:])
}

func webhookBusinessKey(payload map[string]any, rule map[string]any) string {
	if rule != nil {
		if fields, ok := rule["fields"].([]any); ok && len(fields) > 0 {
			parts := make([]string, 0, len(fields))
			for _, f := range fields {
				path, _ := f.(string)
				if path == "" {
					continue
				}
				parts = append(parts, fmt.Sprint(payloadValueAt(payload, path)))
			}
			if len(parts) > 0 {
				return strings.Join(parts, ":")
			}
		}
	}
	for _, key := range []string{"id", "record_id", "invoice_number", "bill_id", "order_id"} {
		if v, ok := payload[key]; ok && v != nil && fmt.Sprint(v) != "" {
			return fmt.Sprint(v)
		}
	}
	sum := sha256.Sum256([]byte(fmt.Sprint(payload)))
	return hex.EncodeToString(sum[:16])
}

func payloadValueAt(payload map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var cur any = payload
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = m[p]
		if !ok {
			return nil
		}
	}
	return cur
}
