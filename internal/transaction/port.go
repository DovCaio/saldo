package transaction

import (
	"context"
	"time"

	"github.com/DovCaio/saldo/internal/adapter/dto"
)

type TransactionFilter struct {
    DateFrom  *time.Time
    DateTo    *time.Time
    MinAmount *int64
    MaxAmount *int64
    Type      *TransactionType
    Limit     int
    Offset    int
}

type Repository interface {
    SaveBatch(
        ctx context.Context,
        transactions []Transaction,
    ) error

    FindAll(
        ctx context.Context,
        filter TransactionFilter,
    ) ([]Transaction, error)

    Summary(
    ctx context.Context,
    filter TransactionSummaryFilter,
    ) (*dto.SummaryResponse, error)
}