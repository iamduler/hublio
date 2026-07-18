package application

import (
	"context"

	"github.com/google/uuid"

	"hublio/internal/orchestration/domain"
)

type RetryExecutionResult struct {
	Execution *domain.Execution
	// Job must be enqueued by the caller AFTER the surrounding transaction commits.
	Job *ExecutionJob
}

// RetryExecution manually retries a Failed Execution (Failed -> Queued).
// Caller must enqueue Job after commit. Replay (new Execution for same Intent) is deferred:
// executions.intent_id is UNIQUE in the v1 schema.
func (s *Services) RetryExecution(ctx context.Context, workspaceID, executionID uuid.UUID) (*RetryExecutionResult, error) {
	execution, intent, err := s.findWorkspaceExecution(ctx, workspaceID, executionID)
	if err != nil {
		return nil, err
	}

	now := s.clock().Now()
	if err := execution.ScheduleRetry(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.appendTimeline(execution, "execution_retry_requested", "manual retry requested", map[string]any{
		"retry_attempt": execution.RetryAttempt(),
	}, now); err != nil {
		return nil, err
	}
	if err := s.Executions.Update(ctx, execution); err != nil {
		return nil, mapRepoErr(err)
	}

	return &RetryExecutionResult{
		Execution: execution,
		Job: &ExecutionJob{
			ExecutionID:    execution.ID(),
			IntentID:       intent.ID(),
			OrganizationID: intent.OrganizationID(),
			WorkspaceID:    intent.WorkspaceID(),
			CorrelationID:  intent.CorrelationID(),
		},
	}, nil
}
