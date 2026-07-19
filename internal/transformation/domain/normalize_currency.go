package domain

import (
	"fmt"
	"strings"
)

// NormalizeCurrency uppercases an ISO 4217 currency code at Path (docs/06 §11). Path may
// point at a plain currency string ("currency": "vnd") or at a nested Money object
// ({"amount": ..., "currency": "vnd"}, docs/06 §17); CurrencyField selects the sub-field for
// the Money case and defaults to "currency". It is a no-op when Path is absent or null, so a
// Pipeline stays safe against Canonical payloads without the field.
type NormalizeCurrency struct {
	Path          string
	CurrencyField string
}

func (op NormalizeCurrency) Apply(doc Document) (Document, error) {
	if op.Path == "" {
		return nil, fmt.Errorf("%w: normalize_currency requires path", ErrMissingParam)
	}
	value, ok := doc.Get(op.Path)
	if !ok || value == nil {
		return doc, nil
	}

	switch v := value.(type) {
	case string:
		doc.Set(op.Path, strings.ToUpper(v))
	case map[string]any:
		field := op.CurrencyField
		if field == "" {
			field = "currency"
		}
		code, ok := v[field].(string)
		if !ok {
			return doc, nil
		}
		v[field] = strings.ToUpper(code)
		doc.Set(op.Path, v)
	default:
		return nil, fmt.Errorf("%w: field %q is not a currency string or money object", ErrTypeConversion, op.Path)
	}
	return doc, nil
}
