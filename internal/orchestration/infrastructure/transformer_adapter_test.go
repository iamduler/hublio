package infrastructure

import (
	"context"
	"testing"

	transformationapp "hublio/internal/transformation/application"
)

func TestTransformerAdapterEchoCapabilityPassesThroughUnchanged(t *testing.T) {
	adapter := NewTransformerAdapter(transformationapp.NewServices())

	doc := map[string]any{"hello": "world"}
	got, err := adapter.TransformRequest(context.Background(), "fake.echo", doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["hello"] != "world" {
		t.Fatalf("expected echo payload unchanged, got %v", got)
	}
}

func TestTransformerAdapterInvoiceCreateNormalizes(t *testing.T) {
	adapter := NewTransformerAdapter(transformationapp.NewServices())

	doc := map[string]any{
		"invoice_number": "INV-1",
		"issue_date":     "2026-07-15T10:00:00+07:00",
		"currency":       "vnd",
		"total":          "1000.50",
		"buyer_name":     "Acme",
	}
	got, err := adapter.TransformRequest(context.Background(), "misa.invoice.create", doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	customer, ok := got["customer"].(map[string]any)
	if !ok || customer["name"] != "Acme" {
		t.Fatalf("expected customer.name = Acme, got %v", got["customer"])
	}
	if got["currency"] != "VND" {
		t.Fatalf("expected currency VND, got %v", got["currency"])
	}
}

func TestTransformerAdapterInvoiceGetDoesNotRequireInvoiceFields(t *testing.T) {
	adapter := NewTransformerAdapter(transformationapp.NewServices())

	doc := map[string]any{"id": 99}
	got, err := adapter.TransformRequest(context.Background(), "invoice.get", doc)
	if err != nil {
		t.Fatalf("invoice.get should not ValidateRequired invoice fields: %v", err)
	}
	if got["id"] != 99 {
		t.Fatalf("expected identity transform, got %v", got)
	}
}

func TestTransformerAdapterInvoiceUpdateStatusPassesThrough(t *testing.T) {
	adapter := NewTransformerAdapter(transformationapp.NewServices())

	doc := map[string]any{"order_id": 55, "status": 2}
	got, err := adapter.TransformRequest(context.Background(), "nhanh.invoice.update_status", doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["order_id"] != 55 || got["status"] != 2 {
		t.Fatalf("expected identity transform, got %v", got)
	}
}

func TestIsInvoiceCreateCapability(t *testing.T) {
	tests := []struct {
		capability string
		want       bool
	}{
		{"fake.echo", false},
		{"misa.invoice.create", true},
		{"invoice.create", true},
		{"nhanh.CreateInvoice", true},
		{"invoice.publish", true},
		{"invoice.get", false},
		{"invoice.update_status", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isInvoiceCreateCapability(tt.capability); got != tt.want {
			t.Fatalf("%q: got %v want %v", tt.capability, got, tt.want)
		}
	}
}
