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
	httpadapter "github.com/DovCaio/saldo/internal/adapter/http"
	"github.com/DovCaio/saldo/internal/adapter/postgres"
	"github.com/DovCaio/saldo/internal/application"
	"github.com/DovCaio/saldo/internal/transaction"
)

func main() {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
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
	mux.HandleFunc("GET /health", httpadapter.NewHealthHandler())

	repository := postgres.NewTransactionRepository(db)
	service := transaction.NewService(repository)
	parser := csv.NewParser()

	importTransactions := application.NewImportTransactions(
		parser,
		service,
	)

	importHandler := httpadapter.NewImportHandler(importTransactions)
	
	mux.Handle("POST /transactions/import", importHandler)
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("saldo API running on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}	
}