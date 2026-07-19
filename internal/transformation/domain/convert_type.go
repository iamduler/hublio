package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ConvertKind is a supported Canonical type conversion target (docs/06 §8).
type ConvertKind string

const (
	ConvertToString ConvertKind = "string"
	ConvertToNumber ConvertKind = "number"
	ConvertToBool   ConvertKind = "bool"
	ConvertToTime   ConvertKind = "time"
)

// ConvertType converts the value at Path to the target Kind. It is a no-op when Path is
// absent or null, so a Pipeline stays safe against Canonical payloads without the field.
// Conversions are deterministic (docs/06 §8): no locale-aware parsing/formatting.
type ConvertType struct {
	Path string
	To   ConvertKind
}

func (op ConvertType) Apply(doc Document) (Document, error) {
	if op.Path == "" {
		return nil, fmt.Errorf("%w: convert_type requires path", ErrMissingParam)
	}
	value, ok := doc.Get(op.Path)
	if !ok || value == nil {
		return doc, nil
	}
	converted, err := convertValue(value, op.To)
	if err != nil {
		return nil, fmt.Errorf("%w: field %q: %v", ErrTypeConversion, op.Path, err)
	}
	doc.Set(op.Path, converted)
	return doc, nil
}

func convertValue(value any, to ConvertKind) (any, error) {
	switch to {
	case ConvertToString:
		return toStringValue(value), nil
	case ConvertToNumber:
		return toNumberValue(value)
	case ConvertToBool:
		return toBoolValue(value)
	case ConvertToTime:
		return toTimeValue(value)
	default:
		return nil, fmt.Errorf("unsupported convert kind %q", to)
	}
}

func toStringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toNumberValue(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to number", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func toBoolValue(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return false, fmt.Errorf("cannot convert %q to bool", v)
		}
		return b, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

func toTimeValue(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTime(v)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}
