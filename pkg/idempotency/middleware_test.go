package idempotency

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TharunNo1/payflow/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIdempotencyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rdb := utils.SetupTestRedis() // Helper returning *redis.Client

	r := gin.New()
	r.POST("/transfer", Middleware(rdb), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	t.Run("Blocks Duplicate Keys", func(t *testing.T) {
		key := "tx-unique-123"

		// 1. First Request
		req1, _ := http.NewRequest("POST", "/transfer", nil)
		req1.Header.Set("X-Idempotency-Key", key)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 2. Duplicate Request
		req2, _ := http.NewRequest("POST", "/transfer", nil)
		req2.Header.Set("X-Idempotency-Key", key)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
		assert.Contains(t, w2.Body.String(), "duplicate request")
	})
}
