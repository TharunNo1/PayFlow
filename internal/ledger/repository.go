package ledger

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type TransferParams struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        int64 // In cents
}

func (r *Repository) ExecuteTransfer(ctx context.Context, p TransferParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback. If Commit() is called, this does nothing.
	defer tx.Rollback()

	txID := uuid.New()

	// 1. Record Debit
	if _, err := tx.ExecContext(ctx, 
		"INSERT INTO entries (id, account_id, amount, transaction_id) VALUES ($1, $2, $3, $4)",
		uuid.New(), p.FromAccountID, -p.Amount, txID); err != nil {
		return err
	}

	// 2. Record Credit
	if _, err := tx.ExecContext(ctx, 
		"INSERT INTO entries (id, account_id, amount, transaction_id) VALUES ($1, $2, $3, $4)",
		uuid.New(), p.ToAccountID, p.Amount, txID); err != nil {
		return err
	}

	// 3. Update Balances (Optimistic check: ensure balance doesn't go negative if business rules require)
	// [Optional: Add a check here to verify sufficient funds]

	return tx.Commit()
}