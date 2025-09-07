package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// RedisCache implements the CacheInterface
type RedisCache struct{}

// InitRedis initializes the Redis client
func InitRedis() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvInt("REDIS_DB", 0)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test the connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache() CacheInterface {
	return &RedisCache{}
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}

// Set stores a value in Redis with an expiration time
func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(key string, dest interface{}) error {
	data, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(key string) error {
	return redisClient.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisCache) Exists(key string) (bool, error) {
	result, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Set stores a value in Redis with an expiration time (package function)
func Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from Redis (package function)
func Get(key string, dest interface{}) error {
	data, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a key from Redis (package function)
func Delete(key string) error {
	return redisClient.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis (package function)
func Exists(key string) (bool, error) {
	result, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// getEnv returns the value of the environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt returns the integer value of the environment variable or a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return defaultValue
	}
	return result
}