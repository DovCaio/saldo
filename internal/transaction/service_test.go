package transaction

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/adapter/dto"
)

type fakeRepository struct {
    saved  []Transaction
    err    error
	called bool
	summaryResult *dto.SummaryResponse
}

func (f *fakeRepository) SaveBatch(
    ctx context.Context,
    tx []Transaction,
) error {
    if f.err != nil {
        return f.err
    }

	f.called = true

    f.saved = append(f.saved, tx...)
    return nil
}

func (f *fakeRepository) FindAll(
    ctx context.Context,
    filter TransactionFilter,
) ([]Transaction, error) {
	f.called = true

    return nil, nil
}

func (f *fakeRepository) Summary(
    ctx context.Context,
    filter TransactionSummaryFilter) (*dto.SummaryResponse, error) {
	f.called = true

	if f.err != nil {
		return nil, f.err
	}

    return f.summaryResult, nil
}

func TestServiceRegister(t *testing.T) {
    repository := &fakeRepository{called: false}
    service := NewService(repository)

    tx, err := NewTransaction(
        time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC),
        "Supermercado",
        8590,
        Expense,
    )
    if err != nil {
        t.Fatal(err)
    }

    err = service.Register(context.Background(), tx)
    if err != nil {
        t.Fatalf("Register() error = %v", err)
    }

    if len(repository.saved) != 1 {
        t.Fatalf(
            "saved transactions = %d, want 1",
            len(repository.saved),
        )
    }

    if repository.saved[0].AmountCents != 8590 {
        t.Errorf("saved amount = %d, want 8590",
            repository.saved[0].AmountCents)
    }
}

func TestServiceRegisterReturnsRepositoryError(t *testing.T) {
    expectedErr := errors.New("database unavailable")

    repository := &fakeRepository{err: expectedErr}
    service := NewService(repository)

    tx, err := NewTransaction(
        time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC),
        "Supermercado",
        8590,
        Expense,
    )
    if err != nil {
        t.Fatal(err)
    }

    err = service.Register(context.Background(), tx)

    if !errors.Is(err, expectedErr) {
        t.Fatalf("error = %v, want %v", err, expectedErr)
    }
}

func TestImportTransactions(t *testing.T) {
	tests := []struct {
		name            string
		transactions    []Transaction
		repositoryError error
		wantError       error
		wantSaved       int
	}{
		{
			name: "imports transactions successfully",
			transactions: []Transaction{
				{
					Date:        time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Description: "Salário",
					AmountCents: 300000,
					Type:        Income,
				},
				{
					Date:        time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC),
					Description: "Mercado",
					AmountCents: 15000,
					Type:        Expense,
				},
			},
			wantSaved: 2,
		},
		{
			name:       "rejects empty import",
			wantError:  ErrEmptyImport,
			wantSaved:  0,
		},
		{
			name: "propagates repository error",
			transactions: []Transaction{
				{
					Date:        time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
					Description: "Salário",
					AmountCents: 300000,
					Type:        Income,
				},
			},
			repositoryError: errors.New("database unavailable"),
			wantError:       nil,
			wantSaved:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				err: tt.repositoryError,
			}
			service := NewService(repo)

			err := service.ImportTransactions(
				context.Background(),
				tt.transactions,
			)

			if tt.name == "propagates repository error" {
				if err == nil || err.Error() != tt.repositoryError.Error() {
					t.Fatalf("expected repository error %v, got %v",
						tt.repositoryError, err)
				}
			} else if !errors.Is(err, tt.wantError) {
				t.Fatalf("expected error %v, got %v", tt.wantError, err)
			}

			if len(repo.saved) != tt.wantSaved {
				t.Errorf("expected %d saved transactions, got %d",
					tt.wantSaved, len(repo.saved))
			}

			if tt.wantSaved > 0 &&
				!reflect.DeepEqual(repo.saved, tt.transactions) {
				t.Errorf("saved transactions differ from imported transactions")
			}
		})
	}
}


func TestSumarySucessful(t *testing.T){
	expectedResult := &dto.SummaryResponse{
			TotalIncome: 1000,
			TotalExpense: 200,
			Balance: 800,
			Series: []dto.Serie{
				{Date: "2026-10-02", Income: 1000, Expense: 200},
			},
	}
	
	repository := &fakeRepository{
		called: false,
		summaryResult: expectedResult,
	}
	service := NewService(repository)

	dateFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

	filter := TransactionSummaryFilter{
		DateFrom: &dateFrom,
		DateTo: &dateTo,
		GroupBy: "Day",
	}

	result, err := service.Summary(context.Background(), filter)

	if (err != nil) {
		t.Fatalf("some error ocurred")
	}

	if !repository.called {
		t.Fatalf("repository wasnt called")
	}

 	if !reflect.DeepEqual(result, expectedResult) {
        t.Errorf("expected %#v, got %#v", expectedResult, result)
    }
}

func TestSumaryFail(t *testing.T){

	repository := &fakeRepository{
		err:  errors.New("error mocking"),
	}

	dateFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)

	filter := TransactionSummaryFilter{
		DateFrom: &dateFrom,
		DateTo: &dateTo,
		GroupBy: "Day",
	}

	service := NewService(repository)

	result, err := service.Summary(context.Background(), filter)

	if result != nil {
       t.Errorf("result must not be returned when an error ocurred")		
	}

	if err == nil {
		t.Errorf("error must be returned when an error ocurred")		
	}

}