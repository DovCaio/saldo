//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openTestDB(t *testing.T) (*sql.DB, *TransactionRepository) {
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

	return db, NewTransactionRepository(db)
}

func TestTransactionRepository_SaveBatch_PersistsTransactions(t *testing.T) {
	db, repository := openTestDB(t)

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
	db, repository := openTestDB(t)

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

func insertTransaction(
	t *testing.T,
	db *sql.DB,
	date string,
	description string,
	amount int64,
	txType transaction.TransactionType,
) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO transactions (
			transaction_date,
			description,
			amount_cents,
			type
		) VALUES ($1, $2, $3, $4)
	`, date, description, amount, string(txType))

	if err != nil {
		t.Fatalf("inserting transaction: %v", err)
	}
}

func DeleteAll(t *testing.T, db *sql.DB) {
	t.Cleanup(func() {
		_, err := db.ExecContext(
			context.Background(),
			"DELETE FROM transactions",
		)
		if err != nil {
			t.Errorf("cleaning up test transaction: %v", err)
		}
	})
}

func TestFindAll_FiltersByType(t *testing.T) {
	db, repository := openTestDB(t)
	ctx := context.Background()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	incomeDescription := "findall-income-" + suffix
	expenseDescription := "findall-expense-" + suffix

	DeleteAll(t, db)

	insertTransaction(t, db, "2026-10-01", incomeDescription, 5000, transaction.Income)
	insertTransaction(t, db, "2026-10-02", expenseDescription, 2500, transaction.Expense)

	txType := transaction.Income

	got, err := repository.FindAll(ctx, transaction.TransactionFilter{
		Type:  &txType,
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}

	found := false
	for _, item := range got {
		if item.Description == incomeDescription {
			found = true
			if item.Type != transaction.Income {
				t.Errorf("expected INCOME, got %s", item.Type)
			}
		}

		if item.Description == expenseDescription {
			t.Errorf("expense should not match INCOME filter")
		}
	}

	if !found {
		t.Errorf("expected to find transaction %q", incomeDescription)
	}
}

func TestFindAll_FiltersByAmount(t *testing.T) {
	db, repository := openTestDB(t)
	ctx := context.Background()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	lowDescription := "findall-low-" + suffix
	highDescription := "findall-high-" + suffix

	DeleteAll(t, db)

	insertTransaction(t, db, "2026-10-01", lowDescription, 1000, transaction.Expense)
	insertTransaction(t, db, "2026-10-02", highDescription, 5000, transaction.Expense)

	minAmount := int64(2000)
	maxAmount := int64(6000)

	got, err := repository.FindAll(ctx, transaction.TransactionFilter{
		MinAmount: &minAmount,
		MaxAmount: &maxAmount,
		Limit:     100,
	})
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}

	for _, item := range got {
		if item.Description == lowDescription {
			t.Errorf("transaction below minimum amount should not match")
		}
	}

	foundHigh := false
	for _, item := range got {
		if item.Description == highDescription {
			foundHigh = true
		}
	}
	if !foundHigh {
		t.Errorf("expected to find transaction %q", highDescription)
	}
}

func TestFindAll_FiltersByDateRange(t *testing.T) {
	db, repository := openTestDB(t)
	ctx := context.Background()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	before := "findall-before-" + suffix
	inside := "findall-inside-" + suffix
	after := "findall-after-" + suffix

	DeleteAll(t, db)


	insertTransaction(t, db, "2026-09-30", before, 1000, transaction.Expense)
	insertTransaction(t, db, "2026-10-05", inside, 2000, transaction.Expense)
	insertTransaction(t, db, "2026-10-20", after, 3000, transaction.Expense)

	dateFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)

	got, err := repository.FindAll(ctx, transaction.TransactionFilter{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}

	foundInside := false
	for _, item := range got {
		switch item.Description {
		case before:
			t.Errorf("transaction before date range should not match")
		case after:
			t.Errorf("transaction after date range should not match")
		case inside:
			foundInside = true
		}
	}

	if !foundInside {
		t.Errorf("expected to find transaction %q", inside)
	}
}

func TestFindAll_AppliesPagination(t *testing.T) {
	db, repository := openTestDB(t)
	ctx := context.Background()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	first := "findall-page-1-" + suffix
	second := "findall-page-2-" + suffix
	third := "findall-page-3-" + suffix

	DeleteAll(t, db)

	insertTransaction(t, db, "2026-10-01", first, 1000, transaction.Expense)
	insertTransaction(t, db, "2026-10-02", second, 2000, transaction.Expense)
	insertTransaction(t, db, "2026-10-03", third, 3000, transaction.Expense)

	minAmount := int64(1000)
	maxAmount := int64(3000)

	got, err := repository.FindAll(ctx, transaction.TransactionFilter{
		MinAmount: &minAmount,
		MaxAmount: &maxAmount,
		Limit:     2,
		Offset:    1,
	})
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(got))
	}

	if got[0].Description != second {
		t.Errorf("expected %q first, got %q", second, got[0].Description)
	}
	if got[1].Description != first {
		t.Errorf("expected %q second, got %q", first, got[1].Description)
	}
}

func TestFindAll_ReturnsEmptySliceWhenNoMatches(t *testing.T) {
	_, repository := openTestDB(t)

	txType := transaction.Income
	impossibleAmount := int64(999999999)

	got, err := repository.FindAll(context.Background(), transaction.TransactionFilter{
		Type:      &txType,
		MinAmount: &impossibleAmount,
		Limit:     20,
	})
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected no transactions, got %d", len(got))
	}
}

func insertSummaryFixtures(t *testing.T, db *sql.DB, prefix string) {
	t.Helper()

	ctx := context.Background()

	fixtures := []struct {
		date        string
		typ         transaction.TransactionType
		amountCents int64
		externalID  string
	}{
		{"2099-01-05", transaction.Income, 10000, prefix + "-1"},
		{"2099-01-05", transaction.Expense, 2500, prefix + "-2"},
		{"2099-01-20", transaction.Expense, 1000, prefix + "-3"},
		{"2099-02-03", transaction.Income, 5000, prefix + "-4"},
		{"2099-02-03", transaction.Expense, 3000, prefix + "-5"},
	}
	//insertTransaction(t, db, "2026-10-03", third, 3000, transaction.Expense)

	for _, fixture := range fixtures {
		_, err := db.ExecContext(ctx, `
			INSERT INTO transactions (
				transaction_date,
				type,
				amount_cents,
				description,
				external_id
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
			fixture.date,
			fixture.typ,
			fixture.amountCents,
			"Integration test "+fixture.externalID,
			fixture.externalID,
		)
		if err != nil {
			t.Fatalf("inserting fixture %s: %v", fixture.externalID, err)
		}
	}

	t.Cleanup(func() {
		_, err := db.ExecContext(ctx,
			`DELETE FROM transactions WHERE external_id LIKE $1`,
			prefix+"-%",
		)
		if err != nil {
			t.Errorf("cleaning summary fixtures: %v", err)
		}
	})
}

