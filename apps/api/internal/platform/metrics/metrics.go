// Package metrics provides simple in-memory, per-process counters for Phase F
// Observability. It intentionally avoids a Prometheus client dependency: counters are
// exposed as plain JSON (see internal/platform/server) which is enough for the current
// exit criteria. Counters are constructor-injected (no package-level globals) so the API
// and worker binaries each own their own instance.
package metrics

import "sync/atomic"

// Counters holds atomic, in-memory counters. Safe for concurrent use.
type Counters struct {
	executionsSucceeded atomic.Int64
	executionsFailed    atomic.Int64
	eventsPublished     atomic.Int64
	auditRecords        atomic.Int64
}

// New returns a fresh, zeroed Counters instance.
func New() *Counters {
	return &Counters{}
}

func (c *Counters) IncExecutionsSucceeded() { c.executionsSucceeded.Add(1) }
func (c *Counters) IncExecutionsFailed()    { c.executionsFailed.Add(1) }
func (c *Counters) IncEventsPublished()     { c.eventsPublished.Add(1) }
func (c *Counters) IncAuditRecords()        { c.auditRecords.Add(1) }

// Snapshot is a point-in-time, read-only copy of all counters.
type Snapshot struct {
	ExecutionsSucceeded int64 `json:"executions_succeeded"`
	ExecutionsFailed    int64 `json:"executions_failed"`
	EventsPublished     int64 `json:"events_published"`
	AuditRecords        int64 `json:"audit_records"`
}

func (c *Counters) Snapshot() Snapshot {
	return Snapshot{
		ExecutionsSucceeded: c.executionsSucceeded.Load(),
		ExecutionsFailed:    c.executionsFailed.Load(),
		EventsPublished:     c.eventsPublished.Load(),
		AuditRecords:        c.auditRecords.Load(),
	}
}
