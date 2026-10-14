package application

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
)

type fakeParser struct {
	transactions []transaction.Transaction
	err          error
	called       bool
}

func (f *fakeParser) Parse(
	r io.Reader,
) ([]transaction.Transaction, error) {
	f.called = true

	if f.err != nil {
		return nil, f.err
	}

	return f.transactions, nil
}

type fakeRepository struct {
	saved []transaction.Transaction
	err   error
	calls int
}

func (f *fakeRepository) SaveBatch(
	ctx context.Context,
	transactions []transaction.Transaction,
) error {
	f.calls++

	if f.err != nil {
		return f.err
	}

	f.saved = append(f.saved, transactions...)
	return nil
}

func (f *fakeRepository) FindAll(
	ctx context.Context,
	filter transaction.TransactionFilter,
) ([]transaction.Transaction, error){
	return nil, nil
}

func TestImportTransactionsExecute(t *testing.T) {
	tests := []struct {
		name         string
		transactions []transaction.Transaction
		parserError  error
		repositoryError error
		wantError    bool
		wantSaved    int
		wantRepoCalls int
	}{
		{
			name: "imports parsed transactions",
			transactions: []transaction.Transaction{
				{
					Date:        time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Description: "Salario",
					AmountCents: 500000,
					Type:        transaction.Income,
				},
			},
			wantSaved:     1,
			wantRepoCalls: 1,
		},
		{
			name:        "propagates parser error",
			parserError: errors.New("invalid CSV"),
			wantError:   true,
		},
		{
			name: "rejects empty import",
			wantError: true,
		},
		{
			name: "propagates repository error",
			transactions: []transaction.Transaction{
				{
					Date:        time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Description: "Salario",
					AmountCents: 500000,
					Type:        transaction.Income,
				},
			},
			repositoryError: errors.New("database unavailable"),
			wantError:       true,
			wantRepoCalls:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &fakeParser{
				transactions: tt.transactions,
				err:          tt.parserError,
			}

			repo := &fakeRepository{
				err: tt.repositoryError,
			}

			service := transaction.NewService(repo)
			useCase := NewImportTransactions(parser, service)

			err := useCase.Execute(
				context.Background(),
				nil,
			)

			if (err != nil) != tt.wantError {
				t.Fatalf("Execute() error = %v, wantError = %v",
					err, tt.wantError)
			}

			if !parser.called {
				t.Error("expected parser to be called")
			}

			if len(repo.saved) != tt.wantSaved {
				t.Errorf("got %d saved transactions, want %d",
					len(repo.saved), tt.wantSaved)
			}

			if repo.calls != tt.wantRepoCalls {
				t.Errorf("got %d repository calls, want %d",
					repo.calls, tt.wantRepoCalls)
			}

			if tt.wantSaved > 0 &&
				!reflect.DeepEqual(repo.saved, tt.transactions) {
				t.Error("saved transactions differ from parsed transactions")
			}
		})
	}
}
