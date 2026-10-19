package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)

type transactionFinderMock struct {
	gotContext context.Context
	gotFilter  transaction.TransactionFilter
	result     []transaction.Transaction
	err        error
	called     bool
}

func (m *transactionFinderMock)  SaveBatch(
        ctx context.Context,
        transactions []transaction.Transaction,
    ) error {
	return nil
}

func (m *transactionFinderMock) FindAll(
	ctx context.Context,
	filter transaction.TransactionFilter,
) ([]transaction.Transaction, error) {
	m.called = true
	m.gotContext = ctx
	m.gotFilter = filter

	return m.result, m.err
}

func (m *transactionFinderMock) Summary(
    ctx context.Context,
    filter transaction.TransactionSummaryFilter) (*dto.SummaryResponse, error) {
    return nil, nil
}

func TestFindAllTransaction_Execute_Success(t *testing.T) {
	expected := []transaction.Transaction{
		{
			ID:          "transaction-1",
			Description: "Salário",
			AmountCents: 500000,
			Type:        transaction.Income,
		},
	}

	mock := &transactionFinderMock{
		result: expected,
	}

	service := transaction.NewService(mock)
	useCase := NewFindAllTransaction(service)

	ctx := context.Background()
	filter := transaction.TransactionFilter{
		Limit:  10,
		Offset: 0,
	}

	got, err := useCase.Execute(ctx, filter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}

	if !mock.called {
		t.Fatal("expected FindAll to be called")
	}

	if mock.gotContext != ctx {
		t.Error("expected context to be forwarded")
	}

	if !reflect.DeepEqual(mock.gotFilter, filter) {
		t.Errorf("expected filter %v, got %v", filter, mock.gotFilter)
	}
}

func TestFindAllTransaction_Execute_PropagatesError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	mock := &transactionFinderMock{
		err: expectedErr,
	}

	service := transaction.NewService(mock)
	useCase := NewFindAllTransaction(service)

	got, err := useCase.Execute(
		context.Background(),
		transaction.TransactionFilter{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Errorf("expected nil result, got %v", got)
	}
}

func TestFindAllTransaction_Execute_EmptyResult(t *testing.T) {
	mock := &transactionFinderMock{
		result: []transaction.Transaction{},
	}

	service := transaction.NewService(mock)
	useCase := NewFindAllTransaction(service)

	got, err := useCase.Execute(
		context.Background(),
		transaction.TransactionFilter{},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty result, got %d transactions", len(got))
	}
}