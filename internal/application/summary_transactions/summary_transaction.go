package application

import (
	"context"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)

type Summary interface {
	Execute(ctx context.Context, filter transaction.TransactionSummaryFilter) (*dto.SummaryResponse, error)
}

type SummaryTransactions struct {
	service transaction.ServiceTransaction
}

func NewSummaryTransactions(service transaction.ServiceTransaction) *SummaryTransactions {
	return &SummaryTransactions{
		service: service,
	}
}

func (st SummaryTransactions) Execute(ctx context.Context, filter transaction.TransactionSummaryFilter) (*dto.SummaryResponse, error){

	return st.service.Summary(ctx, filter)

}