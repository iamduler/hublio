package domain

import "strings"

// Document is a Canonical payload (Canonical Resource, per docs/06). It is never a Provider
// DTO: only Connector Runtime is allowed to know provider shapes. Nested objects/arrays use
// map[string]any / []any, matching every other canonical boundary in Hublio (Intent.Payload,
// Execution.Context, Connector Runtime Invoke).
type Document map[string]any

// Clone returns a deep copy so Operations never mutate a caller's map by reference (e.g. the
// Execution context or the Intent payload it was read from).
func (d Document) Clone() Document {
	return Document(cloneMap(d))
}

// Get resolves a dot-notation path (e.g. "customer.name") against nested map[string]any
// values. It returns false when any segment is missing or not a map.
func (d Document) Get(path string) (any, bool) {
	return getPath(map[string]any(d), path)
}

// Set writes value at a dot-notation path, creating intermediate objects as needed.
func (d Document) Set(path string, value any) {
	setPath(map[string]any(d), path, value)
}

// Delete removes the field at a dot-notation path. No-op when the path does not exist.
func (d Document) Delete(path string) {
	deletePath(map[string]any(d), path)
}

func splitPath(path string) []string {
	return strings.Split(path, ".")
}

func getPath(m map[string]any, path string) (any, bool) {
	parts := splitPath(path)
	var cur any = m
	for _, part := range parts {
		curMap, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, exists := curMap[part]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func setPath(m map[string]any, path string, value any) {
	parts := splitPath(path)
	cur := m
	for i, part := range parts {
		if i == len(parts)-1 {
			cur[part] = value
			return
		}
		next, ok := cur[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[part] = next
		}
		cur = next
	}
}

func deletePath(m map[string]any, path string) {
	parts := splitPath(path)
	cur := m
	for i, part := range parts {
		if i == len(parts)-1 {
			delete(cur, part)
			return
		}
		next, ok := cur[part].(map[string]any)
		if !ok {
			return
		}
		cur = next
	}
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = cloneValue(v)
	}
	return out
}

func cloneValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return cloneMap(t)
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = cloneValue(item)
		}
		return out
	default:
		return v
	}
}
