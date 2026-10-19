package http

import (
	"errors"
	"net/http"

	"github.com/DovCaio/saldo/internal/transaction"
)


func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, transaction.ErrEmptyImport), errors.Is(err, transaction.ErrInvalidDate), errors.Is(err, transaction.ErrEmptyDescription), errors.Is(err, transaction.ErrInvalidAmount), errors.Is(err, transaction.ErrInvalidType):
			w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, transaction.ErrDuplicateInBatch):
			w.WriteHeader(http.StatusConflict)
	}
}