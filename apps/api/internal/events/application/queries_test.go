package application

import (
	"context"
	"testing"
	"time"

	"hublio/internal/events/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

type memEventReader struct {
	page *domain.EventListPage
	err  error
	last domain.EventListFilter
}

func (m *memEventReader) ListByWorkspace(_ context.Context, _ uuid.UUID, filter domain.EventListFilter) (*domain.EventListPage, error) {
	m.last = filter
	if m.err != nil {
		return nil, m.err
	}
	if m.page != nil {
		return m.page, nil
	}
	return &domain.EventListPage{Events: nil}, nil
}

func TestListEvents_ClampsLimit(t *testing.T) {
	reader := &memEventReader{}
	svc := &Services{Reader: reader}
	ws := uuid.Must(uuid.NewV7())

	cases := []struct {
		name  string
		limit int32
		want  int32
	}{
		{"zero defaults", 0, 50},
		{"negative defaults", -1, 50},
		{"over max", 500, 50},
		{"ok", 25, 25},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := svc.ListEvents(context.Background(), ListEventsInput{
				WorkspaceID: ws,
				Limit:       tc.limit,
			})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if reader.last.Limit != tc.want {
				t.Fatalf("filter limit = %d, want %d", reader.last.Limit, tc.want)
			}
			if res.Limit != tc.want {
				t.Fatalf("result limit = %d, want %d", res.Limit, tc.want)
			}
		})
	}
}

func TestListEvents_InvalidCursor(t *testing.T) {
	svc := &Services{Reader: &memEventReader{}}
	_, err := svc.ListEvents(context.Background(), ListEventsInput{
		WorkspaceID: uuid.Must(uuid.NewV7()),
		Cursor:      "not-a-cursor",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*apperr.AppError)
	if !ok || appErr.Code != apperr.ErrCodeBadRequest {
		t.Fatalf("got %v, want BAD_REQUEST", err)
	}
}

func TestListEvents_CursorRoundTrip(t *testing.T) {
	ws := uuid.Must(uuid.NewV7())
	id := uuid.Must(uuid.NewV7())
	created := time.Date(2026, 7, 30, 12, 0, 0, 123456789, time.UTC)
	event := domain.ReconstitutePlatformEvent(
		id, nil, &ws, domain.AggregateTypeExecution, uuid.Must(uuid.NewV7()), nil,
		domain.CategoryRuntime, "ExecutionSucceeded", "", nil, nil, "test", created,
	)
	reader := &memEventReader{
		page: &domain.EventListPage{
			Events:  []*domain.PlatformEvent{event},
			HasNext: true,
			Next:    &domain.EventListCursor{CreatedAt: created, ID: id},
		},
	}
	svc := &Services{Reader: reader}

	page1, err := svc.ListEvents(context.Background(), ListEventsInput{WorkspaceID: ws, Limit: 1})
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if !page1.HasNext || page1.NextCursor == "" {
		t.Fatalf("expected next cursor, got %+v", page1)
	}

	page2Reader := &memEventReader{page: &domain.EventListPage{Events: nil, HasNext: false}}
	svc.Reader = page2Reader
	page2, err := svc.ListEvents(context.Background(), ListEventsInput{
		WorkspaceID: ws,
		Limit:       1,
		Cursor:      page1.NextCursor,
	})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if page2.HasNext || page2.NextCursor != "" {
		t.Fatalf("expected empty next, got %+v", page2)
	}
	if page2Reader.last.Cursor == nil {
		t.Fatal("expected cursor passed to reader")
	}
	if !page2Reader.last.Cursor.CreatedAt.Equal(created) {
		t.Fatalf("cursor time = %v, want %v", page2Reader.last.Cursor.CreatedAt, created)
	}
	if page2Reader.last.Cursor.ID != id {
		t.Fatalf("cursor id = %v, want %v", page2Reader.last.Cursor.ID, id)
	}
}

func TestListEvents_PassesCategory(t *testing.T) {
	reader := &memEventReader{}
	svc := &Services{Reader: reader}
	cat := domain.CategoryBusiness
	_, err := svc.ListEvents(context.Background(), ListEventsInput{
		WorkspaceID: uuid.Must(uuid.NewV7()),
		Category:    &cat,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if reader.last.Category == nil || *reader.last.Category != domain.CategoryBusiness {
		t.Fatalf("category = %v, want business", reader.last.Category)
	}
}

func TestEncodeDecodeEventCursor(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	created := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	encoded := encodeEventCursor(&domain.EventListCursor{CreatedAt: created, ID: id})
	decoded, err := decodeEventCursor(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.ID != id || !decoded.CreatedAt.Equal(created) {
		t.Fatalf("got %+v", decoded)
	}
	if encodeEventCursor(nil) != "" {
		t.Fatal("nil cursor should encode empty")
	}
}
