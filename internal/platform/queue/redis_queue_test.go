package queue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"hublio/internal/platform/queue"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func TestRedisQueueEnqueueConsumeHealth(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger := zerolog.Nop()
	q := queue.NewRedisQueue(rdb, &logger)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var (
		mu      sync.Mutex
		gotType string
	)

	done := make(chan struct{})
	go func() {
		_ = q.Consume(ctx, func(ctx context.Context, job queue.Job) error {
			mu.Lock()
			gotType = job.Type
			mu.Unlock()
			close(done)
			cancel()
			return nil
		})
	}()

	if err := queue.EnqueueHealth(t.Context(), q); err != nil {
		t.Fatalf("EnqueueHealth: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for job")
	}

	mu.Lock()
	defer mu.Unlock()
	if gotType != queue.TypeHealth {
		t.Fatalf("got type %q want %q", gotType, queue.TypeHealth)
	}
}
