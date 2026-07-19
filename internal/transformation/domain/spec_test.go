package domain

import "testing"

func TestBuildPipelineEmptySpecIsIdentity(t *testing.T) {
	pipeline, err := BuildPipeline(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := pipeline.Run(Document{"hello": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["hello"] != "world" || len(out) != 1 {
		t.Fatalf("expected identity transform, got %v", out)
	}
}

func TestBuildPipelineFromSpecs(t *testing.T) {
	specs := []OperationSpec{
		{Type: OpTypeRenameField, Params: map[string]any{"from": "buyer_name", "to": "customer.name"}},
		{Type: OpTypeConvertType, Params: map[string]any{"path": "total", "to": "number"}},
		{Type: OpTypeNormalizeCurrency, Params: map[string]any{"path": "currency"}},
		{Type: OpTypeNormalizeTimezone, Params: map[string]any{"path": "issue_date", "target_tz": "UTC"}},
		{Type: OpTypeSetDefault, Params: map[string]any{"path": "status", "value": "pending"}},
		{Type: OpTypeValidateRequired, Params: map[string]any{"paths": []string{"invoice_number"}}},
	}

	pipeline, err := BuildPipeline(specs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := pipeline.Run(Document{
		"invoice_number": "INV-1",
		"issue_date":     "2026-07-15T10:00:00+07:00",
		"currency":       "vnd",
		"total":          "1000.50",
		"buyer_name":     "Acme",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, _ := out.Get("customer.name"); v != "Acme" {
		t.Fatalf("customer.name = %v, want Acme", v)
	}
	if v, _ := out.Get("currency"); v != "VND" {
		t.Fatalf("currency = %v, want VND", v)
	}
	if v, _ := out.Get("status"); v != "pending" {
		t.Fatalf("status = %v, want pending", v)
	}
}

func TestBuildPipelineUnknownOperationType(t *testing.T) {
	_, err := BuildPipeline([]OperationSpec{{Type: "not_a_real_op"}})
	if err == nil {
		t.Fatalf("expected error for unknown operation type")
	}
}

func TestDefaultPipelineSpecsAreSafeOnUnrelatedPayload(t *testing.T) {
	// RenameField/ConvertType/NormalizeCurrency/NormalizeTimezone must all be true no-ops on
	// a payload that carries none of the invoice fields. SetDefault/ValidateRequired are
	// intentionally excluded here: SetDefault always adds its field when missing (by design),
	// and ValidateRequired would fail such a payload — which is exactly why Orchestration
	// only applies the full default pipeline to invoice-like capabilities (see
	// internal/orchestration/infrastructure/transformer_adapter.go).
	requestSpec := DefaultRequestPipelineSpec()
	safeOps := requestSpec[:4]

	pipeline, err := BuildPipeline(safeOps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := pipeline.Run(Document{"hello": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out["hello"] != "world" {
		t.Fatalf("expected unrelated payload untouched, got %v", out)
	}
}
