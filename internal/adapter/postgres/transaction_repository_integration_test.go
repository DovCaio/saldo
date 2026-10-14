//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}

	return db
}

func TestTransactionRepository_SaveBatch_PersistsTransactions(t *testing.T) {
	db := openTestDB(t)
	repository := NewTransactionRepository(db)

	ctx := context.Background()
	externalID := fmt.Sprintf("integration-%d", time.Now().UnixNano())

	t.Cleanup(func() {
		_, err := db.ExecContext(
			context.Background(),
			"DELETE FROM transactions WHERE external_id = $1",
			externalID,
		)
		if err != nil {
			t.Errorf("cleaning up test transaction: %v", err)
		}
	})

	date := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	tx := transaction.Transaction{
		Date:        date,
		Description: "Integration test transaction",
		AmountCents: 12550,
		Type:        transaction.Expense,
		ExternalID:  externalID,
	}

	if err := repository.SaveBatch(ctx, []transaction.Transaction{tx}); err != nil {
		t.Fatalf("SaveBatch() error: %v", err)
	}

	var (
		gotDate        time.Time
		gotDescription string
		gotAmount      int64
		gotType        string
		gotExternalID  sql.NullString
	)

	err := db.QueryRowContext(ctx, `
		SELECT transaction_date, description, amount_cents, type, external_id
		FROM transactions
		WHERE external_id = $1
	`, externalID).Scan(
		&gotDate,
		&gotDescription,
		&gotAmount,
		&gotType,
		&gotExternalID,
	)
	
	if err != nil {
		t.Fatalf("querying persisted transaction: %v", err)
	}

	if !gotDate.Equal(date) {
		t.Errorf("date = %v, want %v", gotDate, date)
	}
	if gotDescription != tx.Description {
		t.Errorf("description = %q, want %q", gotDescription, tx.Description)
	}
	if gotAmount != tx.AmountCents {
		t.Errorf("amount = %d, want %d", gotAmount, tx.AmountCents)
	}
	if gotType != string(tx.Type) {
		t.Errorf("type = %q, want %q", gotType, tx.Type)
	}
	if !gotExternalID.Valid || gotExternalID.String != externalID {
		t.Errorf("external_id = %v, want %q", gotExternalID, externalID)
	}
}

func TestTransactionRepository_SaveBatch_RollsBackOnError(t *testing.T) {
	db := openTestDB(t)
	repository := NewTransactionRepository(db)

	ctx := context.Background()
	externalID := fmt.Sprintf("rollback-test-%d", time.Now().UnixNano())

	t.Cleanup(func() {
		_, err := db.ExecContext(
			context.Background(),
			"DELETE FROM transactions WHERE external_id = $1",
			externalID,
		)
		if err != nil {
			t.Errorf("cleaning up test transaction: %v", err)
		}
	})

	date := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	batch := []transaction.Transaction{
		{
			Date:        date,
			Description: "Should be rolled back",
			AmountCents: 1000,
			Type:        transaction.Expense,
			ExternalID:  externalID,
		},
		{
			Date:        date,
			Description: "Invalid transaction type",
			AmountCents: 2000,
			Type:        transaction.TransactionType("INVALID"),
			ExternalID:  externalID + "-invalid",
		},
	}

	err := repository.SaveBatch(ctx, batch)
	if err == nil {
		t.Fatal("SaveBatch() error = nil, want database constraint error")
	}

	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM transactions
		WHERE external_id IN ($1, $2)
	`, externalID, externalID+"-invalid").Scan(&count)
	if err != nil {
		t.Fatalf("checking rollback: %v", err)
	}

	if count != 0 {
		t.Errorf("found %d persisted transactions, want 0 after rollback", count)
	}
}
