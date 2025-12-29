func (r *Repository) GlobalAudit(ctx context.Context) (int64, error) {
	var totalBalanceSum int64
	// In a perfect ledger, the sum of all debits and credits is ALWAYS 0.
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM entries").Scan(&totalBalanceSum)
	return totalBalanceSum, err
}