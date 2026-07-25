package domain

import (
	"errors"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name        string
		doc         Document
		op          ValidateRequired
		wantMissing []string
	}{
		{
			name: "all required fields present",
			doc:  Document{"invoice_number": "INV-1", "issue_date": "2026-07-15T10:00:00Z"},
			op:   ValidateRequired{Paths: []string{"invoice_number", "issue_date"}},
		},
		{
			name:        "missing field reported",
			doc:         Document{"invoice_number": "INV-1"},
			op:          ValidateRequired{Paths: []string{"invoice_number", "issue_date"}},
			wantMissing: []string{"issue_date"},
		},
		{
			name:        "null field treated as missing",
			doc:         Document{"invoice_number": "INV-1", "issue_date": nil},
			op:          ValidateRequired{Paths: []string{"invoice_number", "issue_date"}},
			wantMissing: []string{"issue_date"},
		},
		{
			name:        "empty string treated as missing",
			doc:         Document{"invoice_number": ""},
			op:          ValidateRequired{Paths: []string{"invoice_number"}},
			wantMissing: []string{"invoice_number"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.op.Apply(tt.doc)
			if len(tt.wantMissing) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var verr *ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("expected *ValidationError, got %v", err)
			}
			if len(verr.MissingFields) != len(tt.wantMissing) {
				t.Fatalf("got missing fields %v, want %v", verr.MissingFields, tt.wantMissing)
			}
			for i, f := range tt.wantMissing {
				if verr.MissingFields[i] != f {
					t.Fatalf("got missing fields %v, want %v", verr.MissingFields, tt.wantMissing)
				}
			}
		})
	}
}

func TestValidateRequiredMissingParams(t *testing.T) {
	op := ValidateRequired{}
	if _, err := op.Apply(Document{}); err == nil {
		t.Fatalf("expected error for empty paths")
	}
}
