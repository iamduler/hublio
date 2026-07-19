package domain

import "fmt"

// RenameField moves a value from From to To (dot-notation nested paths supported, e.g.
// "buyer_name" -> "customer.name"). It is a no-op when From is absent, so a Pipeline stays
// safe against Canonical payloads that never had the field.
type RenameField struct {
	From string
	To   string
}

func (op RenameField) Apply(doc Document) (Document, error) {
	if op.From == "" || op.To == "" {
		return nil, fmt.Errorf("%w: rename_field requires from/to", ErrMissingParam)
	}
	value, ok := doc.Get(op.From)
	if !ok {
		return doc, nil
	}
	doc.Delete(op.From)
	doc.Set(op.To, value)
	return doc, nil
}
