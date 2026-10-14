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
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" && testDatabaseURL == ""{
		log.Fatal("DATABASE_URL or TEST_DATABASE_URL is required")
	}

	url := databaseURL
	if url == "" {
		url = testDatabaseURL
	}
	println(url)
	db, err := sql.Open("pgx", url)
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