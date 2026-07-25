package interfaces

import (
	"net/http"
	"strconv"

	"hublio/internal/events/application"
	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves the read-only Platform Events API (F3 Observability):
// GET /api/v1/events — persisted PlatformEvents, Workspace-scoped via API-Key auth.
type Handler struct {
	svc *application.Services
}

func NewHandler(svc *application.Services) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, apiKeyAuth gin.HandlerFunc) {
	machine := api.Group("")
	machine.Use(apiKeyAuth)
	{
		machine.GET("/events", h.listEvents)
	}
}

func (h *Handler) listEvents(c *gin.Context) {
	workspaceID, ok := workspaceIDFromPrincipal(c)
	if !ok {
		return
	}

	var executionID *uuid.UUID
	if raw := c.Query("execution_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			httpx.ResponseError(c, apperr.New("invalid execution_id", apperr.ErrCodeBadRequest))
			return
		}
		executionID = &id
	}

	limit := int32(50)
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httpx.ResponseError(c, apperr.New("invalid limit", apperr.ErrCodeBadRequest))
			return
		}
		limit = int32(parsed)
	}

	events, err := h.svc.ListEvents(c.Request.Context(), workspaceID, executionID, limit)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	out := make([]gin.H, 0, len(events))
	for _, e := range events {
		out = append(out, eventDTO(e))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "events", out)
}

func eventDTO(e *domain.PlatformEvent) gin.H {
	var orgID, wsID, execID any
	if v := e.OrganizationID(); v != nil {
		orgID = v.String()
	}
	if v := e.WorkspaceID(); v != nil {
		wsID = v.String()
	}
	if v := e.ExecutionID(); v != nil {
		execID = v.String()
	}
	return gin.H{
		"id":              e.ID().String(),
		"organization_id": orgID,
		"workspace_id":    wsID,
		"aggregate_type":  string(e.AggregateType()),
		"aggregate_id":    e.AggregateID().String(),
		"execution_id":    execID,
		"category":        string(e.Category()),
		"event_name":      e.EventName(),
		"correlation_id":  e.CorrelationID(),
		"payload":         e.Payload(),
		"metadata":        e.Metadata(),
		"published_by":    e.PublishedBy(),
		"created_at":      e.CreatedAt(),
	}
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
