package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/DovCaio/saldo/internal/adapter/bank_parsers/nubank"
	"github.com/DovCaio/saldo/internal/transaction"
)

type ImportTransactionsUseCase interface {
    Execute(ctx context.Context, r io.Reader) error
}


type TransactionParser interface {
	Parse(csvReader *csv.Reader) ([]transaction.Transaction, error)
}

type ImportTransactions struct {
	parser  TransactionParser
	service *transaction.Service
}

type BankParser interface {

	Parse() error
	Supports() bool
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

	content, err := csvReader.ReadAll()

	nubankParser, err := nubank.NewNubankParserIMPL(&content)
	if err != nil {
		return err
	}

	bankParsers :=  []BankParser{
		nubankParser,
	};


	for i := range bankParsers {
		if bankParsers[i].Supports(){
			bankParsers[i].Parse()
		}
	}

	fmt.Printf("Parsed header: %v\n", content[0])

	VerifyHeader(csvReader)


	transactions, err := uc.parser.Parse(csvReader)
	if err != nil {
		return err
	}

	fmt.Println("Parsed transactions:", transactions)

	return uc.service.ImportTransactions(ctx, transactions)
}

func VerifyHeader(reader *csv.Reader) error {
	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("CSV is empty")
		}
		return fmt.Errorf("reading CSV header: %w", err)
	}

	expectedHeader := []string{"date", "description", "amount", "type"}

	if len(header) != len(expectedHeader) {
		return fmt.Errorf("invalid CSV header: expected %v, got %v", expectedHeader, header)
	}

	for i, expected := range expectedHeader {
		if strings.TrimSpace(header[i]) != expected {
			return fmt.Errorf("invalid CSV header: expected %v, got %v", expectedHeader, header)
		}
	}

	reader.FieldsPerRecord = len(expectedHeader)
	return nil
}
