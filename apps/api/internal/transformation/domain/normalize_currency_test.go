package domain

import "testing"

func TestNormalizeCurrency(t *testing.T) {
	tests := []struct {
		name    string
		doc     Document
		op      NormalizeCurrency
		check   func(t *testing.T, out Document)
		wantErr bool
	}{
		{
			name: "uppercases plain currency string",
			doc:  Document{"currency": "vnd"},
			op:   NormalizeCurrency{Path: "currency"},
			check: func(t *testing.T, out Document) {
				if v, _ := out.Get("currency"); v != "VND" {
					t.Fatalf("got %v, want VND", v)
				}
			},
		},
		{
			name: "uppercases nested money object currency field",
			doc:  Document{"total": map[string]any{"amount": 1000.5, "currency": "usd"}},
			op:   NormalizeCurrency{Path: "total"},
			check: func(t *testing.T, out Document) {
				money := out["total"].(map[string]any)
				if money["currency"] != "USD" {
					t.Fatalf("got %v, want USD", money["currency"])
				}
				if money["amount"] != 1000.5 {
					t.Fatalf("amount changed unexpectedly: %v", money["amount"])
				}
			},
		},
		{
			name: "no-op when path absent",
			doc:  Document{"hello": "world"},
			op:   NormalizeCurrency{Path: "currency"},
			check: func(t *testing.T, out Document) {
				if len(out) != 1 || out["hello"] != "world" {
					t.Fatalf("expected document unchanged, got %v", out)
				}
			},
		},
		{
			name:    "missing params",
			doc:     Document{},
			op:      NormalizeCurrency{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.op.Apply(tt.doc)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, out)
		})
	}
}
