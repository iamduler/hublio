package domain

import "testing"

func TestSetDefault(t *testing.T) {
	tests := []struct {
		name string
		doc  Document
		op   SetDefault
		want any
	}{
		{
			name: "sets missing field",
			doc:  Document{},
			op:   SetDefault{Path: "status", Value: "pending"},
			want: "pending",
		},
		{
			name: "sets null field",
			doc:  Document{"status": nil},
			op:   SetDefault{Path: "status", Value: "pending"},
			want: "pending",
		},
		{
			name: "leaves existing non-null value untouched",
			doc:  Document{"status": "paid"},
			op:   SetDefault{Path: "status", Value: "pending"},
			want: "paid",
		},
		{
			name: "leaves zero-value untouched",
			doc:  Document{"retry_count": 0.0},
			op:   SetDefault{Path: "retry_count", Value: 3},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.op.Apply(tt.doc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, _ := out.Get(tt.op.Path)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDefaultMissingParams(t *testing.T) {
	op := SetDefault{}
	if _, err := op.Apply(Document{}); err == nil {
		t.Fatalf("expected error for missing path")
	}
}
