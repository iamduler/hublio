package domain

import (
	"fmt"
	"strings"
	"time"
)

// NormalizeTimezone converts an RFC3339 timestamp at Path into TargetTZ (an IANA zone name,
// e.g. "Asia/Ho_Chi_Minh") and stores it back as RFC3339. TargetTZ defaults to UTC, matching
// the Canonical wire format documented in docs/06 §10. It is a no-op when Path is absent or
// null, so a Pipeline stays safe against Canonical payloads without the field.
type NormalizeTimezone struct {
	Path     string
	TargetTZ string
}

func (op NormalizeTimezone) Apply(doc Document) (Document, error) {
	if op.Path == "" {
		return nil, fmt.Errorf("%w: normalize_timezone requires path", ErrMissingParam)
	}
	value, ok := doc.Get(op.Path)
	if !ok || value == nil {
		return doc, nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case time.Time:
		str = v.Format(time.RFC3339)
	default:
		return nil, fmt.Errorf("%w: field %q is not a timestamp string", ErrTypeConversion, op.Path)
	}

	t, err := parseTime(str)
	if err != nil {
		return nil, fmt.Errorf("%w: field %q: %v", ErrTypeConversion, op.Path, err)
	}
	loc, err := resolveLocation(op.TargetTZ)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
	}
	doc.Set(op.Path, t.In(loc).Format(time.RFC3339))
	return doc, nil
}

func resolveLocation(tz string) (*time.Location, error) {
	if tz == "" || strings.EqualFold(tz, "UTC") {
		return time.UTC, nil
	}
	return time.LoadLocation(tz)
}
