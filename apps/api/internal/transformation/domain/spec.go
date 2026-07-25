package domain

import "fmt"

// OperationType names one of the built-in Canonical Operations.
type OperationType string

const (
	OpTypeRenameField       OperationType = "rename_field"
	OpTypeConvertType       OperationType = "convert_type"
	OpTypeNormalizeTimezone OperationType = "normalize_timezone"
	OpTypeNormalizeCurrency OperationType = "normalize_currency"
	OpTypeSetDefault        OperationType = "set_default"
	OpTypeValidateRequired  OperationType = "validate_required"
)

// OperationSpec is a serializable description of one Operation (type + params). It lets
// Orchestration (or any other caller) describe a Pipeline without depending on the concrete
// Operation structs, and without the Domain knowing anything about HTTP/DB/provider details.
type OperationSpec struct {
	Type   OperationType
	Params map[string]any
}

// BuildPipeline turns a slice of specs into an executable Pipeline. A nil/empty slice yields
// an identity Pipeline (Run returns a clone of the input Document, unchanged).
func BuildPipeline(specs []OperationSpec) (*Pipeline, error) {
	ops := make([]Operation, 0, len(specs))
	for _, spec := range specs {
		op, err := buildOperation(spec)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return NewPipeline(ops...), nil
}

func buildOperation(spec OperationSpec) (Operation, error) {
	switch spec.Type {
	case OpTypeRenameField:
		return RenameField{
			From: paramString(spec.Params, "from"),
			To:   paramString(spec.Params, "to"),
		}, nil
	case OpTypeConvertType:
		return ConvertType{
			Path: paramString(spec.Params, "path"),
			To:   ConvertKind(paramString(spec.Params, "to")),
		}, nil
	case OpTypeNormalizeTimezone:
		return NormalizeTimezone{
			Path:     paramString(spec.Params, "path"),
			TargetTZ: paramString(spec.Params, "target_tz"),
		}, nil
	case OpTypeNormalizeCurrency:
		return NormalizeCurrency{
			Path:          paramString(spec.Params, "path"),
			CurrencyField: paramString(spec.Params, "currency_field"),
		}, nil
	case OpTypeSetDefault:
		return SetDefault{
			Path:  paramString(spec.Params, "path"),
			Value: spec.Params["value"],
		}, nil
	case OpTypeValidateRequired:
		return ValidateRequired{Paths: paramStringSlice(spec.Params, "paths")}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownOperation, spec.Type)
	}
}

func paramString(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	s, _ := params[key].(string)
	return s
}

func paramStringSlice(params map[string]any, key string) []string {
	if params == nil {
		return nil
	}
	switch v := params[key].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
