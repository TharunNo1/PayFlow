package ledger

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type TransferParams struct {
	FromAccountID uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id" binding:"required"`
	Amount        int64     `json:"amount" binding:"required,gt=0"`
}

// ExecuteTransfer handles the atomic movement of funds and creates a payout task
func (r *Repository) ExecuteTransfer(ctx context.Context, p TransferParams) error {
	// Start a database transaction
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback if anything fails
	defer tx.Rollback()

	// 1. Validation: Ensure sender has enough funds (Locking the row for update)
	var balance int64
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM accounts WHERE id = $1 FOR UPDATE",
		p.FromAccountID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("failed to fetch sender account: %w", err)
	}

	if balance < p.Amount {
		return fmt.Errorf("insufficient funds: available %d, requested %d", balance, p.Amount)
	}

	transactionID := uuid.New()

	// 2. Insert Entries (Immutable Ledger)
	// Debit Sender
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO entries (id, account_id, amount, transaction_id) VALUES ($1, $2, $3, $4)",
		uuid.New(), p.FromAccountID, -p.Amount, transactionID); err != nil {
		return fmt.Errorf("failed to record debit entry: %w", err)
	}

	// Credit Receiver
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO entries (id, account_id, amount, transaction_id) VALUES ($1, $2, $3, $4)",
		uuid.New(), p.ToAccountID, p.Amount, transactionID); err != nil {
		return fmt.Errorf("failed to record credit entry: %w", err)
	}

	// 3. Update Denormalized Balances
	if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", p.Amount, p.FromAccountID); err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", p.Amount, p.ToAccountID); err != nil {
		return fmt.Errorf("failed to update receiver balance: %w", err)
	}

	// 4. THE OUTBOX: Queue the task for the background worker
	// This ensures the worker only sees this task if the transaction commits successfully.
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO payout_tasks (id, entry_id, status) VALUES ($1, $2, $3)",
		uuid.New(), transactionID, "PENDING"); err != nil {
		return fmt.Errorf("failed to queue payout task: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
