package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const defaultQueueKey = "hublio:workqueue"

// RedisQueue is a list-based Redis work queue (RPUSH / BLPOP).
type RedisQueue struct {
	rdb      *redis.Client
	key      string
	logger   *zerolog.Logger
	blockFor time.Duration
}

func NewRedisQueue(rdb *redis.Client, logger *zerolog.Logger) *RedisQueue {
	return &RedisQueue{
		rdb:      rdb,
		key:      defaultQueueKey,
		logger:   logger,
		blockFor: 5 * time.Second,
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now().UTC()
	}
	if job.Type == "" {
		return fmt.Errorf("queue: job type required")
	}

	raw, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("queue: marshal job: %w", err)
	}

	if err := q.rdb.RPush(ctx, q.key, raw).Err(); err != nil {
		return fmt.Errorf("queue: enqueue: %w", err)
	}

	if q.logger != nil {
		q.logger.Info().
			Str("job_id", job.ID).
			Str("job_type", job.Type).
			Msg("job enqueued")
	}
	return nil
}

func (q *RedisQueue) Consume(ctx context.Context, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("queue: nil handler")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := q.rdb.BLPop(ctx, q.blockFor, q.key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if q.logger != nil {
				q.logger.Error().Err(err).Msg("queue: blpop failed")
			}
			time.Sleep(time.Second)
			continue
		}

		if len(result) < 2 {
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			if q.logger != nil {
				q.logger.Error().Err(err).Msg("queue: invalid job payload")
			}
			continue
		}

		if err := handler(ctx, job); err != nil {
			if q.logger != nil {
				q.logger.Error().
					Err(err).
					Str("job_id", job.ID).
					Str("job_type", job.Type).
					Msg("queue: job handler failed; job dropped (no retry in v1 platform queue)")
			}
			continue
		}

		if q.logger != nil {
			q.logger.Info().
				Str("job_id", job.ID).
				Str("job_type", job.Type).
				Msg("queue: job processed")
		}
	}
}

// Depth returns the number of jobs currently waiting in the Redis list (LLEN).
func (q *RedisQueue) Depth(ctx context.Context) (int64, error) {
	depth, err := q.rdb.LLen(ctx, q.key).Result()
	if err != nil {
		return 0, fmt.Errorf("queue: llen: %w", err)
	}
	return depth, nil
}

// EnqueueHealth enqueues a platform.health no-op job.
func EnqueueHealth(ctx context.Context, q Queue) error {
	return q.Enqueue(ctx, Job{
		Type:    TypeHealth,
		Payload: map[string]any{"source": "platform"},
	})
}

// EnqueueExecution enqueues an orchestration.execution job for the worker to run.
func EnqueueExecution(ctx context.Context, q Queue, payload map[string]any) error {
	return q.Enqueue(ctx, Job{
		Type:    TypeExecution,
		Payload: payload,
	})
}
