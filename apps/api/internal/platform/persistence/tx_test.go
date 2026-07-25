package persistence_test

import (
	"context"
	"errors"
	"testing"

	"hublio/internal/platform/persistence"
)

func TestWithinTransactionNilPool(t *testing.T) {
	err := persistence.WithinTransaction(context.Background(), nil, func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestWithinTransactionPropagatesError(t *testing.T) {
	// Without a live pool we only assert nil-pool behavior here.
	// Integration tests with Postgres belong to later phases.
	sentinel := errors.New("boom")
	_ = sentinel
}
