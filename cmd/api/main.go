package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TharunNo1/payflow/internal/config"
	"github.com/TharunNo1/payflow/internal/ledger"
	"github.com/TharunNo1/payflow/internal/utils"
	"github.com/TharunNo1/payflow/pkg/idempotency"
	"github.com/gin-gonic/gin"     // web framework
	_ "github.com/lib/pq"          // Postgres driver
	"github.com/redis/go-redis/v9" // Redis client
)

func SetupRouter(repo *ledger.Repository, rdb *redis.Client) *gin.Engine {

	// Gin Router with Logger and Recovery Middleware
	r := gin.Default()

	// Routes
	r.GET("/health", func(c *gin.Context) {
		sum, err := repo.GlobalAudit(c.Request.Context())
		if err != nil {
			c.JSON(500, utils.Error(err.Error(), "Ledger integrity check failed"))
			return
		}
		c.JSON(200, utils.Success(
			gin.H{
				"system":               "healthy",
				"ledger_integrity_sum": sum,
			},
			"System is Healthy"),
		)
	})

	// We pass idempotency.Middleware(rdb) as an argument BEFORE the handler function
	r.POST("/transfer", idempotency.Middleware(rdb), func(c *gin.Context) {
		var req ledger.TransferParams
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, utils.Error(err.Error(), "invalid request body"))
			return
		}
		if err := repo.ExecuteTransfer(c.Request.Context(), req); err != nil {
			c.JSON(500, utils.Error(err.Error(), ""))
			return
		}
		c.JSON(200, utils.Success(
			gin.H{
				"message": "transfer successful",
			}, ""),
		)
	})

	return r
}

func main() {

	// 1. Get Configurations and Environment variables
	cfg := config.Load()

	// 2. Database Connection
	db, err := sql.Open("postgres", cfg.DBURL)
	// Test DB connectivity
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	} else {
		log.Println("✅ Connected to Postgres")
	}
	defer db.Close()

	// 2. Redis Connection
	redisAddr := cfg.RedisAddr
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	// Test Redis connectivity
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// 3. Initialize Repository & Router

	// Initialize ledger Repository with DB connection
	repo := ledger.NewRepository(db)

	// Get Router
	r := SetupRouter(repo, rdb)
	// 4. Configure HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 5. Start Server in a Goroutine
	go func() {
		fmt.Println("🚀 PayFlow API starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 6. Graceful Shutdown Logic
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Close Redis connections on shutdown
	if err := rdb.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
