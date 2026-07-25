package domain

import "time"

// timeLayouts are tried in order when parsing a Canonical timestamp string. RFC3339 (with an
// explicit offset) is the Canonical wire format (docs/06 §10); the rest are accepted for
// leniency but every layout is deterministic (no locale-dependent parsing).
var timeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parseTime(s string) (time.Time, error) {
	var lastErr error
	for _, layout := range timeLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
