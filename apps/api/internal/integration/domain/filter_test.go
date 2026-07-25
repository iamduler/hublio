package domain

import "testing"

func TestMatchFilter(t *testing.T) {
	t.Parallel()
	payload := map[string]any{
		"status": "paid",
		"total":  100.0,
		"customer": map[string]any{
			"name": "Acme",
		},
	}
	tests := []struct {
		name    string
		filter  map[string]any
		want    bool
		wantErr bool
	}{
		{name: "nil filter", filter: nil, want: true},
		{name: "eq match", filter: map[string]any{"op": "eq", "path": "status", "value": "paid"}, want: true},
		{name: "eq miss", filter: map[string]any{"op": "eq", "path": "status", "value": "draft"}, want: false},
		{name: "nested path", filter: map[string]any{"op": "eq", "path": "customer.name", "value": "Acme"}, want: true},
		{name: "gt", filter: map[string]any{"op": "gt", "path": "total", "value": 50}, want: true},
		{
			name: "and",
			filter: map[string]any{
				"op": "and",
				"args": []any{
					map[string]any{"op": "eq", "path": "status", "value": "paid"},
					map[string]any{"op": "gte", "path": "total", "value": 100},
				},
			},
			want: true,
		},
		{
			name: "or short-circuit",
			filter: map[string]any{
				"op": "or",
				"args": []any{
					map[string]any{"op": "eq", "path": "status", "value": "x"},
					map[string]any{"op": "eq", "path": "status", "value": "paid"},
				},
			},
			want: true,
		},
		{name: "unknown op", filter: map[string]any{"op": "bogus"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchFilter(tt.filter, payload)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("MatchFilter: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestMatchWebhookSecret(t *testing.T) {
	t.Parallel()
	if !MatchWebhookSecret("abc", "abc") {
		t.Fatal("expected match")
	}
	if MatchWebhookSecret("abc", "abd") {
		t.Fatal("expected mismatch")
	}
	if MatchWebhookSecret("", "") {
		t.Fatal("empty must not match")
	}
}
