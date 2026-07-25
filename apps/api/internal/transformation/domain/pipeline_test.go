package domain

import "testing"

func TestPipelineRunEmptyIsIdentity(t *testing.T) {
	doc := Document{"hello": "world"}
	pipeline := NewPipeline()

	out, err := pipeline.Run(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["hello"] != "world" || len(out) != 1 {
		t.Fatalf("expected identity transform, got %v", out)
	}

	// Run must not mutate the caller's Document.
	out.Set("hello", "changed")
	if doc["hello"] != "world" {
		t.Fatalf("Run mutated the caller's input Document")
	}
}

func TestPipelineRunStopsOnFirstError(t *testing.T) {
	pipeline := NewPipeline(
		SetDefault{Path: "status", Value: "pending"},
		ConvertType{Path: "total", To: ConvertToNumber},
	)
	_, err := pipeline.Run(Document{"total": "not-a-number"})
	if err == nil {
		t.Fatalf("expected error from second operation")
	}
}

// TestInvoiceLikeCanonicalFixture exercises the full invoice normalization pipeline described
// in the Phase E checklist against a Canonical fixture (no Provider DTO involved).
func TestInvoiceLikeCanonicalFixture(t *testing.T) {
	fixture := Document{
		"invoice_number": "INV-1",
		"issue_date":     "2026-07-15T10:00:00+07:00",
		"currency":       "vnd",
		"total":          "1000.50",
		"buyer_name":     "Acme",
		"lines": []any{
			map[string]any{"qty": 1.0, "unit_price": 1000.50},
		},
	}

	pipeline := NewPipeline(
		RenameField{From: "buyer_name", To: "customer.name"},
		ConvertType{Path: "total", To: ConvertToNumber},
		NormalizeCurrency{Path: "currency"},
		NormalizeTimezone{Path: "issue_date", TargetTZ: "UTC"},
		SetDefault{Path: "status", Value: "pending"},
		ValidateRequired{Paths: []string{"invoice_number", "issue_date", "customer.name"}},
	)

	out, err := pipeline.Run(fixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := out.Get("buyer_name"); ok {
		t.Fatalf("expected buyer_name renamed away")
	}
	if v, _ := out.Get("customer.name"); v != "Acme" {
		t.Fatalf("customer.name = %v, want Acme", v)
	}
	if v, _ := out.Get("total"); v != 1000.50 {
		t.Fatalf("total = %v, want 1000.50", v)
	}
	if v, _ := out.Get("currency"); v != "VND" {
		t.Fatalf("currency = %v, want VND", v)
	}
	if v, _ := out.Get("issue_date"); v != "2026-07-15T03:00:00Z" {
		t.Fatalf("issue_date = %v, want 2026-07-15T03:00:00Z", v)
	}
	if v, _ := out.Get("status"); v != "pending" {
		t.Fatalf("status = %v, want pending", v)
	}
	if v, _ := out.Get("invoice_number"); v != "INV-1" {
		t.Fatalf("invoice_number = %v, want INV-1", v)
	}
	lines, ok := out["lines"].([]any)
	if !ok || len(lines) != 1 {
		t.Fatalf("lines not preserved: %v", out["lines"])
	}

	// The original fixture Document must never be mutated: Transformation only ever
	// produces a new Canonical Document (docs/06 §2, "never persists business data").
	if _, ok := fixture.Get("customer.name"); ok {
		t.Fatalf("Pipeline.Run mutated the original fixture Document")
	}
}

func TestInvoiceLikeFixtureFailsValidationWhenRequiredFieldMissing(t *testing.T) {
	fixture := Document{
		"currency": "vnd",
	}
	pipeline := NewPipeline(
		ValidateRequired{Paths: []string{"invoice_number", "issue_date"}},
	)
	_, err := pipeline.Run(fixture)
	if err == nil {
		t.Fatalf("expected validation error for missing invoice_number/issue_date")
	}
}
