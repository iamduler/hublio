// Package queue implements Platform Infrastructure work queues.
//
// Work queues (Redis → Queue → Worker) are NOT part of the Event Platform.
// The Event Platform publishes domain/runtime events after state changes;
// this package schedules asynchronous jobs for workers (for example Executions).
package queue

import (
	"context"
	"time"
)

// Job is a unit of work enqueued for a worker.
type Job struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Payload    map[string]any `json:"payload,omitempty"`
	EnqueuedAt time.Time      `json:"enqueued_at"`
}

const (
	// TypeHealth is a no-op job used to verify worker connectivity.
	TypeHealth = "platform.health"
)

// Handler processes a single job. Return nil to ack/remove; non-nil may requeue depending on impl.
type Handler func(ctx context.Context, job Job) error

// Queue is the Platform Infrastructure work-queue port.
type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	// Consume blocks until ctx is cancelled. handler is called for each job.
	Consume(ctx context.Context, handler Handler) error
}
