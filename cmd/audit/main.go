package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/payflow?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🔍 Starting Financial Reconciliation Audit...")

	// 1. Check Global Zero-Sum (Sum of all entries must be 0)
	var globalSum int64
	err = db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM entries").Scan(&globalSum)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Check Account Consistency (Sum of entries for each account == account.balance)
	rows, err := db.Query(`
		SELECT a.id, a.balance, COALESCE(SUM(e.amount), 0) as entry_sum
		FROM accounts a
		LEFT JOIN entries e ON a.id = e.account_id
		GROUP BY a.id, a.balance
		HAVING a.balance != COALESCE(SUM(e.amount), 0)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	mismatches := 0
	for rows.Next() {
		var id string
		var balance, entrySum int64
		rows.Scan(&id, &balance, &entrySum)
		fmt.Printf("❌ MISMATCH found in Account %s: DB Balance: %d, Ledger Sum: %d\n", id, balance, entrySum)
		mismatches++
	}

	fmt.Println("-------------------------------------------")
	if globalSum == 0 && mismatches == 0 {
		fmt.Println("✅ AUDIT PASSED: System is financially consistent.")
	} else {
		fmt.Printf("⚠️ AUDIT FAILED: Global Sum: %d, Mismatches: %d\n", globalSum, mismatches)
	}
}
