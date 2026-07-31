package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// ListEventsResult is the application-facing page for GET /api/v1/events.
type ListEventsResult struct {
	Events     []*domain.PlatformEvent
	NextCursor string
	HasNext    bool
	Limit      int32
}

type eventCursorPayload struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

func encodeEventCursor(c *domain.EventListCursor) string {
	if c == nil {
		return ""
	}
	raw, err := json.Marshal(eventCursorPayload{
		T:  c.CreatedAt.UTC().Format(time.RFC3339Nano),
		ID: c.ID.String(),
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeEventCursor(raw string) (*domain.EventListCursor, error) {
	if raw == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, apperr.New("invalid cursor", apperr.ErrCodeBadRequest)
	}
	var payload eventCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, apperr.New("invalid cursor", apperr.ErrCodeBadRequest)
	}
	t, err := time.Parse(time.RFC3339Nano, payload.T)
	if err != nil {
		return nil, apperr.New("invalid cursor", apperr.ErrCodeBadRequest)
	}
	id, err := uuid.Parse(payload.ID)
	if err != nil {
		return nil, apperr.New("invalid cursor", apperr.ErrCodeBadRequest)
	}
	return &domain.EventListCursor{CreatedAt: t.UTC(), ID: id}, nil
}

// ListEventsInput carries query params for GET /api/v1/events.
type ListEventsInput struct {
	WorkspaceID uuid.UUID
	ExecutionID *uuid.UUID
	Category    *domain.Category
	Cursor      string
	Limit       int32
}

// ListEvents returns a workspace-scoped page of PlatformEvents.
func (s *Services) ListEvents(ctx context.Context, in ListEventsInput) (*ListEventsResult, error) {
	if s.Reader == nil {
		return nil, apperr.New("events reader not configured", apperr.ErrCodeInternal)
	}
	limit := in.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	cursor, err := decodeEventCursor(in.Cursor)
	if err != nil {
		return nil, err
	}
	page, err := s.Reader.ListByWorkspace(ctx, in.WorkspaceID, domain.EventListFilter{
		ExecutionID: in.ExecutionID,
		Category:    in.Category,
		Cursor:      cursor,
		Limit:       limit,
	})
	if err != nil {
		return nil, apperr.Wrap(err, "failed to list events", apperr.ErrCodeInternal)
	}
	return &ListEventsResult{
		Events:     page.Events,
		NextCursor: encodeEventCursor(page.Next),
		HasNext:    page.HasNext,
		Limit:      limit,
	}, nil
}
