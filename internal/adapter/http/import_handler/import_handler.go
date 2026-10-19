package http

import (
	"errors"
	"log"
	"net/http"

	import_transaction_application "github.com/DovCaio/saldo/internal/application/import_transaction"
)



type ImportHandler struct {
    useCase import_transaction_application.ImportTransactionsUseCase
}

func NewImportHandler(useCase import_transaction_application.ImportTransactionsUseCase) *ImportHandler {
    return &ImportHandler{
        useCase: useCase,
    }
}

const maxUploadSize int64 = 5 << 20 // 5 MiB, this must be insert by an env

func (h *ImportHandler) ServeHTTP(
    w http.ResponseWriter,
    r *http.Request,
) {
    r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

    if err := r.ParseMultipartForm(1 << 20); err != nil {
        var maxErr *http.MaxBytesError

        if errors.As(err, &maxErr) {
            http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
            return
        }

        http.Error(w, "invalid multipart request", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "file is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    if err := h.useCase.Execute(r.Context(), file); err != nil {
        writeError(w, err)
        log.Printf("import transactions failed: %+v", err)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}