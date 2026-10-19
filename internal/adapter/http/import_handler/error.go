package http

import (
	"errors"
	"net/http"

	"github.com/DovCaio/saldo/internal/transaction"
)


func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, transaction.ErrEmptyImport):
			w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, transaction.ErrDuplicateInBatch):
			w.WriteHeader(http.StatusConflict)
	}
}