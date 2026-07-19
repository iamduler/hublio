package domain

import "fmt"

// ValidateRequired fails the Pipeline when any of Paths is missing, null, or an empty string.
// Mapping errors must stop execution before the Connector Runtime is invoked (docs/06 §26).
type ValidateRequired struct {
	Paths []string
}

func (op ValidateRequired) Apply(doc Document) (Document, error) {
	if len(op.Paths) == 0 {
		return nil, fmt.Errorf("%w: validate_required requires at least one path", ErrMissingParam)
	}
	var missing []string
	for _, path := range op.Paths {
		value, ok := doc.Get(path)
		if !ok || value == nil {
			missing = append(missing, path)
			continue
		}
		if s, isStr := value.(string); isStr && s == "" {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		return nil, &ValidationError{MissingFields: missing}
	}
	return doc, nil
}
