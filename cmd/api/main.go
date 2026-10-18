package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DovCaio/saldo/internal/adapter/csv"
	healthadapter "github.com/DovCaio/saldo/internal/adapter/http/health"
	import_handler_adapter "github.com/DovCaio/saldo/internal/adapter/http/import_handler"
	list_transaction_handler_adapter "github.com/DovCaio/saldo/internal/adapter/http/list_transactions"
	summary_transaction_handler_adapter "github.com/DovCaio/saldo/internal/adapter/http/summary"
	"github.com/DovCaio/saldo/internal/adapter/postgres"
	"github.com/DovCaio/saldo/internal/application"
	"github.com/DovCaio/saldo/internal/transaction"
)

func main() {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL  is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connecting to PostgreSQL: %v", err)
	}
	

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthadapter.NewHealthHandler())

	repository := postgres.NewTransactionRepository(db)
	service := transaction.NewService(repository)
	parser := csv.NewParser()

	importTransactions := application.NewImportTransactions(
		parser,
		service,
	)

	listTransactions := application.NewFindAllTransaction(
		service,
	)

	summary := application.NewSummaryTransactions(
		service,
	)

	importHandler := import_handler_adapter.NewImportHandler(importTransactions)
	listTransactionsHandler := list_transaction_handler_adapter.NewListTransactionsHandler(listTransactions)
	summaryHandler := summary_transaction_handler_adapter.NewSummaryTransactionHandler(summary)
	
	mux.Handle("POST /transactions/import", importHandler)
	mux.Handle("GET /transactions", listTransactionsHandler)
	mux.Handle("GET /transactions/summary", summaryHandler)
	
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("saldo API running on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}	
}