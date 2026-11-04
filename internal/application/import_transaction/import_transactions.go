package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/DovCaio/saldo/internal/adapter/bank_parsers/nubank"
	"github.com/DovCaio/saldo/internal/transaction"
)

type ImportTransactionsUseCase interface {
    Execute(ctx context.Context, r io.Reader) error
}


type TransactionParser interface {
	Parse(parser BankParser, csvReader *csv.Reader) ([]transaction.Transaction, error)
}

type ImportTransactions struct {
	parser  TransactionParser
	service *transaction.Service
}

type BankParser interface {

	Supports() bool
	GetColumnIndex(columnName string) (int, bool)
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

	csvReader := csv.NewReader(r)

	header, err := csvReader.Read()

	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("CSV is empty")
		}
		return fmt.Errorf("reading CSV header: %w", err)
	}

	nubankParser, err := nubank.NewNubankParserIMPL(&header)
	if err != nil {
		return err
	}

	bankParsers :=  []BankParser{
		nubankParser,
	};

	var rigthBankParser BankParser

	for i := range bankParsers {
		if bankParsers[i].Supports(){
			rigthBankParser = bankParsers[i]
			break
		}
	}

	transactions, err := uc.parser.Parse(rigthBankParser, csvReader)
	if err != nil {
		return err
	}


	return uc.service.ImportTransactions(ctx, transactions)
}