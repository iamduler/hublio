package config

import (
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/logger"
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
		Address:  utils.GetEnv("REDIS_ADDRESS", "localhost:6379"),
		Username: utils.GetEnv("REDIS_USERNAME", ""),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       utils.GetIntEnv("REDIS_DB", 0),
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
		logger.Log.Fatal().Err(err).Msg("❌ Failed to connect to Redis")
	}

	logger.Log.Info().Msg("🔄 Redis: Connected")

	return client
}
