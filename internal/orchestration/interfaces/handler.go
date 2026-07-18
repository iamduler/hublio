package interfaces

import (
	"context"
	"net/http"

	"hublio/internal/orchestration/application"
	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler serves the Platform (machine) Orchestration API: Intents and Executions.
// Every route is Workspace-scoped API-Key auth (simplest option that meets the Phase D
// exit criteria); a JWT + workspace-membership variant is deferred (see checklist).
type Handler struct {
	svc  *application.Services
	pool *pgxpool.Pool
}

func NewHandler(svc *application.Services, pool *pgxpool.Pool) *Handler {
	return &Handler{svc: svc, pool: pool}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, apiKeyAuth gin.HandlerFunc) {
	machine := api.Group("")
	machine.Use(apiKeyAuth)
	{
		machine.POST("/intents", h.submitIntent)
		machine.GET("/intents/:intentId", h.getIntent)
		machine.GET("/executions/:executionId", h.getExecution)
		machine.POST("/executions/:executionId/cancel", h.cancelExecution)
		machine.POST("/executions/:executionId/retry", h.retryExecution)
	}
}

type submitIntentRequest struct {
	ConnectionID  string         `json:"connection_id" binding:"required"`
	Capability    string         `json:"capability" binding:"required"`
	Payload       map[string]any `json:"payload"`
	CorrelationID string         `json:"correlation_id"`
}

func (h *Handler) submitIntent(c *gin.Context) {
	orgID, ok := organizationIDFromPrincipal(c)
	if !ok {
		return
	}
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		httpx.ResponseError(c, apperr.New("missing Idempotency-Key header", apperr.ErrCodeBadRequest))
		return
	}

	var req submitIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	connectionID, err := uuid.Parse(req.ConnectionID)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid connection_id", apperr.ErrCodeBadRequest))
		return
	}

	correlationID := req.CorrelationID
	if correlationID == "" {
		correlationID = requestctx.CorrelationID(c.Request.Context())
	}

	var result *application.SubmitIntentResult
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.SubmitIntent(ctx, application.SubmitIntentInput{
			OrganizationID: orgID,
			WorkspaceID:    workspaceID,
			ConnectionID:   connectionID,
			Capability:     req.Capability,
			Payload:        req.Payload,
			CorrelationID:  correlationID,
			IdempotencyKey: idempotencyKey,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	events := result.Intent.PullEvents()
	if result.Execution != nil {
		events = append(events, result.Execution.PullEvents()...)
	}
	h.svc.PublishAfterCommit(c.Request.Context(), events...)

	if result.Job != nil && h.svc.Jobs != nil {
		if err := h.svc.Jobs.EnqueueExecution(c.Request.Context(), *result.Job); err != nil {
			httpx.ResponseError(c, apperr.Wrap(err, "intent accepted but failed to enqueue execution", apperr.ErrCodeInternal))
			return
		}
	}

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	httpx.ResponseSuccess(c, status, "intent submitted", gin.H{
		"intent":    intentDTO(result.Intent),
		"execution": executionDTOPtr(result.Execution),
	})
}

func (h *Handler) getIntent(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	intentID, ok := parseUUIDParam(c, "intentId")
	if !ok {
		return
	}
	intent, err := h.svc.GetIntent(c.Request.Context(), workspaceID, intentID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "intent", intentDTO(intent))
}

func (h *Handler) getExecution(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	executionID, ok := parseUUIDParam(c, "executionId")
	if !ok {
		return
	}
	execution, _, err := h.svc.GetExecution(c.Request.Context(), workspaceID, executionID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "execution", executionDTO(execution))
}

func (h *Handler) cancelExecution(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	executionID, ok := parseUUIDParam(c, "executionId")
	if !ok {
		return
	}
	var execution *domain.Execution
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		execution, innerErr = h.svc.CancelExecution(ctx, workspaceID, executionID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), execution.PullEvents()...)
	httpx.ResponseSuccess(c, http.StatusOK, "execution cancelled", executionDTO(execution))
}

