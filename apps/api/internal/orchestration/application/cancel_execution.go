package application

import (
	"context"

	"github.com/google/uuid"

	"hublio/internal/orchestration/domain"
)

// CancelExecution transitions a Queued/Running Execution to Cancelled (terminal).
func (s *Services) CancelExecution(ctx context.Context, workspaceID, executionID uuid.UUID) (*domain.Execution, error) {
	execution, _, err := s.findWorkspaceExecution(ctx, workspaceID, executionID)
	if err != nil {
		return nil, err
	}
	now := s.clock().Now()
	if err := execution.Cancel(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.appendTimeline(execution, "execution_cancelled", "cancelled by request", nil, now); err != nil {
		return nil, err
	}
	if err := s.Executions.Update(ctx, execution); err != nil {
		return nil, mapRepoErr(err)
	}
	return execution, nil
}
