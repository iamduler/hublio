package infrastructure

import (
	"context"
	"testing"

	transformationapp "hublio/internal/transformation/application"
	transformationdomain "hublio/internal/transformation/domain"
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

func TestTransformerAdapterInvoiceCapabilityNormalizes(t *testing.T) {
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

func TestSpecForCapability(t *testing.T) {
	defaultSpec := transformationdomain.DefaultRequestPipelineSpec()
	tests := []struct {
		name       string
		capability string
		wantNil    bool
	}{
		{name: "fake echo", capability: "fake.echo", wantNil: true},
		{name: "invoice lowercase", capability: "misa.invoice.create", wantNil: false},
		{name: "invoice mixed case", capability: "nhanh.CreateInvoice", wantNil: false},
		{name: "empty", capability: "", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := specForCapability(tt.capability, defaultSpec)
			if tt.wantNil && spec != nil {
				t.Fatalf("expected nil spec, got %v", spec)
			}
			if !tt.wantNil && spec == nil {
				t.Fatalf("expected non-nil spec")
			}
		})
	}
}
