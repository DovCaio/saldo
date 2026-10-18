package summary

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)


type SummaryTransactionUseCase interface {
	Execute(ctx context.Context, filter transaction.TransactionSummaryFilter) (*dto.SummaryResponse, error)
}

type SummaryTransactionHandler struct {
	useCase SummaryTransactionUseCase
}

func NewSummaryTransactionHandler(useCase SummaryTransactionUseCase) *SummaryTransactionHandler{
	return &SummaryTransactionHandler{
		useCase: useCase,
	}
}

	
func (h *SummaryTransactionHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	query :=  r.URL.Query()

	var filter transaction.TransactionSummaryFilter

	if value := query.Get("date_from"); value != "" {
		dateFrom, err := time.Parse("2006-01-02", value)
		if err != nil {
			http.Error(w, "invalid date", http.StatusBadRequest)
			return
		}
		filter.DateFrom = &dateFrom
	}

	if value := query.Get("date_to"); value != "" {
		dateTo, err := time.Parse("2006-01-02", value)
		if err != nil {
			http.Error(w, "invalid date", http.StatusBadRequest)
			return
		}
		filter.DateTo = &dateTo
	}

	if value := query.Get("group_by"); value != "" {
		groupBy := transaction.SummaryGroupBy(value)
		switch groupBy{
			case transaction.DAY, transaction.MONTH, transaction.YEAR:
			default :
				http.Error(w, "group_by has an invalid value, try: DAY, MONTH or YEAR", http.StatusBadRequest)
				return

		}
	}

	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo)  {
		http.Error(w, "group_by has an invalid value, try: DAY, MONTH or YEAR", http.StatusBadRequest)
		return
	}

	summary, err := h.useCase.Execute(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to summarize", http.StatusInternalServerError)//this fo sure is wrong, we have to take the real error and return

	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		return
	}

}