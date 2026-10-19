package application

import (
	"context"
	"io"

	"github.com/DovCaio/saldo/internal/transaction"
)

type TransactionParser interface {
	Parse(r io.Reader) ([]transaction.Transaction, error)
}

type ImportTransactions struct {
	parser  TransactionParser
	service *transaction.Service
}

func NewImportTransactions(
	parser TransactionParser,
	service *transaction.Service,
) *ImportTransactions {
	return &ImportTransactions{
		parser:  parser,
		service: service,
	}
}

func (uc *ImportTransactions) Execute(
	ctx context.Context,
	r io.Reader,
) error {
	transactions, err := uc.parser.Parse(r)
	if err != nil {
		return err
	}

	return uc.service.ImportTransactions(ctx, transactions)
}
