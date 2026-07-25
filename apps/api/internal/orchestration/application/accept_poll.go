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

type AcceptPollInput struct {
	SyncRouteID  uuid.UUID
	ResourceType string
}

type AcceptPollResult struct {
	SyncRouteID       uuid.UUID
	ResourceType      string
	Accepted          int
	SkippedFilter     int
	Replayed          int
	NextCursor        map[string]any
	WatermarkAdvanced bool
	// Jobs must be enqueued by the caller AFTER the surrounding transaction commits.
	Jobs []*ExecutionJob
	// Results holds per-record SubmitIntent outcomes (accepted / replayed).
	Results []*SubmitIntentResult
}

// AcceptPoll pulls records from the SyncRoute source Connection since the Postgres watermark,
// filters each record, submits Intents (fan-out), then advances the watermark after a successful
// pull+accept cycle (docs/30 §7.2). Destination Execution failures do not roll back the watermark.
func (s *Services) AcceptPoll(ctx context.Context, in AcceptPollInput) (*AcceptPollResult, error) {
	if s.SyncRoutes == nil {
		return nil, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}
	if s.Connections == nil || s.Connectors == nil {
		return nil, apperr.New("connector gateway not configured", apperr.ErrCodeInternal)
	}

	resourceType := strings.TrimSpace(strings.ToLower(in.ResourceType))
	if resourceType == "" {
		return nil, apperr.New("resource_type is required", apperr.ErrCodeBadRequest)
	}

	route, err := s.SyncRoutes.ResolvePoll(ctx, ResolvePollInput{
		SyncRouteID:  in.SyncRouteID,
		ResourceType: resourceType,
	})
	if err != nil {
		return nil, err
	}

	wm, err := s.SyncRoutes.LoadWatermark(ctx, route.SyncRouteID, resourceType)
	if err != nil {
		return nil, err
	}

	resolved, err := s.Connections.ResolveForIntent(ctx, route.WorkspaceID, route.SourceConnectionID)
	if err != nil {
		return nil, err
	}

	listResp, err := s.Connectors.Invoke(ctx, resolved.ConnectorCode, InvokeRequest{
		ConnectionID: resolved.ConnectionID,
		Capability:   route.ListCapability,
		Config:       resolved.Config,
		Secret:       resolved.Secret,
		Payload: map[string]any{
			"cursor":        wm.Cursor,
			"resource_type": resourceType,
			"page_size":     50,
		},
	})
	if err != nil {
		return nil, err
	}

	records := pollRecordsFromPayload(listResp.Payload)
	nextCursor := pollNextCursor(listResp.Payload, wm.Cursor, records)

	out := &AcceptPollResult{
		SyncRouteID:  route.SyncRouteID,
		ResourceType: resourceType,
		NextCursor:   nextCursor,
	}

	for _, record := range records {
		ok, ferr := matchPollFilter(s, route.Filter, record)
		if ferr != nil {
			return nil, apperr.Wrap(ferr, "invalid filter", apperr.ErrCodeBadRequest)
		}
		if !ok {
			out.SkippedFilter++
			continue
		}

		idempotencyKey := derivePollIdempotencyKey(
			route.WorkspaceID,
			route.SyncRouteID,
			resourceType,
			record,
			route.IdempotencyRule,
		)
		correlationID := fmt.Sprintf("poll-%s-%s", route.SyncRouteID.String(), resourceType)

		result, serr := s.SubmitIntent(ctx, SubmitIntentInput{
			OrganizationID: route.OrganizationID,
			WorkspaceID:    route.WorkspaceID,
			ConnectionID:   route.SourceConnectionID,
			Capability:     route.Capability,
			Payload:        record,
			CorrelationID:  correlationID,
			IdempotencyKey: idempotencyKey,
			SyncRouteID:    route.SyncRouteID,
			FanOutGroups:   route.FanOutGroups,
			FanOutReverse:  route.FanOutReverse,
		})
		if serr != nil {
			return nil, serr
		}
		out.Results = append(out.Results, result)
		if result.Replayed {
			out.Replayed++
		} else if result.Intent != nil && result.Execution != nil {
			out.Accepted++
		}
		for _, job := range result.Jobs {
			if job != nil {
				out.Jobs = append(out.Jobs, job)
			}
		}
	}

	// Advance watermark after successful pull (even when the page is empty / all filtered),
	// so the next tick does not re-scan the same window. Postgres remains source of truth.
	if err := s.SyncRoutes.AdvanceWatermark(ctx, route.SyncRouteID, resourceType, nextCursor); err != nil {
		return nil, err
	}
	out.WatermarkAdvanced = true
	return out, nil
}

// EnqueueDuePolls lists SyncRoutes that are due and returns poll jobs for the worker ticker.
// Callers enqueue after this returns (no DB writes here beyond reads).
func (s *Services) EnqueueDuePolls(ctx context.Context) ([]PollJob, error) {
	if s.SyncRoutes == nil {
		return nil, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}
	targets, err := s.SyncRoutes.ListDuePollTargets(ctx, s.clock().Now())
	if err != nil {
		return nil, err
	}
	jobs := make([]PollJob, 0, len(targets))
	for _, t := range targets {
		jobs = append(jobs, PollJob{
			SyncRouteID:  t.SyncRouteID,
			WorkspaceID:  t.WorkspaceID,
			ResourceType: t.ResourceType,
		})
	}
	return jobs, nil
}

// PollJob is the Platform work-queue payload for integration.poll_sync_route.
type PollJob struct {
	SyncRouteID  uuid.UUID
	WorkspaceID  uuid.UUID
	ResourceType string
}

func pollRecordsFromPayload(payload map[string]any) []map[string]any {
	if payload == nil {
		return nil
	}
	raw, ok := payload["records"]
	if !ok {
		return nil
	}
	switch t := raw.(type) {
	case []map[string]any:
		return t
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func pollNextCursor(payload map[string]any, previous map[string]any, records []map[string]any) map[string]any {
	if payload != nil {
		if next, ok := payload["next_cursor"].(map[string]any); ok && next != nil {
			return next
		}
	}
	if len(records) == 0 {
		if previous != nil {
			return previous
		}
		return map[string]any{}
	}
	last := records[len(records)-1]
	next := map[string]any{}
	if previous != nil {
		for k, v := range previous {
			next[k] = v
		}
	}
	for _, key := range []string{"updated_at", "last_updated_at"} {
		if v, ok := last[key]; ok && v != nil && fmt.Sprint(v) != "" {
			next["last_updated_at"] = fmt.Sprint(v)
			break
		}
	}
	for _, key := range []string{"id", "record_id", "invoice_number"} {
		if v, ok := last[key]; ok && v != nil && fmt.Sprint(v) != "" {
			next["last_record_id"] = fmt.Sprint(v)
			break
		}
	}
	return next
}

func matchPollFilter(s *Services, filter map[string]any, payload map[string]any) (bool, error) {
	if len(filter) == 0 {
		return true, nil
	}
	return s.SyncRoutes.MatchRouteFilter(filter, payload)
}

func derivePollIdempotencyKey(
	workspaceID, syncRouteID uuid.UUID,
	resourceType string,
	payload map[string]any,
	rule map[string]any,
) string {
	businessKey := webhookBusinessKey(payload, rule)
	raw := fmt.Sprintf("%s|%s|%s|poll|%s", workspaceID.String(), syncRouteID.String(), resourceType, businessKey)
	sum := sha256.Sum256([]byte(raw))
	return "poll_" + hex.EncodeToString(sum[:])
}
