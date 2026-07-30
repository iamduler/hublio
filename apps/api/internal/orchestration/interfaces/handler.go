package interfaces

import (
	"context"
	"net/http"
	"strconv"

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

// Handler serves the Platform Orchestration API: Intents and Executions.
// Auth: workspace API key (X-API-KEY) or user JWT + X-Workspace-ID (membership checked).
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
		machine.GET("/intents", h.listIntents)
		machine.POST("/intents", h.submitIntent)
		machine.GET("/intents/:intentId", h.getIntent)
		machine.GET("/executions", h.listExecutions)
		machine.GET("/executions/:executionId", h.getExecution)
		machine.GET("/executions/:executionId/timeline", h.getExecutionTimeline)
		machine.POST("/executions/:executionId/cancel", h.cancelExecution)
		machine.POST("/executions/:executionId/retry", h.retryExecution)
		machine.POST("/sync-routes/:syncRouteId/poll", h.triggerPoll)
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
	for _, exec := range result.Executions {
		if result.Execution != nil && exec.ID() == result.Execution.ID() {
			continue
		}
		events = append(events, exec.PullEvents()...)
	}
	events = application.EnrichEvents(events, orgID, workspaceID, correlationID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)

	jobs := result.Jobs
	if len(jobs) == 0 && result.Job != nil {
		jobs = []*application.ExecutionJob{result.Job}
	}
	for _, job := range jobs {
		if job == nil || h.svc.Jobs == nil {
			continue
		}
		if err := h.svc.Jobs.EnqueueExecution(c.Request.Context(), *job); err != nil {
			httpx.ResponseError(c, apperr.Wrap(err, "intent accepted but failed to enqueue execution", apperr.ErrCodeInternal))
			return
		}
	}

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	resp := gin.H{
		"intent":    intentDTO(result.Intent),
		"execution": executionDTOPtr(result.Execution),
	}
	if len(result.Executions) > 1 {
		execDTOs := make([]gin.H, 0, len(result.Executions))
		for _, e := range result.Executions {
			execDTOs = append(execDTOs, executionDTO(e))
		}
		resp["executions"] = execDTOs
	}
	httpx.ResponseSuccess(c, status, "intent submitted", resp)
}

func (h *Handler) listIntents(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	status, ok := optionalStatusQuery(c, validIntentStatuses)
	if !ok {
		return
	}
	limit, ok := parseLimitQuery(c)
	if !ok {
		return
	}
	list, err := h.svc.ListIntents(c.Request.Context(), workspaceID, status, limit)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, intent := range list {
		out = append(out, intentListDTO(intent))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "intents", out)
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

func (h *Handler) listExecutions(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}
	status, ok := optionalStatusQuery(c, validExecutionStatuses)
	if !ok {
		return
	}
	limit, ok := parseLimitQuery(c)
	if !ok {
		return
	}
	list, err := h.svc.ListExecutions(c.Request.Context(), workspaceID, status, limit)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, execution := range list {
		out = append(out, executionListDTO(execution))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "executions", out)
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

// getExecutionTimeline backs GET /api/v1/executions/:executionId/timeline (F3 Observability):
// a focused view of the same execution_timelines rows already returned by getExecution,
// tenant-scoped via API-Key workspace.
func (h *Handler) getExecutionTimeline(c *gin.Context) {
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
	timeline := make([]gin.H, 0, len(execution.Timeline()))
	for _, entry := range execution.Timeline() {
		timeline = append(timeline, gin.H{
			"id":         entry.ID().String(),
			"event":      entry.Event(),
			"message":    entry.Message(),
			"metadata":   entry.Metadata(),
			"created_at": entry.CreatedAt(),
		})
	}
	httpx.ResponseSuccess(c, http.StatusOK, "execution timeline", gin.H{
		"execution_id": execution.ID().String(),
		"status":       string(execution.Status()),
		"timeline":     timeline,
	})
}

func (h *Handler) cancelExecution(c *gin.Context) {
	orgID, ok := organizationIDFromPrincipal(c)
	if !ok {
		return
	}
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
	events := application.EnrichEvents(execution.PullEvents(), orgID, workspaceID, requestctx.CorrelationID(c.Request.Context()))
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	httpx.ResponseSuccess(c, http.StatusOK, "execution cancelled", executionDTO(execution))
}

func (h *Handler) retryExecution(c *gin.Context) {
	orgID, ok := organizationIDFromPrincipal(c)
	if !ok {
		return
	}
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
	events := application.EnrichEvents(result.Execution.PullEvents(), orgID, workspaceID, requestctx.CorrelationID(c.Request.Context()))
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	apiKeyID, _ := uuid.Parse(requestctx.APIKeyID(c.Request.Context()))
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "api_key",
		ActorID:      apiKeyID,
		Action:       "execution.retry",
		ResourceType: "execution",
		ResourceID:   executionID,
	})
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

var validIntentStatuses = map[string]struct{}{
	string(domain.IntentStatusSubmitted): {},
	string(domain.IntentStatusAccepted):  {},
	string(domain.IntentStatusRejected):  {},
	string(domain.IntentStatusExpired):   {},
}

var validExecutionStatuses = map[string]struct{}{
	string(domain.ExecutionStatusCreated):    {},
	string(domain.ExecutionStatusQueued):     {},
	string(domain.ExecutionStatusRunning):    {},
	string(domain.ExecutionStatusSucceeded):  {},
	string(domain.ExecutionStatusFailed):     {},
	string(domain.ExecutionStatusCancelled):  {},
	string(domain.ExecutionStatusExpired):    {},
	string(domain.ExecutionStatusDeadLetter): {},
}

func parseLimitQuery(c *gin.Context) (int32, bool) {
	limit := int32(50)
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httpx.ResponseError(c, apperr.New("invalid limit", apperr.ErrCodeBadRequest))
			return 0, false
		}
		limit = int32(parsed)
	}
	return limit, true
}

func optionalStatusQuery(c *gin.Context, allowed map[string]struct{}) (*string, bool) {
	raw := c.Query("status")
	if raw == "" {
		return nil, true
	}
	if _, ok := allowed[raw]; !ok {
		httpx.ResponseError(c, apperr.New("invalid status", apperr.ErrCodeBadRequest))
		return nil, false
	}
	return &raw, true
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

func intentListDTO(intent *domain.Intent) gin.H {
	return gin.H{
		"id":              intent.ID().String(),
		"organization_id": intent.OrganizationID().String(),
		"workspace_id":    intent.WorkspaceID().String(),
		"connection_id":   intent.ConnectionID().String(),
		"capability":      intent.Capability(),
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

func executionListDTO(execution *domain.Execution) gin.H {
	var result any
	if execution.Result() != nil {
		result = string(*execution.Result())
	}
	return gin.H{
		"id":              execution.ID().String(),
		"intent_id":       execution.IntentID().String(),
		"status":          string(execution.Status()),
		"result":          result,
		"retry_attempt":   execution.RetryAttempt(),
		"current_step_no": execution.CurrentStepNo(),
		"failure_reason":  execution.FailureReason(),
		"started_at":      execution.StartedAt(),
		"completed_at":    execution.CompletedAt(),
		"created_at":      execution.CreatedAt(),
	}
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
