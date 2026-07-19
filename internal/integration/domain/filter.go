package domain

import (
	"fmt"
	"strings"
)

// MatchFilter evaluates a SyncRoute JSON filter tree against a Canonical-ish payload.
// Empty/nil filter always matches. This is a small condition tree (docs/30 §3), not a Rule Engine.
//
// Supported shapes:
//
//	{"op":"eq","path":"status","value":"paid"}
//	{"op":"and","args":[ ...nodes... ]}
//	{"op":"or","args":[ ...nodes... ]}
//
// Operators: eq, neq, in, gt, gte, lt, lte (numeric/string compare via fmt.Sprint for non-numbers).
func MatchFilter(filter map[string]any, payload map[string]any) (bool, error) {
	if len(filter) == 0 {
		return true, nil
	}
	return evalFilterNode(filter, payload)
}

func evalFilterNode(node map[string]any, payload map[string]any) (bool, error) {
	op, _ := node["op"].(string)
	op = strings.ToLower(strings.TrimSpace(op))
	switch op {
	case "and":
		args, err := filterArgs(node)
		if err != nil {
			return false, err
		}
		for _, arg := range args {
			ok, err := evalFilterNode(arg, payload)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	case "or":
		args, err := filterArgs(node)
		if err != nil {
			return false, err
		}
		if len(args) == 0 {
			return true, nil
		}
		for _, arg := range args {
			ok, err := evalFilterNode(arg, payload)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	case "eq", "neq", "in", "gt", "gte", "lt", "lte":
		path, _ := node["path"].(string)
		if strings.TrimSpace(path) == "" {
			return false, ErrInvalidFilter
		}
		actual, _ := getPath(payload, path)
		switch op {
		case "eq":
			return valuesEqual(actual, node["value"]), nil
		case "neq":
			return !valuesEqual(actual, node["value"]), nil
		case "in":
			list, ok := node["value"].([]any)
			if !ok {
				return false, ErrInvalidFilter
			}
			for _, v := range list {
				if valuesEqual(actual, v) {
					return true, nil
				}
			}
			return false, nil
		case "gt", "gte", "lt", "lte":
			return compareOrdered(op, actual, node["value"])
		default:
			return false, ErrInvalidFilter
		}
	default:
		return false, fmt.Errorf("%w: unknown op %q", ErrInvalidFilter, op)
	}
}

func filterArgs(node map[string]any) ([]map[string]any, error) {
	raw, ok := node["args"].([]any)
	if !ok {
		return nil, ErrInvalidFilter
	}
	out := make([]map[string]any, 0, len(raw))
	for _, a := range raw {
		m, ok := a.(map[string]any)
		if !ok {
			return nil, ErrInvalidFilter
		}
		out = append(out, m)
	}
	return out, nil
}

func getPath(doc map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = doc
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func compareOrdered(op string, actual, expected any) (bool, error) {
	af, aOK := asFloat(actual)
	bf, bOK := asFloat(expected)
	if aOK && bOK {
		switch op {
		case "gt":
			return af > bf, nil
		case "gte":
			return af >= bf, nil
		case "lt":
			return af < bf, nil
		case "lte":
			return af <= bf, nil
		}
	}
	as, bs := fmt.Sprint(actual), fmt.Sprint(expected)
	switch op {
	case "gt":
		return as > bs, nil
	case "gte":
		return as >= bs, nil
	case "lt":
		return as < bs, nil
	case "lte":
		return as <= bs, nil
	default:
		return false, ErrInvalidFilter
	}
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		var f float64
		n, err := fmt.Sscanf(strings.TrimSpace(fmt.Sprint(v)), "%f", &f)
		return f, err == nil && n == 1
	}
}
