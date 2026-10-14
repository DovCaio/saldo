package transaction

import (
	"context"
	"time"
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
}