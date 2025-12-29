package ledger

import (
	"context"
)

// GlobalAudit checks if the sum of all entries is zero
func (r *Repository) GlobalAudit(ctx context.Context) (int64, error) {
	var totalBalanceSum int64
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM entries").Scan(&totalBalanceSum)
	return totalBalanceSum, err
}
