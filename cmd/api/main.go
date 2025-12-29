package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/TharunNo1/payflow/internal/ledger"
)

func main() {
	// 1. Connection String - In industry, use Environment Variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/payflow?sslmode=disable"
	}

	// 2. Connect to Database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// 3. Initialize Repository
	repo := ledger.NewRepository(db)

	// 4. Setup Router (Using Gin for speed in a 10hr sprint)
	r := gin.Default()

	// Health Check & Audit Result
	r.GET("/health", func(c *gin.Context) {
		sum, err := repo.GlobalAudit(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"status": "healthy",
			"ledger_integrity_sum": sum, // Should be 0
		})
	})

	// Transfer Endpoint
	r.POST("/transfer", func(c *gin.Context) {
		var req ledger.TransferParams
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request body"})
			return
		}

		err := repo.ExecuteTransfer(c.Request.Context(), req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "transfer successful"})
	})

	fmt.Println("PayFlow API running on :8080")
	r.Run(":8080")
}

quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit // Wait for Ctrl+C

log.Println("Shutting down gracefully...")
// Close DB, stop worker ticker, etc.