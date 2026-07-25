package application

import (
	"context"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// GetIntent returns a workspace-scoped Intent, verifying tenant ownership.
func (s *Services) GetIntent(ctx context.Context, workspaceID, intentID uuid.UUID) (*domain.Intent, error) {
	intent, err := s.Intents.FindByID(ctx, intentID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if intent.WorkspaceID() != workspaceID {
		return nil, apperr.New("intent not found", apperr.ErrCodeNotFound)
	}
	return intent, nil
}

// GetExecution returns a workspace-scoped Execution (with its parent Intent), verifying
// tenant ownership via the Intent.
func (s *Services) GetExecution(ctx context.Context, workspaceID, executionID uuid.UUID) (*domain.Execution, *domain.Intent, error) {
	execution, err := s.Executions.FindByID(ctx, executionID)
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	intent, err := s.Intents.FindByID(ctx, execution.IntentID())
	if err != nil {
		return nil, nil, mapRepoErr(err)
	}
	if intent.WorkspaceID() != workspaceID {
		return nil, nil, apperr.New("execution not found", apperr.ErrCodeNotFound)
	}
	return execution, intent, nil
}

func (s *Services) findWorkspaceExecution(ctx context.Context, workspaceID, executionID uuid.UUID) (*domain.Execution, *domain.Intent, error) {
	return s.GetExecution(ctx, workspaceID, executionID)
}
