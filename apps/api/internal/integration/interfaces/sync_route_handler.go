package interfaces

import (
	"context"
	"net/http"

	"hublio/internal/integration/application"
	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type activityStepRequest struct {
	DestinationConnectionID string `json:"destination_connection_id" binding:"required"`
	Capability              string `json:"capability" binding:"required"`
	MappingKey              string `json:"mapping_key"`
}

type activityGroupRequest struct {
	Mode  string                `json:"group_mode" binding:"required"`
	Steps []activityStepRequest `json:"steps" binding:"required,min=1"`
}

type reverseRequest struct {
	Capability string `json:"capability" binding:"required"`
	On         string `json:"on"`
}

type createSyncRouteRequest struct {
	SourceConnectionID string                 `json:"source_connection_id" binding:"required"`
	Name               string                 `json:"name" binding:"required"`
	TriggerType        string                 `json:"trigger_type" binding:"required"`
	ResourceTypes      []string               `json:"resource_types" binding:"required,min=1"`
	Schedule           map[string]any         `json:"schedule"`
	Filter             map[string]any         `json:"filter"`
	IdempotencyRule    map[string]any         `json:"idempotency_rule"`
	Activities         []activityGroupRequest `json:"activities" binding:"required,min=1"`
	Reverse            *reverseRequest        `json:"reverse"`
	RetryPolicy        map[string]any         `json:"retry_policy"`
}

type updateSyncRouteRequest struct {
	Name               *string                `json:"name"`
	SourceConnectionID *string                `json:"source_connection_id"`
	TriggerType        *string                `json:"trigger_type"`
	ResourceTypes      []string               `json:"resource_types"`
	Schedule           map[string]any         `json:"schedule"`
	Filter             map[string]any         `json:"filter"`
	IdempotencyRule    map[string]any         `json:"idempotency_rule"`
	Activities         []activityGroupRequest `json:"activities"`
	Reverse            *reverseRequest        `json:"reverse"`
	ClearReverse       bool                   `json:"clear_reverse"`
	RetryPolicy        map[string]any         `json:"retry_policy"`
}

type upsertWatermarkRequest struct {
	Cursor map[string]any `json:"cursor" binding:"required"`
}

func (h *Handler) listSyncRoutes(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	list, err := h.svc.ListSyncRoutes(c.Request.Context(), workspaceID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, route := range list {
		out = append(out, syncRouteDTO(route, ""))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "ok", out)
}

func (h *Handler) createSyncRoute(c *gin.Context) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return
	}
	actorID, ok := h.requireWorkspaceMember(c, workspaceID)
	if !ok {
		return
	}
	var req createSyncRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	sourceID, err := uuid.Parse(req.SourceConnectionID)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid source_connection_id", apperr.ErrCodeBadRequest))
		return
	}
	activities, err := mapActivityGroupRequests(req.Activities)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	var result *application.CreateSyncRouteResult
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.CreateSyncRoute(ctx, application.CreateSyncRouteInput{
			WorkspaceID:        workspaceID,
			SourceConnectionID: sourceID,
			Name:               req.Name,
			Trigger:            req.TriggerType,
			ResourceTypes:      req.ResourceTypes,
			Schedule:           req.Schedule,
			Filter:             req.Filter,
			IdempotencyRule:    req.IdempotencyRule,
			Activities:         activities,
			Reverse:            mapReverseRequest(req.Reverse),
			RetryPolicy:        req.RetryPolicy,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	events := application.EnrichEvents(result.Route.PullEvents(), workspaceID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "sync_route.create",
		ResourceType: "sync_route",
		ResourceID:   result.Route.ID(),
		Metadata: map[string]any{
			"workspace_id": workspaceID.String(),
			"name":         result.Route.Name(),
			"trigger_type": string(result.Route.Trigger()),
		},
	})
	httpx.ResponseSuccess(c, http.StatusCreated, "sync route created", syncRouteDTO(result.Route, result.WebhookSecretPlaintext))
}

