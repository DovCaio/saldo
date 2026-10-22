//go:build integration

package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	csv_adapter "github.com/DovCaio/saldo/internal/adapter/csv"
	import_http "github.com/DovCaio/saldo/internal/adapter/http/import_handler"
	"github.com/DovCaio/saldo/internal/adapter/postgres"
	application "github.com/DovCaio/saldo/internal/application/import_transaction"
	"github.com/DovCaio/saldo/internal/transaction"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupImportIntegration(t *testing.T) (*sql.DB, http.Handler) {
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
		if err := db.Close(); err != nil {
			t.Errorf("closing test database: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}

	repository := postgres.NewTransactionRepository(db)
	service := transaction.NewService(repository)
	parser := csv_adapter.NewParser()
	useCase := application.NewImportTransactions(parser, service)
	handler := import_http.NewImportHandler(useCase)

	mux := http.NewServeMux()
	mux.Handle("POST /transactions/import", handler)

	return db, mux
}

func createImportRequest(
	t *testing.T,
	csvContent string,
) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	file, err := writer.CreateFormFile("file", "transactions.csv")
	if err != nil {
		t.Fatalf("creating multipart file: %v", err)
	}

	if _, err := file.Write([]byte(csvContent)); err != nil {
		t.Fatalf("writing CSV content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/transactions/import",
		&body,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestImportIntegration_PersistsTransactions(t *testing.T) {
	db, handler := setupImportIntegration(t)

	token := fmt.Sprintf("%d", time.Now().UnixNano())
	description1 := "integration-test-" + token + "-salary"
	description2 := "integration-test-" + token + "-market"

	t.Cleanup(func() {
		_, err := db.Exec(
			`DELETE FROM transactions
			 WHERE description IN ($1, $2)`,
			description1,
			description2,
		)
		if err != nil {
			t.Errorf("cleaning up integration transactions: %v", err)
		}
	})

	csvContent := fmt.Sprintf(
		"date,description,amount,type\n"+
			"2026-10-01,%s,5000.00,INCOME\n"+
			"2026-10-02,%s,150.50,EXPENSE\n",
		description1,
		description2,
	)

	req := createImportRequest(t, csvContent)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d; response: %s",
			recorder.Code,
			http.StatusNoContent,
			strings.TrimSpace(recorder.Body.String()),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT description, amount_cents, type
		FROM transactions
		WHERE description IN ($1, $2)
		ORDER BY description
	`, description1, description2)
	if err != nil {
		t.Fatalf("querying imported transactions: %v", err)
	}
	defer rows.Close()

	type persistedTransaction struct {
		description string
		amount      int64
		txType      string
	}

	var got []persistedTransaction

	for rows.Next() {
		var item persistedTransaction
		if err := rows.Scan(
			&item.description,
			&item.amount,
			&item.txType,
		); err != nil {
			t.Fatalf("scanning imported transaction: %v", err)
		}

		got = append(got, item)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterating imported transactions: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("persisted %d transactions, want 2", len(got))
	}

	expected := map[string]struct {
		amount int64
		txType string
	}{
		description1: {amount: 500000, txType: "INCOME"},
		description2: {amount: 15050, txType: "EXPENSE"},
	}

	for _, item := range got {
		want, ok := expected[item.description]
		if !ok {
			t.Errorf("unexpected description %q", item.description)
			continue
		}

		if item.amount != want.amount {
			t.Errorf(
				"%s: amount = %d, want %d",
				item.description,
				item.amount,
				want.amount,
			)
		}

		if item.txType != want.txType {
			t.Errorf(
				"%s: type = %q, want %q",
				item.description,
				item.txType,
				want.txType,
			)
		}
	}
}
