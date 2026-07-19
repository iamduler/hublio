package domain

import "testing"

func TestDocumentGetSetDeletePath(t *testing.T) {
	doc := Document{
		"invoice_number": "INV-1",
		"customer": map[string]any{
			"name": "Acme",
		},
	}

	if v, ok := doc.Get("invoice_number"); !ok || v != "INV-1" {
		t.Fatalf("Get(invoice_number) = %v, %v", v, ok)
	}
	if v, ok := doc.Get("customer.name"); !ok || v != "Acme" {
		t.Fatalf("Get(customer.name) = %v, %v", v, ok)
	}
	if _, ok := doc.Get("customer.missing"); ok {
		t.Fatalf("Get(customer.missing) expected ok=false")
	}
	if _, ok := doc.Get("missing.nested"); ok {
		t.Fatalf("Get(missing.nested) expected ok=false")
	}

	doc.Set("customer.address.city", "Hanoi")
	if v, ok := doc.Get("customer.address.city"); !ok || v != "Hanoi" {
		t.Fatalf("Set/Get nested path failed, got %v, %v", v, ok)
	}

	doc.Delete("customer.name")
	if _, ok := doc.Get("customer.name"); ok {
		t.Fatalf("Delete(customer.name) did not remove field")
	}
}

func TestDocumentCloneIsIndependent(t *testing.T) {
	original := Document{
		"customer": map[string]any{"name": "Acme"},
		"lines":    []any{map[string]any{"qty": 1}},
	}
	clone := original.Clone()

	clone.Set("customer.name", "Changed")
	clone["lines"].([]any)[0].(map[string]any)["qty"] = 99

	if v, _ := original.Get("customer.name"); v != "Acme" {
		t.Fatalf("mutating clone leaked into original: customer.name = %v", v)
	}
	origLines := original["lines"].([]any)
	if origLines[0].(map[string]any)["qty"] != 1 {
		t.Fatalf("mutating clone leaked into original slice: qty = %v", origLines[0].(map[string]any)["qty"])
	}
}
