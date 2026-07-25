package domain

import "testing"

func TestRenameField(t *testing.T) {
	tests := []struct {
		name    string
		doc     Document
		op      RenameField
		wantErr bool
		check   func(t *testing.T, doc Document)
	}{
		{
			name: "renames existing field to nested path",
			doc:  Document{"buyer_name": "Acme"},
			op:   RenameField{From: "buyer_name", To: "customer.name"},
			check: func(t *testing.T, doc Document) {
				if _, ok := doc.Get("buyer_name"); ok {
					t.Fatalf("expected buyer_name to be removed")
				}
				if v, ok := doc.Get("customer.name"); !ok || v != "Acme" {
					t.Fatalf("expected customer.name = Acme, got %v, %v", v, ok)
				}
			},
		},
		{
			name: "no-op when source field absent",
			doc:  Document{"hello": "world"},
			op:   RenameField{From: "buyer_name", To: "customer.name"},
			check: func(t *testing.T, doc Document) {
				if len(doc) != 1 || doc["hello"] != "world" {
					t.Fatalf("expected document unchanged, got %v", doc)
				}
			},
		},
		{
			name:    "missing params",
			doc:     Document{},
			op:      RenameField{},
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
