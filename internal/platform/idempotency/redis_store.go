package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{
		rdb:    rdb,
		prefix: "idempotency:",
	}
}

func (s *RedisStore) key(k string) string {
	return s.prefix + k
}

func (s *RedisStore) Reserve(ctx context.Context, key, requestHash string, ttl time.Duration) error {
	rec := Record{
		Key:         key,
		RequestHash: requestHash,
		CreatedAt:   time.Now().UTC(),
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("idempotency: marshal: %w", err)
	}

	ok, err := s.rdb.SetNX(ctx, s.key(key), raw, ttl).Result()
	if err != nil {
		return fmt.Errorf("idempotency: setnx: %w", err)
	}
	if !ok {
		existing, getErr := s.Get(ctx, key)
		if getErr != nil {
			return ErrConflict
		}
		if existing.RequestHash != requestHash {
			return ErrConflict
		}
		return nil
	}
	return nil
}

func (s *RedisStore) Get(ctx context.Context, key string) (Record, error) {
	raw, err := s.rdb.Get(ctx, s.key(key)).Bytes()
	if err == redis.Nil {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("idempotency: get: %w", err)
	}

	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return Record{}, fmt.Errorf("idempotency: unmarshal: %w", err)
	}
	return rec, nil
}

func (s *RedisStore) SaveResponse(ctx context.Context, key, responseRef string, ttl time.Duration) error {
	rec, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	rec.ResponseRef = responseRef
	raw, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("idempotency: marshal: %w", err)
	}
	if err := s.rdb.Set(ctx, s.key(key), raw, ttl).Err(); err != nil {
		return fmt.Errorf("idempotency: save response: %w", err)
	}
	return nil
}
