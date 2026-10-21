package application

import (
	"context"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)

type ListTransationUseCase interface {
	Execute(ctx context.Context, filter transaction.TransactionFilter) ([]dto.TransactionResponse, error)
}

type FindAllTransaction struct {
	service transaction.ServiceTransaction
}

func NewFindAllTransaction(service transaction.ServiceTransaction) *FindAllTransaction {
	return &FindAllTransaction{
		service: service,
	}

}

func (f *FindAllTransaction) Execute(ctx context.Context,
	filter transaction.TransactionFilter) ([]dto.TransactionResponse, error) {

		return f.service.FindAll(ctx, filter)

}
