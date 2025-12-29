package worker

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/TharunNo1/payflow/internal/provider"
	"github.com/google/uuid"
)

// StartWorker runs the background processor
func StartWorker(ctx context.Context, db *sql.DB, p provider.PaymentProvider) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("👷 Worker started: Polling for pending payouts...")

	for {
		select {
		case <-ctx.Done():
			log.Println("👷 Worker stopping: Received shutdown signal")
			return
		case <-ticker.C:
			processNextTask(ctx, db, p)
		}
	}
}

func processNextTask(ctx context.Context, db *sql.DB, p provider.PaymentProvider) {
	var taskID uuid.UUID
	var amount int64
	var accountName string

	// 1. Claim a task using FOR UPDATE SKIP LOCKED (Standard for High-Concurrency Queues)
	// We join with accounts/entries to get the data needed for the bank call
	query := `
		UPDATE payout_tasks 
		SET status = 'PROCESSING' 
		WHERE id = (
			SELECT pt.id 
			FROM payout_tasks pt
			JOIN entries e ON pt.entry_id = e.transaction_id
			JOIN accounts a ON e.account_id = a.id
			WHERE pt.status = 'PENDING' AND e.amount > 0
			LIMIT 1 
			FOR UPDATE SKIP LOCKED
		) 
		RETURNING id, e.amount, a.owner_name`

	err := db.QueryRowContext(ctx, query).Scan(&taskID, &amount, &accountName)
	if err == sql.ErrNoRows {
		return // Nothing to process
	}
	if err != nil {
		log.Printf("❌ Worker error fetching task: %v", err)
		return
	}

	log.Printf("💸 Processing Payout %s: Sending %d to %s", taskID, amount, accountName)

	// 2. Call the Provider
	err = p.SendPayout(ctx, amount, accountName)

	// 3. Update Status
	if err != nil {
		log.Printf("⚠️ Payout %s failed: %v", taskID, err)
		db.ExecContext(ctx, "UPDATE payout_tasks SET status = 'FAILED', last_error = $1 WHERE id = $2", err.Error(), taskID)
	} else {
		log.Printf("✅ Payout %s completed successfully", taskID)
		db.ExecContext(ctx, "UPDATE payout_tasks SET status = 'COMPLETED' WHERE id = $1", taskID)
	}
}