func datePtr(value string) *time.Time {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return &date
}

func assertSummaryTotals(
	t *testing.T,
	got *dto.SummaryResponse,
	income, expense, balance int64,
) {
	t.Helper()

	if got.TotalIncome != income {
		t.Errorf("TotalIncome = %d, want %d", got.TotalIncome, income)
	}
	if got.TotalExpense != expense {
		t.Errorf("TotalExpense = %d, want %d", got.TotalExpense, expense)
	}
	if got.Balance != balance {
		t.Errorf("Balance = %d, want %d", got.Balance, balance)
	}
}

func TestTransactionRepository_Summary(t *testing.T) {
	db, repository := openTestDB(t)
	prefix := fmt.Sprintf("summary-test-%d", time.Now().UnixNano())
	insertSummaryFixtures(t, db, prefix)

	ctx := context.Background()
	from := datePtr("2099-01-01")
	to := datePtr("2099-02-28")

	t.Run("monthly grouping", func(t *testing.T) {
		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			DateFrom: from,
			DateTo:   to,
			GroupBy:  transaction.MONTH,
		})
		if err != nil {
			t.Fatalf("Summary() error = %v", err)
		}

		assertSummaryTotals(t, got, 15000, 6500, 8500)

		want := []dto.Serie{
			{Date: "2099-01", Income: 10000, Expense: 3500},
			{Date: "2099-02", Income: 5000, Expense: 3000},
		}

		if !reflect.DeepEqual(got.Series, want) {
			t.Errorf("Series = %#v, want %#v", got.Series, want)
		}
	})

	t.Run("daily grouping", func(t *testing.T) {
		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			DateFrom: from,
			DateTo:   to,
			GroupBy:  transaction.DAY,
		})
		if err != nil {
			t.Fatalf("Summary() error = %v", err)
		}

		assertSummaryTotals(t, got, 15000, 6500, 8500)

		want := []dto.Serie{
			{Date: "2099-01-05", Income: 10000, Expense: 2500},
			{Date: "2099-01-20", Income: 0, Expense: 1000},
			{Date: "2099-02-03", Income: 5000, Expense: 3000},
		}

		if !reflect.DeepEqual(got.Series, want) {
			t.Errorf("Series = %#v, want %#v", got.Series, want)
		}
	})

	t.Run("yearly grouping", func(t *testing.T) {
		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			DateFrom: from,
			DateTo:   to,
			GroupBy:  transaction.YEAR,
		})
		if err != nil {
			t.Fatalf("Summary() error = %v", err)
		}

		assertSummaryTotals(t, got, 15000, 6500, 8500)

		want := []dto.Serie{
			{Date: "2099", Income: 15000, Expense: 6500},
		}

		if !reflect.DeepEqual(got.Series, want) {
			t.Errorf("Series = %#v, want %#v", got.Series, want)
		}
	})

	t.Run("date filters", func(t *testing.T) {
		februaryFrom := datePtr("2099-02-01")
		februaryTo := datePtr("2099-02-28")

		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			DateFrom: februaryFrom,
			DateTo:   februaryTo,
			GroupBy:  transaction.MONTH,
		})
		if err != nil {
			t.Fatalf("Summary() error = %v", err)
		}

		assertSummaryTotals(t, got, 5000, 3000, 2000)

		if len(got.Series) != 1 || got.Series[0].Date != "2099-02" {
			t.Errorf("unexpected February series: %#v", got.Series)
		}
	})

	t.Run("empty period", func(t *testing.T) {
		marchFrom := datePtr("2099-03-01")
		marchTo := datePtr("2099-03-31")

		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			DateFrom: marchFrom,
			DateTo:   marchTo,
			GroupBy:  transaction.MONTH,
		})
		if err != nil {
			t.Fatalf("Summary() error = %v", err)
		}

		assertSummaryTotals(t, got, 0, 0, 0)

		if len(got.Series) != 0 {
			t.Errorf("Series length = %d, want 0", len(got.Series))
		}
	})

	t.Run("invalid grouping", func(t *testing.T) {
		got, err := repository.Summary(ctx, transaction.TransactionSummaryFilter{
			GroupBy: transaction.SummaryGroupBy("week"),
		})

		if err == nil {
			t.Fatal("expected error for invalid grouping")
		}
		if got != nil {
			t.Errorf("response = %#v, want nil", got)
		}
	})
}