func (h *Handler) getSyncRoute(c *gin.Context) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	route, err := h.svc.GetSyncRoute(c.Request.Context(), workspaceID, syncRouteID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "ok", syncRouteDTO(route, ""))
}

func (h *Handler) updateSyncRoute(c *gin.Context) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	var req updateSyncRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	in := application.UpdateSyncRouteInput{
		WorkspaceID:     workspaceID,
		SyncRouteID:     syncRouteID,
		Name:            req.Name,
		Trigger:         req.TriggerType,
		ResourceTypes:   req.ResourceTypes,
		Schedule:        req.Schedule,
		Filter:          req.Filter,
		IdempotencyRule: req.IdempotencyRule,
		ClearReverse:    req.ClearReverse,
		RetryPolicy:     req.RetryPolicy,
		Reverse:         mapReverseRequest(req.Reverse),
	}
	if req.SourceConnectionID != nil {
		id, err := uuid.Parse(*req.SourceConnectionID)
		if err != nil {
			httpx.ResponseError(c, apperr.New("invalid source_connection_id", apperr.ErrCodeBadRequest))
			return
		}
		in.SourceConnectionID = &id
	}
	if req.Activities != nil {
		activities, err := mapActivityGroupRequests(req.Activities)
		if err != nil {
			httpx.ResponseError(c, err)
			return
		}
		in.Activities = activities
	}

	var route *domain.SyncRoute
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		route, innerErr = h.svc.UpdateSyncRoute(ctx, in)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	events := application.EnrichEvents(route.PullEvents(), workspaceID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	httpx.ResponseSuccess(c, http.StatusOK, "sync route updated", syncRouteDTO(route, ""))
}

func (h *Handler) enableSyncRoute(c *gin.Context) {
	h.mutateSyncRoute(c, h.svc.EnableSyncRoute, "sync route enabled")
}

func (h *Handler) disableSyncRoute(c *gin.Context) {
	h.mutateSyncRoute(c, h.svc.DisableSyncRoute, "sync route disabled")
}

func (h *Handler) deleteSyncRoute(c *gin.Context) {
	h.mutateSyncRoute(c, h.svc.DeleteSyncRoute, "sync route deleted")
}

func (h *Handler) mutateSyncRoute(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) (*domain.SyncRoute, error), message string) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	var route *domain.SyncRoute
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		route, innerErr = fn(ctx, workspaceID, syncRouteID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	events := application.EnrichEvents(route.PullEvents(), workspaceID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	httpx.ResponseSuccess(c, http.StatusOK, message, syncRouteDTO(route, ""))
}

func (h *Handler) rotateSyncRouteWebhookSecret(c *gin.Context) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	actorID, ok := h.requireWorkspaceMember(c, workspaceID)
	if !ok {
		return
	}
	var result *application.RotateSyncRouteWebhookSecretResult
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.RotateSyncRouteWebhookSecret(ctx, workspaceID, syncRouteID)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	events := application.EnrichEvents(result.Route.PullEvents(), workspaceID)
	h.svc.PublishAfterCommit(c.Request.Context(), events...)
	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "user",
		ActorID:      actorID,
		Action:       "sync_route.webhook_secret.rotate",
		ResourceType: "sync_route",
		ResourceID:   result.Route.ID(),
		Metadata:     map[string]any{"workspace_id": workspaceID.String()},
	})
	httpx.ResponseSuccess(c, http.StatusOK, "webhook secret rotated", syncRouteDTO(result.Route, result.WebhookSecretPlaintext))
}

func (h *Handler) listSyncRouteWatermarks(c *gin.Context) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	list, err := h.svc.ListSyncRouteWatermarks(c.Request.Context(), workspaceID, syncRouteID)
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, wm := range list {
		out = append(out, watermarkDTO(wm))
	}
	httpx.ResponseSuccess(c, http.StatusOK, "ok", out)
}

