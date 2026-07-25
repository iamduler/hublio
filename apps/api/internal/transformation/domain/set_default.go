package domain

import "fmt"

// SetDefault sets Path to Value only when the field is missing or null. Existing values
// (including zero values like "" or 0) are left untouched.
type SetDefault struct {
	Path  string
	Value any
}

func (op SetDefault) Apply(doc Document) (Document, error) {
	if op.Path == "" {
		return nil, fmt.Errorf("%w: set_default requires path", ErrMissingParam)
	}
	if value, ok := doc.Get(op.Path); ok && value != nil {
		return doc, nil
	}
	doc.Set(op.Path, op.Value)
	return doc, nil
}
