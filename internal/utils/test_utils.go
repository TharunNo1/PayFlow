package utils

import (
	"bytes"
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/TharunNo1/payflow/internal/config"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// SetupTestDB initializes a clean Postgres connection and wipes data
func SetupTestDB() *sql.DB {

	dbURL := config.Load().DBURL

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Clean slate for every test run
	_, _ = db.Exec("TRUNCATE TABLE payout_tasks, entries, accounts CASCADE")
	return db
}

// SetupTestRedis initializes a clean Redis connection on DB index 1
func SetupTestRedis() *redis.Client {
	rdsAddr := config.Load().RedisAddr
	rdb := redis.NewClient(&redis.Options{
		Addr: rdsAddr,
		DB:   1, // Use a separate DB for testing
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to test Redis: %v", err)
	}

	rdb.FlushDB(context.Background())
	return rdb
}

// SeedAccount is a helper to create an account for testing
func SeedAccount(db *sql.DB, id uuid.UUID, balance int64) {
	query := `INSERT INTO accounts (id, owner_name, balance) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, id, "Test User", balance)
	if err != nil {
		log.Fatalf("Failed to seed account: %v", err)
	}
}

// PerformRequest is a helper to simulate HTTP calls to your Gin router
func PerformRequest(r http.Handler, method, path, idempotencyKey string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("X-Idempotency-Key", idempotencyKey)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
