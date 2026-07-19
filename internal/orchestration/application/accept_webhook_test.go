package application

import (
	"testing"

	"github.com/google/uuid"
)

func TestDeriveWebhookIdempotencyKey_Stable(t *testing.T) {
	t.Parallel()
	ws := uuid.Must(uuid.NewV7())
	route := uuid.Must(uuid.NewV7())
	payload := map[string]any{"invoice_number": "INV-1", "total": 10}
	a := deriveWebhookIdempotencyKey(ws, route, "invoice", "create", payload, nil)
	b := deriveWebhookIdempotencyKey(ws, route, "invoice", "create", payload, nil)
	if a != b || a == "" {
		t.Fatalf("expected stable key, got %q / %q", a, b)
	}
	c := deriveWebhookIdempotencyKey(ws, route, "invoice", "update", payload, nil)
	if a == c {
		t.Fatal("operation should change key")
	}
}

func TestWebhookBusinessKey_FromRuleFields(t *testing.T) {
	t.Parallel()
	payload := map[string]any{"account_id": "a1", "record_id": "r9"}
	rule := map[string]any{"fields": []any{"account_id", "record_id"}}
	if got := webhookBusinessKey(payload, rule); got != "a1:r9" {
		t.Fatalf("got %q", got)
	}
}
