package domain

import "testing"

func TestConvertType(t *testing.T) {
	tests := []struct {
		name    string
		doc     Document
		op      ConvertType
		want    any
		wantErr bool
	}{
		{
			name: "string to number",
			doc:  Document{"total": "1000.50"},
			op:   ConvertType{Path: "total", To: ConvertToNumber},
			want: 1000.50,
		},
		{
			name: "number stays number",
			doc:  Document{"total": 42.0},
			op:   ConvertType{Path: "total", To: ConvertToNumber},
			want: 42.0,
		},
		{
			name: "string to bool",
			doc:  Document{"active": "true"},
			op:   ConvertType{Path: "active", To: ConvertToBool},
			want: true,
		},
		{
			name: "number to string",
			doc:  Document{"qty": 5.0},
			op:   ConvertType{Path: "qty", To: ConvertToString},
			want: "5",
		},
		{
			name:    "invalid number conversion errors",
			doc:     Document{"total": "not-a-number"},
			op:      ConvertType{Path: "total", To: ConvertToNumber},
			wantErr: true,
		},
		{
			name: "no-op when path absent",
			doc:  Document{"hello": "world"},
			op:   ConvertType{Path: "total", To: ConvertToNumber},
			want: nil, // checked separately below
		},
		{
			name:    "missing params",
			doc:     Document{},
			op:      ConvertType{},
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
			if tt.name == "no-op when path absent" {
				if len(out) != 1 || out["hello"] != "world" {
					t.Fatalf("expected document unchanged, got %v", out)
				}
				return
			}
			got, _ := out.Get(tt.op.Path)
			if got != tt.want {
				t.Fatalf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}
