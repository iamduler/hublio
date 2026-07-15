package config

import (
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Address  string
	Username string
	Password string
	DB       int
}

func NewRedisClient() *redis.Client {
	config := RedisConfig{
		Address:  env.GetEnv("REDIS_ADDRESS", "localhost:6379"),
		Username: env.GetEnv("REDIS_USERNAME", ""),
		Password: env.GetEnv("REDIS_PASSWORD", ""),
		DB:       env.GetIntEnv("REDIS_DB", 0),
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Username:     config.Username,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()

	if err != nil {
		logging.Log.Fatal().Err(err).Msg("❌ Failed to connect to Redis")
	}

	logging.Log.Info().Msg("🔄 Redis: Connected")

	return client
}
