package ledger

import (
	"context"
	"testing"

	"github.com/TharunNo1/payflow/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepository_ExecuteTransfer(t *testing.T) {
	db := utils.SetupTestDB()
	repo := NewRepository(db)

	t.Run("Successful Transfer", func(t *testing.T) {
		fromID, toID := uuid.New(), uuid.New()
		utils.SeedAccount(db, fromID, 10000) // $100.00
		utils.SeedAccount(db, toID, 0)

		params := TransferParams{
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        5000, // $50.00
		}

		err := repo.ExecuteTransfer(context.Background(), params)
		assert.NoError(t, err)

		// Verify Balances
		var fromBal, toBal int64
		db.QueryRow("SELECT balance FROM accounts WHERE id = $1", fromID).Scan(&fromBal)
		db.QueryRow("SELECT balance FROM accounts WHERE id = $1", toID).Scan(&toBal)

		assert.Equal(t, int64(5000), fromBal)
		assert.Equal(t, int64(5000), toBal)
	})

	t.Run("Insufficient Funds", func(t *testing.T) {
		fromID, toID := uuid.New(), uuid.New()
		utils.SeedAccount(db, fromID, 1000)
		utils.SeedAccount(db, toID, 0)

		err := repo.ExecuteTransfer(context.Background(), TransferParams{
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        5000,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})
}
