package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)

var _ transaction.ServiceTransaction

type transactionFinderMock struct {
	gotContext context.Context
	gotFilter  transaction.TransactionFilter
	result     []dto.TransactionResponse
	err        error
	called     bool
}

func (m *transactionFinderMock) ImportTransactions(ctx context.Context, transactions []transaction.Transaction) error {
	return nil
}

func (m *transactionFinderMock) FindAll(
	ctx context.Context,
	filter transaction.TransactionFilter,
) ([]dto.TransactionResponse, error) {
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
	expected := []dto.TransactionResponse{
		{
			ID:          "transaction-1",
			Description: "Salário",
			AmountCents: 500000,
			Type:        dto.Income,
		},
	}

	mock := &transactionFinderMock{
		result: expected,
	}

	useCase := NewFindAllTransaction(mock)

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

	useCase := NewFindAllTransaction(mock)

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
		result: []dto.TransactionResponse{},
	}

	useCase := NewFindAllTransaction(mock)

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