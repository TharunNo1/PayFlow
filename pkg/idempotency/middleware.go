package idempotency

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Middleware prevents the same request from being processed multiple times
func Middleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the Idempotency Key from the Header
		key := c.GetHeader("X-Idempotency-Key")

		// In fintech, we should reject any request missing this key
		if key == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "X-Idempotency-Key header is required for transfers",
			})
			return
		}

		// 2. Check Redis (Atomic SETNX)
		// We use a prefix to avoid collisions with other data in Redis
		redisKey := "idempotency:" + key

		// SetNX returns true if the key was set (new request), false if it exists (duplicate)
		isNew, err := rdb.SetNX(c.Request.Context(), redisKey, "locked", 24*time.Hour).Result()

		if err != nil {
			// If Redis is down, we fail safe by returning an error
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to verify request uniqueness",
			})
			return
		}

		// 3. Handle Duplicate Case
		if !isNew {
			// STOP: This request has been seen before
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": "duplicate request detected; this transaction is already processed or in progress",
			})
			return
		}

		// 4. Continue to the Actual Transfer Handler
		c.Next()
	}
}
