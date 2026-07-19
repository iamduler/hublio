package application

import (
	"context"
	"testing"

	"hublio/internal/transformation/domain"
)

func TestServicesTransformEmptySpecIsIdentity(t *testing.T) {
	svc := NewServices()

	result, err := svc.Transform(context.Background(), TransformInput{
		Direction: "request",
		Document:  map[string]any{"hello": "world"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Document) != 1 || result.Document["hello"] != "world" {
		t.Fatalf("expected identity transform, got %v", result.Document)
	}
}

func TestServicesTransformInvoiceFixture(t *testing.T) {
	svc := NewServices()

	fixture := map[string]any{
		"invoice_number": "INV-1",
		"issue_date":     "2026-07-15T10:00:00+07:00",
		"currency":       "vnd",
		"total":          "1000.50",
		"buyer_name":     "Acme",
		"lines": []any{
			map[string]any{"qty": 1.0, "unit_price": 1000.50},
		},
	}

	result, err := svc.Transform(context.Background(), TransformInput{
		Direction: "request",
		Document:  fixture,
		Spec:      domain.DefaultRequestPipelineSpec(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := domain.Document(result.Document)
	if v, _ := doc.Get("customer.name"); v != "Acme" {
		t.Fatalf("customer.name = %v, want Acme", v)
	}
	if v, _ := doc.Get("total"); v != 1000.50 {
		t.Fatalf("total = %v, want 1000.50", v)
	}
	if v, _ := doc.Get("currency"); v != "VND" {
		t.Fatalf("currency = %v, want VND", v)
	}
	if v, _ := doc.Get("issue_date"); v != "2026-07-15T03:00:00Z" {
		t.Fatalf("issue_date = %v, want 2026-07-15T03:00:00Z", v)
	}
	if v, _ := doc.Get("status"); v != "pending" {
		t.Fatalf("status = %v, want pending", v)
	}
}

func TestServicesTransformInvoiceFixtureMissingRequiredFieldFails(t *testing.T) {
	svc := NewServices()

	_, err := svc.Transform(context.Background(), TransformInput{
		Direction: "request",
		Document:  map[string]any{"currency": "vnd"},
		Spec:      domain.DefaultRequestPipelineSpec(),
	})
	if err == nil {
		t.Fatalf("expected validation error for missing invoice_number/issue_date")
	}
}

func TestServicesTransformEchoPayloadUnchangedWithNilSpec(t *testing.T) {
	svc := NewServices()

	// Mirrors how the Orchestration adapter treats a non-invoice-like capability such as the
	// Fake connector's "fake.echo": no Spec at all, so the Document must pass through as-is.
	result, err := svc.Transform(context.Background(), TransformInput{
		Direction: "request",
		Document:  map[string]any{"hello": "world"},
		Spec:      nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Document) != 1 || result.Document["hello"] != "world" {
		t.Fatalf("expected echo payload unchanged, got %v", result.Document)
	}
}
