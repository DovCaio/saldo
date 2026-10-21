//go:build integration

package http_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	listtransactions "github.com/DovCaio/saldo/internal/adapter/http/list_transactions"
	"github.com/DovCaio/saldo/internal/adapter/postgres"
	application "github.com/DovCaio/saldo/internal/application/list_transactions"
	"github.com/DovCaio/saldo/internal/transaction"
)

func setupListIntegration(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connecting to PostgreSQL: %v", err)
	}

	repository := postgres.NewTransactionRepository(db)
	service := transaction.NewService(repository)
	useCase := application.NewFindAllTransaction(service)

	mux := http.NewServeMux()
	mux.Handle(
		"GET /transactions",
		listtransactions.NewListTransactionsHandler(useCase),
	)

	return db, mux
}

func TestListTransactionsIntegration_FiltersAndReturnsTransactions(
	t *testing.T,
) {
	db, handler := setupListIntegration(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	description := "integration-list-" + suffix

	// Um valor específico ajuda a isolar o registro deste teste.
	amount := int64(100_000_000 + time.Now().UnixNano()%1_000_000_000)
	date := "2020-01-02"

	_, err := db.Exec(`
		INSERT INTO transactions (
			transaction_date,
			description,
			amount_cents,
			type
		) VALUES ($1, $2, $3, $4)
	`, date, description, amount, string(transaction.Expense))
	if err != nil {
		t.Fatalf("inserting test transaction: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM transactions WHERE description = $1",
			description,
		)
		if err != nil {
			t.Errorf("cleaning up test transaction: %v", err)
		}
	})

	url := fmt.Sprintf(
		"/transactions?type=EXPENSE&date_from=%s&date_to=%s&min_amount=%d&max_amount=%d",
		date,
		date,
		amount,
		amount,
	)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(
		contentType,
		"application/json",
	) {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}

	var got []transaction.Transaction
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	found := false

	for _, item := range got {
		if item.Description == description {
			found = true

			if item.AmountCents != amount {
				t.Errorf(
					"expected amount %d, got %d",
					amount,
					item.AmountCents,
				)
			}

			if item.Type != transaction.Expense {
				t.Errorf(
					"expected type %s, got %s",
					transaction.Expense,
					item.Type,
				)
			}

			if item.Date.Format("2006-01-02") != date {
				t.Errorf(
					"expected date %s, got %s",
					date,
					item.Date.Format("2006-01-02"),
				)
			}
		}
	}

	if !found {
		t.Fatalf(
			"expected transaction %q in response, got %d transactions",
			description,
			len(got),
		)
	}
}