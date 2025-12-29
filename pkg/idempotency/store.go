package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CheckIdempotency attempts to claim a unique key in Redis.
// It returns true if the key was successfully set (first time seeing this request).
// It returns false if the key already exists (request is a duplicate).
func CheckIdempotency(ctx context.Context, redisClient *redis.Client, key string) (bool, error) {
	// SETNX (Set if Not eXists) is atomic.
	// We set a TTL (Time To Live) of 24 hours so the Redis memory stays clean.
	return redisClient.SetNX(ctx, "idempotency:"+key, "processing", 24*time.Hour).Result()
}
