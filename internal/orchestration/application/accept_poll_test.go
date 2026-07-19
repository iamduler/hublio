package application

import (
	"testing"

	"github.com/google/uuid"
)

func TestPollRecordsFromPayload(t *testing.T) {
	t.Parallel()
	got := pollRecordsFromPayload(map[string]any{
		"records": []any{
			map[string]any{"id": "1"},
			map[string]any{"id": "2"},
		},
	})
	if len(got) != 2 || got[0]["id"] != "1" {
		t.Fatalf("%#v", got)
	}
	if pollRecordsFromPayload(nil) != nil {
		t.Fatal("nil payload")
	}
}

func TestPollNextCursor(t *testing.T) {
	t.Parallel()
	next := pollNextCursor(map[string]any{
		"next_cursor": map[string]any{"last_record_id": "9"},
	}, nil, nil)
	if next["last_record_id"] != "9" {
		t.Fatalf("%#v", next)
	}
	derived := pollNextCursor(nil, map[string]any{"keep": true}, []map[string]any{
		{"id": "a", "updated_at": "2026-01-02T00:00:00Z"},
	})
	if derived["last_record_id"] != "a" || derived["last_updated_at"] != "2026-01-02T00:00:00Z" || derived["keep"] != true {
		t.Fatalf("%#v", derived)
	}
}

func TestDerivePollIdempotencyKey_Stable(t *testing.T) {
	t.Parallel()
	ws := uuid.MustParse("018f0000-0000-7000-8000-000000000001")
	route := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	payload := map[string]any{"invoice_number": "INV-1"}
	a := derivePollIdempotencyKey(ws, route, "invoice", payload, nil)
	b := derivePollIdempotencyKey(ws, route, "invoice", payload, nil)
	if a == "" || a != b {
		t.Fatalf("%q %q", a, b)
	}
	if a[:5] != "poll_" {
		t.Fatalf("prefix: %q", a)
	}
}
