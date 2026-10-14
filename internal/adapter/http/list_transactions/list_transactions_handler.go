package http

import (
	"net/http"

	"github.com/DovCaio/saldo/internal/transaction"
)

type ListTransactionHandler struct {
	service *transaction.Service
}

func NewListTransactionHandler(service *transaction.Service) *ListTransactionHandler {
	return &ListTransactionHandler{
		service : service,
	}
}

func (h *ListTransactionHandler) ServeHTTP ( w http.ResponseWriter,
    r *http.Request,){


	//query := r.URL.Query()

	/*txType := query.Get("type")
	dateFrom := query.Get("dateFrom")
	dateTo := query.Get("dateTo")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		// Tratar o erro de escrita
	}*/

}