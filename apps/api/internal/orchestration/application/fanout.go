package application

import (
	"context"
	"fmt"
	"time"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

// Fan-out metadata keys stored on Execution.context (no secrets).
const (
	CtxFanoutConnectionID    = "fanout_connection_id"
	CtxFanoutCapability      = "fanout_capability"
	CtxFanoutSyncRouteID     = "sync_route_id"
	CtxFanoutGroupIndex      = "fanout_group_index"
	CtxFanoutStepIndex       = "fanout_step_index"
	CtxFanoutGroupMode       = "fanout_group_mode"
	CtxFanoutNextExecution   = "fanout_next_execution_id"
	CtxFanoutSiblingIDs      = "fanout_sibling_ids"
	CtxFanoutRequireSiblings = "fanout_require_siblings"
	CtxFanoutOnFailure       = "fanout_on_failure_execution_id"
	CtxFanoutMappingKey      = "fanout_mapping_key"
	CtxFanoutIsReverse       = "fanout_is_reverse"
)

// FanOutStep is one destination activity under a SyncRoute group.
type FanOutStep struct {
	ConnectionID uuid.UUID
	Capability   string
	MappingKey   string
}

// FanOutGroup is sequential or parallel enqueue of destination Executions.
type FanOutGroup struct {
	Mode  string // "sequential" | "parallel"
	Steps []FanOutStep
}

// FanOutReverse optionally updates the SyncRoute source after destinations.
type FanOutReverse struct {
	ConnectionID uuid.UUID
	Capability   string
	On           string // success | failure | always
}

type fanoutPrepared struct {
	execution *domain.Execution
	meta      map[string]any
}

// createFanOutExecutions creates one Execution per activity step, wires sequential/parallel
// continuation (and optional reverse), queues the first wave, and returns jobs for after-commit enqueue.
func (s *Services) createFanOutExecutions(
	ctx context.Context,
	intent *domain.Intent,
	syncRouteID uuid.UUID,
	groups []FanOutGroup,
	reverse *FanOutReverse,
	now time.Time,
) ([]*domain.Execution, []*ExecutionJob, error) {
	if len(groups) == 0 {
		return nil, nil, apperr.New("fan-out requires at least one activity group", apperr.ErrCodeBadRequest)
	}

	type groupBucket struct {
		mode string
		prep []*fanoutPrepared
	}
	buckets := make([]groupBucket, 0, len(groups))

	for gi, g := range groups {
		mode := g.Mode
		if mode == "" {
			mode = "sequential"
		}
		if mode != "sequential" && mode != "parallel" {
			return nil, nil, apperr.New("activity group_mode must be sequential or parallel", apperr.ErrCodeBadRequest)
		}
		if len(g.Steps) == 0 {
			return nil, nil, apperr.New("activity group has no steps", apperr.ErrCodeBadRequest)
		}
		bucket := groupBucket{mode: mode}
		for si, step := range g.Steps {
			if step.ConnectionID == uuid.Nil || step.Capability == "" {
				return nil, nil, apperr.New("fan-out step requires connection_id and capability", apperr.ErrCodeBadRequest)
			}
			if _, err := s.Connections.ResolveForIntent(ctx, intent.WorkspaceID(), step.ConnectionID); err != nil {
				return nil, nil, err
			}
			meta := map[string]any{
				CtxFanoutConnectionID: step.ConnectionID.String(),
				CtxFanoutCapability:   step.Capability,
				CtxFanoutSyncRouteID:  syncRouteID.String(),
				CtxFanoutGroupIndex:   gi,
				CtxFanoutStepIndex:    si,
				CtxFanoutGroupMode:    mode,
			}
			if step.MappingKey != "" {
				meta[CtxFanoutMappingKey] = step.MappingKey
			}
			exec, err := s.createExecutionWithContext(ctx, intent.ID(), meta, now)
			if err != nil {
				return nil, nil, err
			}
			bucket.prep = append(bucket.prep, &fanoutPrepared{execution: exec, meta: meta})
		}
		buckets = append(buckets, bucket)
	}

	var reversePrep *fanoutPrepared
	if reverse != nil && reverse.ConnectionID != uuid.Nil && reverse.Capability != "" {
		if _, err := s.Connections.ResolveForIntent(ctx, intent.WorkspaceID(), reverse.ConnectionID); err != nil {
			return nil, nil, err
		}
		on := reverse.On
		if on == "" {
			on = "success"
		}
		meta := map[string]any{
			CtxFanoutConnectionID: reverse.ConnectionID.String(),
			CtxFanoutCapability:   reverse.Capability,
			CtxFanoutSyncRouteID:  syncRouteID.String(),
			CtxFanoutGroupIndex:   len(buckets),
			CtxFanoutStepIndex:    0,
			CtxFanoutGroupMode:    "sequential",
			CtxFanoutIsReverse:    true,
		}
		exec, err := s.createExecutionWithContext(ctx, intent.ID(), meta, now)
		if err != nil {
			return nil, nil, err
		}
		reversePrep = &fanoutPrepared{execution: exec, meta: meta}
		_ = on
	}

	// Wire sequential next pointers and parallel sibling sets; link groups.
	for gi, bucket := range buckets {
		ids := make([]string, 0, len(bucket.prep))
		for _, p := range bucket.prep {
			ids = append(ids, p.execution.ID().String())
		}
		if bucket.mode == "parallel" {
			for _, p := range bucket.prep {
				p.meta[CtxFanoutSiblingIDs] = append([]string(nil), ids...)
			}
		} else {
			for i := 0; i < len(bucket.prep)-1; i++ {
				bucket.prep[i].meta[CtxFanoutNextExecution] = bucket.prep[i+1].execution.ID().String()
			}
		}
		if gi < len(buckets)-1 {
			nextFirst := buckets[gi+1].prep[0].execution.ID().String()
			if bucket.mode == "parallel" {
				for _, p := range bucket.prep {
					p.meta[CtxFanoutNextExecution] = nextFirst
					p.meta[CtxFanoutRequireSiblings] = true
				}
			} else {
				last := bucket.prep[len(bucket.prep)-1]
				last.meta[CtxFanoutNextExecution] = nextFirst
			}
		}
	}

	if reversePrep != nil {
		on := reverse.On
		if on == "" {
			on = "success"
		}
		revID := reversePrep.execution.ID().String()
		lastBucket := buckets[len(buckets)-1]
		if on == "success" || on == "always" {
			if lastBucket.mode == "parallel" {
				for _, p := range lastBucket.prep {
					p.meta[CtxFanoutNextExecution] = revID
					p.meta[CtxFanoutRequireSiblings] = true
				}
			} else {
				last := lastBucket.prep[len(lastBucket.prep)-1]
				last.meta[CtxFanoutNextExecution] = revID
			}
		}
		if on == "failure" || on == "always" {
			for _, bucket := range buckets {
				for _, p := range bucket.prep {
					p.meta[CtxFanoutOnFailure] = revID
				}
			}
		}
	}

	var executions []*domain.Execution
	var jobs []*ExecutionJob
	for gi, bucket := range buckets {
		for _, p := range bucket.prep {
			p.execution.MergeContext(p.meta)
			enqueue := false
			if gi == 0 {
				if bucket.mode == "parallel" {
					enqueue = true
				} else if intFromAny(p.meta[CtxFanoutStepIndex]) == 0 {
					enqueue = true
				}
			}
			if enqueue {
				if err := p.execution.Queue(now); err != nil {
					return nil, nil, mapDomainErr(err)
				}
				jobs = append(jobs, executionJobFor(intent, p.execution.ID()))
			}
			if err := s.Executions.Update(ctx, p.execution); err != nil {
				return nil, nil, mapRepoErr(err)
			}
			executions = append(executions, p.execution)
		}
	}
	if reversePrep != nil {
		reversePrep.execution.MergeContext(reversePrep.meta)
		if err := s.Executions.Update(ctx, reversePrep.execution); err != nil {
			return nil, nil, mapRepoErr(err)
		}
		executions = append(executions, reversePrep.execution)
	}
	return executions, jobs, nil
}

func (s *Services) createExecutionWithContext(
	ctx context.Context,
	intentID uuid.UUID,
	initialContext map[string]any,
	now time.Time,
) (*domain.Execution, error) {
	execID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate execution id", apperr.ErrCodeInternal)
	}
	stepIDs := make([]uuid.UUID, len(domain.DefaultStepTypes()))
	for i := range stepIDs {
		stepID, err := id.NewV7()
		if err != nil {
			return nil, apperr.Wrap(err, "failed to generate execution step id", apperr.ErrCodeInternal)
		}
		stepIDs[i] = stepID
	}
	execution, err := domain.NewExecution(execID, intentID, stepIDs, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if len(initialContext) > 0 {
		execution.MergeContext(initialContext)
	}
	if err := s.Executions.Save(ctx, execution); err != nil {
		return nil, mapRepoErr(err)
	}
	return execution, nil
}

// continueFanOutAfterSuccess queues the next sequential Execution and/or the next parallel group.
func (s *Services) continueFanOutAfterSuccess(ctx context.Context, execution *domain.Execution, intent *domain.Intent) ([]*ExecutionJob, error) {
	meta := execution.Context()
	if meta == nil {
		return nil, nil
	}
	nextRaw, _ := meta[CtxFanoutNextExecution].(string)
	if nextRaw == "" {
		return nil, nil
	}
	if requireSiblings, _ := meta[CtxFanoutRequireSiblings].(bool); requireSiblings {
		ready, err := s.fanOutSiblingsSucceeded(ctx, execution, meta)
		if err != nil {
			return nil, err
		}
		if !ready {
			return nil, nil
		}
	}

	nextID, err := uuid.Parse(nextRaw)
	if err != nil {
		return nil, nil
	}
	return s.queueFanOutWave(ctx, intent, nextID, s.clock().Now())
}

// continueFanOutAfterFailure queues the optional reverse-on-failure Execution when present.
func (s *Services) continueFanOutAfterFailure(ctx context.Context, execution *domain.Execution, intent *domain.Intent) ([]*ExecutionJob, error) {
	meta := execution.Context()
	if meta == nil {
		return nil, nil
	}
	raw, _ := meta[CtxFanoutOnFailure].(string)
	if raw == "" {
		return nil, nil
	}
	nextID, err := uuid.Parse(raw)
	if err != nil {
		return nil, nil
	}
	return s.queueFanOutWave(ctx, intent, nextID, s.clock().Now())
}

func (s *Services) fanOutSiblingsSucceeded(ctx context.Context, execution *domain.Execution, meta map[string]any) (bool, error) {
	sibIDs, ok := stringSliceFromAny(meta[CtxFanoutSiblingIDs])
	if !ok || len(sibIDs) == 0 {
		return true, nil
	}
	for _, sid := range sibIDs {
		sibID, err := uuid.Parse(sid)
		if err != nil {
			continue
		}
		if sibID == execution.ID() {
			if execution.Status() != domain.ExecutionStatusSucceeded {
				return false, nil
			}
			continue
		}
		sib, err := s.Executions.FindByID(ctx, sibID)
		if err != nil {
			return false, mapRepoErr(err)
		}
		if sib.Status() != domain.ExecutionStatusSucceeded {
			return false, nil
		}
	}
	return true, nil
}

func (s *Services) queueFanOutWave(ctx context.Context, intent *domain.Intent, firstID uuid.UUID, now time.Time) ([]*ExecutionJob, error) {
	first, err := s.Executions.FindByID(ctx, firstID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if first.Status() != domain.ExecutionStatusCreated {
		return nil, nil
	}
	if err := first.Queue(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Executions.Update(ctx, first); err != nil {
		return nil, mapRepoErr(err)
	}

	jobs := []*ExecutionJob{executionJobFor(intent, first.ID())}
	mode, _ := first.Context()[CtxFanoutGroupMode].(string)
	if mode != "parallel" {
		return jobs, nil
	}
	sibIDs, _ := stringSliceFromAny(first.Context()[CtxFanoutSiblingIDs])
	for _, sid := range sibIDs {
		idParsed, err := uuid.Parse(sid)
		if err != nil || idParsed == first.ID() {
			continue
		}
		sib, err := s.Executions.FindByID(ctx, idParsed)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		if sib.Status() != domain.ExecutionStatusCreated {
			continue
		}
		if err := sib.Queue(now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.Executions.Update(ctx, sib); err != nil {
			return nil, mapRepoErr(err)
		}
		jobs = append(jobs, executionJobFor(intent, sib.ID()))
	}
	return jobs, nil
}

func executionJobFor(intent *domain.Intent, executionID uuid.UUID) *ExecutionJob {
	return &ExecutionJob{
		ExecutionID:    executionID,
		IntentID:       intent.ID(),
		OrganizationID: intent.OrganizationID(),
		WorkspaceID:    intent.WorkspaceID(),
		CorrelationID:  intent.CorrelationID(),
	}
}

func stringSliceFromAny(v any) ([]string, bool) {
	switch t := v.(type) {
	case []string:
		return t, true
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			out = append(out, fmt.Sprint(x))
		}
		return out, true
	default:
		return nil, false
	}
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return -1
	}
}

// executionTarget returns the Connection/Capability this Execution should invoke.
// Fan-out overrides Intent fields via Execution.context.
func executionTarget(intent *domain.Intent, execution *domain.Execution) (connectionID uuid.UUID, capability string) {
	connectionID = intent.ConnectionID()
	capability = intent.Capability()
	ctx := execution.Context()
	if ctx == nil {
		return
	}
	if raw, ok := ctx[CtxFanoutConnectionID].(string); ok {
		if id, err := uuid.Parse(raw); err == nil {
			connectionID = id
		}
	}
	if raw, ok := ctx[CtxFanoutCapability].(string); ok && raw != "" {
		capability = raw
	}
	return
}
