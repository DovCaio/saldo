package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
)

type ListTransactionsUseCase interface {
	Execute(
		ctx context.Context,
		filter transaction.TransactionFilter,
	) ([]transaction.Transaction, error)
}

type ListTransactionsHandler struct {
	useCase ListTransactionsUseCase
}

func NewListTransactionsHandler(
	useCase ListTransactionsUseCase,
) *ListTransactionsHandler {
	return &ListTransactionsHandler{
		useCase: useCase,
	}
}

func (h *ListTransactionsHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	filter := transaction.TransactionFilter{
		Limit:  20,
		Offset: 0,
	}

	if value := query.Get("type"); value != "" {
		txType := transaction.TransactionType(value)

		if txType != transaction.Income &&
			txType != transaction.Expense {
			http.Error(w, "invalid transaction type", http.StatusBadRequest)
			return
		}

		filter.Type = &txType
	}

	if value := query.Get("dateFrom"); value != "" {
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			http.Error(w, "invalid dateFrom", http.StatusBadRequest)
			return
		}
		filter.DateFrom = &date
	}

	if value := query.Get("dateTo"); value != "" {
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			http.Error(w, "invalid dateTo", http.StatusBadRequest)
			return
		}
		filter.DateTo = &date
	}

	if value := query.Get("minAmount"); value != "" {
		amount, err := strconv.ParseInt(value, 10, 64)
		if err != nil || amount < 0 {
			http.Error(w, "invalid minAmount", http.StatusBadRequest)
			return
		}
		filter.MinAmount = &amount
	}

	if value := query.Get("maxAmount"); value != "" {
		amount, err := strconv.ParseInt(value, 10, 64)
		if err != nil || amount < 0 {
			http.Error(w, "invalid maxAmount", http.StatusBadRequest)
			return
		}
		filter.MaxAmount = &amount
	}

	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	}

	if value := query.Get("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		filter.Offset = offset
	}

	if filter.DateFrom != nil &&
		filter.DateTo != nil &&
		filter.DateFrom.After(*filter.DateTo) {
		http.Error(w, "dateFrom must be before or equal to dateTo", http.StatusBadRequest)
		return
	}

	if filter.MinAmount != nil &&
		filter.MaxAmount != nil &&
		*filter.MinAmount > *filter.MaxAmount {
		http.Error(w, "minAmount must be less than or equal to maxAmount", http.StatusBadRequest)
		return
	}

	transactions, err := h.useCase.Execute(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to list transactions", http.StatusInternalServerError)//this fo sure is wrong, we have to take the real error and return
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		return
	}
}
