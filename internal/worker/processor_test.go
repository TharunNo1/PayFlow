package worker

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TharunNo1/payflow/cmd/api"
	"github.com/TharunNo1/payflow/internal/ledger"
	"github.com/TharunNo1/payflow/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFullTransferFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use our central helpers to get real connections
	db := utils.SetupTestDB()
	rdb := utils.SetupTestRedis()
	repo := ledger.NewRepository(db)

	fromID := "00000000-0000-0000-0000-000000000001"
	toID := "00000000-0000-0000-0000-000000000002"

	// Convert string to UUID for the seed helper
	u1, _ := uuid.Parse(fromID)
	u2, _ := uuid.Parse(toID)

	utils.SeedAccount(db, u1, 10000)
	utils.SeedAccount(db, u2, 0)

	// Setup real router with all dependencies and env setup
	router := api.SetupRouter(repo, rdb)

	payload := map[string]interface{}{
		"from_account_id": "00000000-0000-0000-0000-000000000001",
		"to_account_id":   "00000000-0000-0000-0000-000000000002",
		"amount":          1000,
	}

	body, _ := json.Marshal(payload)

	t.Run("End-to-End Transfer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/transfer", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Idempotency-Key", "e2e-test-key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify Global Audit via Health Check
		reqHealth, _ := http.NewRequest("GET", "/health", nil)
		wHealth := httptest.NewRecorder()
		router.ServeHTTP(wHealth, reqHealth)

		assert.Contains(t, wHealth.Body.String(), `"ledger_integrity_sum":0`)
	})
}
