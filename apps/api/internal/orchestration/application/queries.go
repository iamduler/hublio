package application

import (
	"context"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

const (
	defaultListLimit int32 = 50
	maxListLimit     int32 = 200
)

func clampListLimit(limit int32) int32 {
	if limit <= 0 || limit > maxListLimit {
		return defaultListLimit
	}
	return limit
}

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

// ListIntents returns workspace-scoped Intents, newest first.
func (s *Services) ListIntents(ctx context.Context, workspaceID uuid.UUID, status *string, limit int32) ([]*domain.Intent, error) {
	limit = clampListLimit(limit)
	list, err := s.Intents.ListByWorkspace(ctx, workspaceID, status, limit)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
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

// ListExecutions returns workspace-scoped Executions (via Intent join), newest first.
// Rows are slim (no Steps/Timeline hydration).
func (s *Services) ListExecutions(ctx context.Context, workspaceID uuid.UUID, status *string, limit int32) ([]*domain.Execution, error) {
	limit = clampListLimit(limit)
	list, err := s.Executions.ListByWorkspace(ctx, workspaceID, status, limit)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

func (s *Services) findWorkspaceExecution(ctx context.Context, workspaceID, executionID uuid.UUID) (*domain.Execution, *domain.Intent, error) {
	return s.GetExecution(ctx, workspaceID, executionID)
}
