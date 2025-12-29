package main

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

	"github.com/TharunNo1/payflow/internal/ledger"
	"github.com/TharunNo1/payflow/pkg/idempotency" // Ensure this path is correct
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9" // New Redis import
)

func main() {
	// 1. Database Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/payflow?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// 2. Redis Connection (New Section)
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Quick check to ensure Redis is alive before starting
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// 3. Initialize Repository & Router
	repo := ledger.NewRepository(db)
	r := gin.Default()

	// Routes
	r.GET("/health", func(c *gin.Context) {
		sum, err := repo.GlobalAudit(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy", "ledger_integrity_sum": sum})
	})

	// We pass idempotency.Middleware(rdb) as an argument BEFORE the handler function
	r.POST("/transfer", idempotency.Middleware(rdb), func(c *gin.Context) {
		var req ledger.TransferParams
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request body"})
			return
		}
		if err := repo.ExecuteTransfer(c.Request.Context(), req); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "transfer successful"})
	})

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
