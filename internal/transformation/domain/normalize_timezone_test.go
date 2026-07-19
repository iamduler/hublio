package domain

import "testing"

func TestNormalizeTimezone(t *testing.T) {
	tests := []struct {
		name    string
		doc     Document
		op      NormalizeTimezone
		want    string
		wantErr bool
	}{
		{
			name: "converts +07:00 offset to UTC",
			doc:  Document{"issue_date": "2026-07-15T10:00:00+07:00"},
			op:   NormalizeTimezone{Path: "issue_date", TargetTZ: "UTC"},
			want: "2026-07-15T03:00:00Z",
		},
		{
			name: "defaults to UTC when TargetTZ empty",
			doc:  Document{"issue_date": "2026-07-15T10:00:00+07:00"},
			op:   NormalizeTimezone{Path: "issue_date"},
			want: "2026-07-15T03:00:00Z",
		},
		{
			name: "no-op when path absent",
			doc:  Document{"hello": "world"},
			op:   NormalizeTimezone{Path: "issue_date", TargetTZ: "UTC"},
		},
		{
			name:    "invalid timestamp errors",
			doc:     Document{"issue_date": "not-a-date"},
			op:      NormalizeTimezone{Path: "issue_date", TargetTZ: "UTC"},
			wantErr: true,
		},
		{
			name:    "unknown timezone errors",
			doc:     Document{"issue_date": "2026-07-15T10:00:00+07:00"},
			op:      NormalizeTimezone{Path: "issue_date", TargetTZ: "Nowhere/Fake"},
			wantErr: true,
		},
		{
			name:    "missing params",
			doc:     Document{},
			op:      NormalizeTimezone{},
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
			if tt.want == "" {
				if len(out) != 1 || out["hello"] != "world" {
					t.Fatalf("expected document unchanged, got %v", out)
				}
				return
			}
			got, _ := out.Get(tt.op.Path)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
