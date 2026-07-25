package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCacheService struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisCacheService(rdb *redis.Client) RedisCacheService {
	return &redisCacheService{
		ctx: context.Background(),
		rdb: rdb,
	}
}

func (s *redisCacheService) Get(key string, dest any) error {
	data, err := s.rdb.Get(s.ctx, key).Result()

	if err == redis.Nil {
		return err
	}

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (s *redisCacheService) Set(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return s.rdb.Set(s.ctx, key, data, ttl).Err()
}

func (s *redisCacheService) Delete(key string) error {
	cursor := uint64(0)

	for {
		keys, nextCursor, err := s.rdb.Scan(s.ctx, cursor, key, 100).Result()

		if err != nil {
			return err
		}

		if len(keys) > 0 {
			s.rdb.Del(s.ctx, keys...)
		}

		cursor = nextCursor

		if nextCursor == 0 {
			break
		}
	}

	return nil
}

func (s *redisCacheService) Exists(key string) (bool, error) {
	count, err := s.rdb.Exists(s.ctx, key).Result()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
