package application

import (
	"context"

	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// ListEvents returns the most recent PlatformEvents for a Workspace (tenant-scoped),
// optionally filtered to one Execution. Backs GET /api/v1/events.
func (s *Services) ListEvents(ctx context.Context, workspaceID uuid.UUID, executionID *uuid.UUID, limit int32) ([]*domain.PlatformEvent, error) {
	if s.Reader == nil {
		return nil, apperr.New("events reader not configured", apperr.ErrCodeInternal)
	}
	events, err := s.Reader.ListByWorkspace(ctx, workspaceID, executionID, limit)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to list events", apperr.ErrCodeInternal)
	}
	return events, nil
}
