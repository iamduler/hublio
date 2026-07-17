// Package idempotency provides a skeleton store for Intent Idempotency-Key handling.
// Intent routes wire this in Phase D; Phase A only defines the contract and Redis impl.
package idempotency

import (
	"context"
	"errors"
	"time"
)

var (
	ErrConflict = errors.New("idempotency: key already exists")
	ErrNotFound = errors.New("idempotency: key not found")
)

// Record holds the cached outcome of an idempotent request.
type Record struct {
	Key        string
	RequestHash string
	ResponseRef string
	CreatedAt  time.Time
}

// Store persists idempotency keys with TTL semantics.
type Store interface {
	// Reserve tries to claim key for requestHash. Returns ErrConflict if already reserved with a different hash.
	Reserve(ctx context.Context, key, requestHash string, ttl time.Duration) error
	Get(ctx context.Context, key string) (Record, error)
	SaveResponse(ctx context.Context, key, responseRef string, ttl time.Duration) error
}