func (h *Handler) upsertSyncRouteWatermark(c *gin.Context) {
	workspaceID, syncRouteID, ok := h.parseSyncRouteParams(c)
	if !ok {
		return
	}
	if _, ok := h.requireWorkspaceMember(c, workspaceID); !ok {
		return
	}
	resourceType := c.Param("resourceType")
	var req upsertWatermarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}
	var wm *domain.SyncRouteWatermark
	err := persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		wm, innerErr = h.svc.UpsertSyncRouteWatermark(ctx, workspaceID, syncRouteID, resourceType, req.Cursor)
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}
	httpx.ResponseSuccess(c, http.StatusOK, "watermark upserted", watermarkDTO(wm))
}

func (h *Handler) parseSyncRouteParams(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	workspaceID, ok := parseUUIDParam(c, "workspaceId")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	syncRouteID, ok := parseUUIDParam(c, "syncRouteId")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	return workspaceID, syncRouteID, true
}

func mapActivityGroupRequests(in []activityGroupRequest) ([]application.ActivityGroupInput, error) {
	out := make([]application.ActivityGroupInput, 0, len(in))
	for _, g := range in {
		steps := make([]application.ActivityStepInput, 0, len(g.Steps))
		for _, s := range g.Steps {
			destID, err := uuid.Parse(s.DestinationConnectionID)
			if err != nil {
				return nil, apperr.New("invalid destination_connection_id", apperr.ErrCodeBadRequest)
			}
			steps = append(steps, application.ActivityStepInput{
				DestinationConnectionID: destID,
				Capability:              s.Capability,
				MappingKey:              s.MappingKey,
			})
		}
		out = append(out, application.ActivityGroupInput{Mode: g.Mode, Steps: steps})
	}
	return out, nil
}

func mapReverseRequest(in *reverseRequest) *application.ReverseInput {
	if in == nil {
		return nil
	}
	return &application.ReverseInput{Capability: in.Capability, On: in.On}
}

func syncRouteDTO(route *domain.SyncRoute, webhookPlaintext string) gin.H {
	activities := make([]gin.H, 0, len(route.Activities()))
	for _, g := range route.Activities() {
		steps := make([]gin.H, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, gin.H{
				"destination_connection_id": s.DestinationConnectionID.String(),
				"capability":                s.Capability,
				"mapping_key":               s.MappingKey,
			})
		}
		activities = append(activities, gin.H{
			"group_mode": string(g.Mode),
			"steps":      steps,
		})
	}
	out := gin.H{
		"id":                   route.ID().String(),
		"workspace_id":         route.WorkspaceID().String(),
		"source_connection_id": route.SourceConnectionID().String(),
		"name":                 route.Name(),
		"status":               string(route.Status()),
		"trigger_type":         string(route.Trigger()),
		"resource_types":       route.ResourceTypes(),
		"schedule":             route.Schedule(),
		"filter":               route.Filter(),
		"idempotency_rule":     route.IdempotencyRule(),
		"activities":           activities,
		"retry_policy":         route.RetryPolicy(),
		"has_webhook_secret":   route.HasWebhookSecret(),
		"created_at":           route.CreatedAt(),
		"updated_at":           route.UpdatedAt(),
	}
	if rev := route.Reverse(); rev != nil {
		out["reverse"] = gin.H{"capability": rev.Capability, "on": rev.On}
	}
	if webhookPlaintext != "" {
		out["webhook_secret"] = webhookPlaintext
	}
	return out
}

func watermarkDTO(wm *domain.SyncRouteWatermark) gin.H {
	return gin.H{
		"sync_route_id": wm.SyncRouteID.String(),
		"resource_type": wm.ResourceType,
		"cursor":        wm.Cursor,
		"updated_at":    wm.UpdatedAt,
	}
}
