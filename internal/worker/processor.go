func StartWorker(db *sql.DB, provider PaymentProvider) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		// 1. Find a pending task
		var taskID uuid.UUID
		err := db.QueryRow("UPDATE payout_tasks SET status = 'PROCESSING' WHERE id = (SELECT id FROM payout_tasks WHERE status = 'PENDING' LIMIT 1) RETURNING id").Scan(&taskID)
		if err == sql.ErrNoRows {
			continue // No work to do
		}

		// 2. Call the Mock Provider
		err = provider.SendPayment(...)
		
		// 3. Update status
		if err != nil {
			db.Exec("UPDATE payout_tasks SET status = 'FAILED', last_error = $1 WHERE id = $2", err.Error(), taskID)
		} else {
			db.Exec("UPDATE payout_tasks SET status = 'COMPLETED' WHERE id = $1", taskID)
		}
	}
}