func (h *Handler) retryExecution(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	executionID, ok := parseUUIDParam(c, "executionId")
	if !ok {
		return
	}
	var result *application.RetryExecutionResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.RetryExecution(ctx, workspaceID, executionID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	h.svc.PublishAfterCommit(c.Request.Context(), result.Execution.PullEvents()...)
	if result.Job != nil && h.svc.Jobs != nil {
		if err := h.svc.Jobs.EnqueueExecution(c.Request.Context(), *result.Job); err != nil {
			httpx.ResponseError(c, apperr.Wrap(err, "retry scheduled but failed to enqueue execution", apperr.ErrCodeInternal))
			return
		}
	}
	httpx.ResponseSuccess(c, http.StatusOK, "execution retry scheduled", executionDTO(result.Execution))
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid "+name, apperr.ErrCodeBadRequest))
		return uuid.Nil, false
	}
	return id, true
}

func workspaceIDFromPrincipal(c *gin.Context) (uuid.UUID, bool) {
	raw, _ := c.Get("workspace_id")
	s, _ := raw.(string)
	id, err := uuid.Parse(s)
	if err != nil {
		httpx.ResponseError(c, apperr.New("unauthorized: missing workspace scope", apperr.ErrCodeUnauthorized))
		return uuid.Nil, false
	}
	return id, true
}

func organizationIDFromPrincipal(c *gin.Context) (uuid.UUID, bool) {
	raw, _ := c.Get("organization_id")
	s, _ := raw.(string)
	id, err := uuid.Parse(s)
	if err != nil {
		httpx.ResponseError(c, apperr.New("unauthorized: missing organization scope", apperr.ErrCodeUnauthorized))
		return uuid.Nil, false
	}
	return id, true
}

func intentDTO(intent *domain.Intent) gin.H {
	return gin.H{
		"id":              intent.ID().String(),
		"organization_id": intent.OrganizationID().String(),
		"workspace_id":    intent.WorkspaceID().String(),
		"connection_id":   intent.ConnectionID().String(),
		"capability":      intent.Capability(),
		"payload":         intent.Payload(),
		"status":          string(intent.Status()),
		"correlation_id":  intent.CorrelationID(),
		"idempotency_key": intent.IdempotencyKey(),
		"submitted_at":    intent.SubmittedAt(),
		"created_at":      intent.CreatedAt(),
	}
}

func executionDTOPtr(execution *domain.Execution) any {
	if execution == nil {
		return nil
	}
	return executionDTO(execution)
}

func executionDTO(execution *domain.Execution) gin.H {
	var result any
	if execution.Result() != nil {
		result = string(*execution.Result())
	}

	steps := make([]gin.H, 0, len(execution.Steps()))
	for _, step := range execution.Steps() {
		steps = append(steps, stepDTO(step))
	}
	timeline := make([]gin.H, 0, len(execution.Timeline()))
	for _, entry := range execution.Timeline() {
		timeline = append(timeline, gin.H{
			"event":      entry.Event(),
			"message":    entry.Message(),
			"metadata":   entry.Metadata(),
			"created_at": entry.CreatedAt(),
		})
	}

	return gin.H{
		"id":              execution.ID().String(),
		"intent_id":       execution.IntentID().String(),
		"status":          string(execution.Status()),
		"result":          result,
		"retry_attempt":   execution.RetryAttempt(),
		"current_step_no": execution.CurrentStepNo(),
		"context":         execution.Context(),
		"failure_reason":  execution.FailureReason(),
		"started_at":      execution.StartedAt(),
		"completed_at":    execution.CompletedAt(),
		"created_at":      execution.CreatedAt(),
		"steps":           steps,
		"timeline":        timeline,
	}
}

func stepDTO(step *domain.ExecutionStep) gin.H {
	return gin.H{
		"step_no":       step.StepNo(),
		"step_type":     string(step.StepType()),
		"status":        string(step.Status()),
		"retry_attempt": step.RetryAttempt(),
		"duration_ms":   step.DurationMs(),
		"error_code":    step.ErrorCode(),
		"error_message": step.ErrorMessage(),
		"started_at":    step.StartedAt(),
		"completed_at":  step.CompletedAt(),
	}
}
