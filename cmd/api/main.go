package main

import (
	"log"
	"net/http"

	httpadapter "github.com/DovCaio/saldo/internal/adapter/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", httpadapter.NewHealthHandler())
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("saldo API running on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}	
}