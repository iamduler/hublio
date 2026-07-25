package idempotency_test

import (
	"testing"
	"time"

	"hublio/internal/platform/idempotency"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisStoreReserveAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := idempotency.NewRedisStore(rdb)

	ctx := t.Context()
	if err := store.Reserve(ctx, "key-1", "hash-a", time.Minute); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	// Same hash is idempotent.
	if err := store.Reserve(ctx, "key-1", "hash-a", time.Minute); err != nil {
		t.Fatalf("Reserve same hash: %v", err)
	}

	if err := store.Reserve(ctx, "key-1", "hash-b", time.Minute); err != idempotency.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}

	rec, err := store.Get(ctx, "key-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.RequestHash != "hash-a" {
		t.Fatalf("unexpected hash %q", rec.RequestHash)
	}

	if err := store.SaveResponse(ctx, "key-1", "resp-ref", time.Minute); err != nil {
		t.Fatalf("SaveResponse: %v", err)
	}

	rec, err = store.Get(ctx, "key-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.ResponseRef != "resp-ref" {
		t.Fatalf("unexpected response ref %q", rec.ResponseRef)
	}
}
