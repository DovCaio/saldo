package application

import (
	"context"

	"github.com/DovCaio/saldo/internal/transaction"
)

type ListTransacion interface {
	FindAll(filter transaction.TransactionFilter) (transaction.Transaction, error)
}

type FindAllTransaction struct {
	service *transaction.Service
}

func NewFindAllTransaction(service *transaction.Service) *FindAllTransaction {
	return &FindAllTransaction{
		service: service,
	}

}


func (f *FindAllTransaction) Execute(ctx context.Context,
	filter transaction.TransactionFilter) ([]transaction.Transaction, error) {

		return f.service.FindAll(ctx, filter)

}
