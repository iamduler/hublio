package application

import (
	"context"
	"errors"
	"time"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

// idempotencyTTL bounds how long a submitted Intent's idempotency key protects retries.
const idempotencyTTL = 24 * time.Hour

type SubmitIntentInput struct {
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	ConnectionID   uuid.UUID
	Capability     string
	Payload        map[string]any
	CorrelationID  string
	IdempotencyKey string
	// Optional SyncRoute fan-out (webhook ingress). When FanOutGroups is non-empty,
	// one Intent fans out to N Executions; ConnectionID should be the SyncRoute source.
	SyncRouteID   uuid.UUID
	FanOutGroups  []FanOutGroup
	FanOutReverse *FanOutReverse
}

type SubmitIntentResult struct {
	Intent     *domain.Intent
	Execution  *domain.Execution   // first / primary Execution (backward compatible)
	Executions []*domain.Execution // all Executions under the Intent (fan-out)
	Replayed   bool
	// Job is the first job (backward compatible). Prefer Jobs for fan-out.
	Job  *ExecutionJob
	Jobs []*ExecutionJob
}

// SubmitIntent is the single entry point for Business Intents. It resolves the target
// Connection, creates the Intent, accepts/rejects it based on payload validity, and on
// Accept creates+queues Execution(s) and returns worker jobs for after-commit enqueue.
func (s *Services) SubmitIntent(ctx context.Context, in SubmitIntentInput) (*SubmitIntentResult, error) {
	now := s.clock().Now()

	if in.IdempotencyKey != "" && s.Idempotency != nil {
		existing, err := s.Idempotency.FindByKey(ctx, in.OrganizationID, in.WorkspaceID, in.IdempotencyKey)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, mapRepoErr(err)
		}
		if err == nil && !existing.IsExpired(now) && existing.IntentID() != nil {
			return s.loadSubmittedIntent(ctx, *existing.IntentID())
		}
	}

	if _, err := s.Connections.ResolveForIntent(ctx, in.WorkspaceID, in.ConnectionID); err != nil {
		return nil, err
	}

	intentID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate intent id", apperr.ErrCodeInternal)
	}
	intent, err := domain.NewIntent(intentID, in.OrganizationID, in.WorkspaceID, in.ConnectionID, in.Capability, in.Payload, in.CorrelationID, in.IdempotencyKey, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Intents.Save(ctx, intent); err != nil {
		return nil, mapRepoErr(err)
	}

	result := &SubmitIntentResult{Intent: intent}

	if !intent.IsValid() {
		if err := intent.Reject("capability and payload are required", now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.Intents.Update(ctx, intent); err != nil {
			return nil, mapRepoErr(err)
		}
		if err := s.saveIdempotencyKey(ctx, in, intent.ID(), now); err != nil {
			return nil, err
		}
		return result, nil
	}

	if err := intent.Accept(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Intents.Update(ctx, intent); err != nil {
		return nil, mapRepoErr(err)
	}

	if len(in.FanOutGroups) > 0 {
		executions, jobs, err := s.createFanOutExecutions(ctx, intent, in.SyncRouteID, in.FanOutGroups, in.FanOutReverse, now)
		if err != nil {
			return nil, err
		}
		result.Executions = executions
		if len(executions) > 0 {
			result.Execution = executions[0]
		}
		result.Jobs = jobs
		if len(jobs) > 0 {
			result.Job = jobs[0]
		}
	} else {
		execution, err := s.createExecution(ctx, intent.ID(), now)
		if err != nil {
			return nil, err
		}
		result.Execution = execution
		result.Executions = []*domain.Execution{execution}
		result.Job = &ExecutionJob{
			ExecutionID:    execution.ID(),
			IntentID:       intent.ID(),
			OrganizationID: in.OrganizationID,
			WorkspaceID:    in.WorkspaceID,
			CorrelationID:  in.CorrelationID,
		}
		result.Jobs = []*ExecutionJob{result.Job}
	}

	if err := s.saveIdempotencyKey(ctx, in, intent.ID(), now); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Services) createExecution(ctx context.Context, intentID uuid.UUID, now time.Time) (*domain.Execution, error) {
	execution, err := s.createExecutionWithContext(ctx, intentID, nil, now)
	if err != nil {
		return nil, err
	}
	if err := execution.Queue(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Executions.Update(ctx, execution); err != nil {
		return nil, mapRepoErr(err)
	}
	return execution, nil
}

func (s *Services) saveIdempotencyKey(ctx context.Context, in SubmitIntentInput, intentID uuid.UUID, now time.Time) error {
	if in.IdempotencyKey == "" || s.Idempotency == nil {
		return nil
	}
	recID, err := id.NewV7()
	if err != nil {
		return apperr.Wrap(err, "failed to generate idempotency key id", apperr.ErrCodeInternal)
	}
	expiresAt := now.Add(idempotencyTTL)
	rec, err := domain.NewIdempotencyKey(recID, in.OrganizationID, in.WorkspaceID, in.IdempotencyKey, intentID, &expiresAt, now)
	if err != nil {
		return mapDomainErr(err)
	}
	if err := s.Idempotency.Save(ctx, rec); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Services) loadSubmittedIntent(ctx context.Context, intentID uuid.UUID) (*SubmitIntentResult, error) {
	intent, err := s.Intents.FindByID(ctx, intentID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	executions, err := s.Executions.ListByIntentID(ctx, intent.ID())
	if err != nil {
		return nil, mapRepoErr(err)
	}
	result := &SubmitIntentResult{Intent: intent, Executions: executions, Replayed: true}
	if len(executions) > 0 {
		result.Execution = executions[0]
	}
	return result, nil
}
