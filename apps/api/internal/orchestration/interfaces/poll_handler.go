package interfaces

import (
	"context"
	"net/http"

	"hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/persistence"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type triggerPollRequest struct {
	ResourceType string `json:"resource_type" binding:"required"`
}

// triggerPoll runs AcceptPoll for ops/smoke (API Key). Same path as the worker poll job.
func (h *Handler) triggerPoll(c *gin.Context) {
	syncRouteID, err := uuid.Parse(c.Param("syncRouteId"))
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid syncRouteId", apperr.ErrCodeBadRequest))
		return
	}
	var req triggerPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ResponseError(c, apperr.New(err.Error(), apperr.ErrCodeBadRequest))
		return
	}

	var result *application.AcceptPollResult
	err = persistence.WithinTransaction(c.Request.Context(), h.pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = h.svc.AcceptPoll(ctx, application.AcceptPollInput{
			SyncRouteID:  syncRouteID,
			ResourceType: req.ResourceType,
		})
		return innerErr
	})
	if err != nil {
		httpx.ResponseError(c, err)
		return
	}

	for _, intentResult := range result.Results {
		if intentResult == nil || intentResult.Intent == nil {
			continue
		}
		events := intentResult.Intent.PullEvents()
		for _, exec := range intentResult.Executions {
			events = append(events, exec.PullEvents()...)
		}
		events = application.EnrichEvents(
			events,
			intentResult.Intent.OrganizationID(),
			intentResult.Intent.WorkspaceID(),
			intentResult.Intent.CorrelationID(),
		)
		h.svc.PublishAfterCommit(c.Request.Context(), events...)
	}

	for _, job := range result.Jobs {
		if job == nil || h.svc.Jobs == nil {
			continue
		}
		if err := h.svc.Jobs.EnqueueExecution(c.Request.Context(), *job); err != nil {
			httpx.ResponseError(c, apperr.Wrap(err, "poll accepted but failed to enqueue execution", apperr.ErrCodeInternal))
			return
		}
	}

	h.svc.RecordAudit(c.Request.Context(), application.AuditEvent{
		ActorType:    "api_key",
		Action:       "sync_route.poll",
		ResourceType: "sync_route",
		ResourceID:   syncRouteID,
		Metadata: map[string]any{
			"resource_type":       req.ResourceType,
			"accepted":            result.Accepted,
			"replayed":            result.Replayed,
			"skipped_filter":      result.SkippedFilter,
			"watermark_advanced":  result.WatermarkAdvanced,
		},
	})

	httpx.ResponseSuccess(c, http.StatusOK, "poll completed", gin.H{
		"sync_route_id":       result.SyncRouteID.String(),
		"resource_type":       result.ResourceType,
		"accepted":            result.Accepted,
		"replayed":            result.Replayed,
		"skipped_filter":      result.SkippedFilter,
		"watermark_advanced":  result.WatermarkAdvanced,
		"next_cursor":         result.NextCursor,
		"jobs_enqueued":       len(result.Jobs),
	})
}
