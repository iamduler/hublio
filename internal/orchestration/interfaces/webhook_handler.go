package interfaces

import (
	"context"
	"net/http"

	"hublio/internal/integration/domain"
	"hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type acceptWebhookRequest struct {
	ResourceType   string         `json:"resource_type" binding:"required"`
	Operation      string         `json:"operation"`
	Payload        map[string]any `json:"payload" binding:"required"`
	IdempotencyKey string         `json:"idempotency_key"`
	CorrelationID  string         `json:"correlation_id"`
}

// RegisterWebhookRoutes mounts public SyncRoute webhook ingress (no JWT / API key).
// Auth is X-Hublio-Webhook-Secret only (docs/30 §7.1).
func (h *Handler) RegisterWebhookRoutes(api *gin.RouterGroup) {
	api.POST("/webhooks/sync-routes/:syncRouteId", h.acceptWebhook)
}

func (h *Handler) acceptWebhook(c *gin.Context) {
	syncRouteID, err := uuid.Parse(c.Param("syncRouteId"))
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid syncRouteId", apperr.ErrCodeBadRequest))
		return
	}

	secret := c.GetHeader(domain.WebhookSecretHeader)
	if secret == "" {
		// Also accept lowercase variant some proxies normalize away — primary is documented header.
		secret = c.GetHeader("x-hublio-webhook-secret")
	}
	if secret == "" {
		httpx.ResponseError(c, apperr.New("unauthorized", apperr.ErrCodeUnauthorized))
		return
	}

	var req acceptWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	correlationID := req.CorrelationID
	if correlationID == "" {
		correlationID = requestctx.CorrelationID(c.Request.Context())
	}

	var result *application.SubmitIntentResult
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.AcceptWebhook(ctx, application.AcceptWebhookInput{
			SyncRouteID:    syncRouteID,
			SecretHeader:   secret,
			ResourceType:   req.ResourceType,
			Operation:      req.Operation,
			Payload:        req.Payload,
			IdempotencyKey: req.IdempotencyKey,
			CorrelationID:  correlationID,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	orgID := result.Intent.OrganizationID()
	workspaceID := result.Intent.WorkspaceID()
	correlationID = result.Intent.CorrelationID()

	events := result.Intent.PullEvents()
	if result.Execution != nil {
		events = append(events, result.Execution.PullEvents()...)
	}
	events = application.EnrichEvents(events, orgID, workspaceID, correlationID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)

	if result.Job != nil && h.svc.Jobs != nil {
		if err := h.svc.Jobs.EnqueueExecution(c.Request.Context(), *result.Job); err != nil {
			httpx.ResponseError(c, apperr.Wrap(err, "intent accepted but failed to enqueue execution", apperr.ErrCodeInternal))
			return
		}
	}

	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "system",
		Action:       "webhook.accept",
		ResourceType: "sync_route",
		ResourceID:   syncRouteID,
		Metadata: map[string]any{
			"workspace_id":  workspaceID.String(),
			"intent_id":     result.Intent.ID().String(),
			"resource_type": req.ResourceType,
			"replayed":      result.Replayed,
		},
	})

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	body := gin.H{
		"intent_id":     result.Intent.ID().String(),
		"status":        string(result.Intent.Status()),
		"replayed":      result.Replayed,
		"sync_route_id": syncRouteID.String(),
		"resource_type": req.ResourceType,
	}
	if result.Execution != nil {
		body["execution_id"] = result.Execution.ID().String()
	}
	httpx.ResponseSuccess(c, status, "webhook accepted", body)
}
