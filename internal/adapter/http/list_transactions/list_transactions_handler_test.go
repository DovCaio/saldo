package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
)

type listTransactionsUseCaseMock struct {
	transactions []transaction.Transaction
	err          error
	gotFilter    transaction.TransactionFilter
	called       bool
}

func (m *listTransactionsUseCaseMock) Execute(
	ctx context.Context,
	filter transaction.TransactionFilter,
) ([]transaction.Transaction, error) {
	m.called = true
	m.gotFilter = filter
	return m.transactions, m.err
}

func TestListTransactionsHandler_Success(t *testing.T) {
	expected := []transaction.Transaction{
		{
			ID:          "transaction-1",
			Date:        time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
			Description: "Salário",
			AmountCents: 500000,
			Type:        transaction.Income,
		},
	}

	mock := &listTransactionsUseCaseMock{
		transactions: expected,
	}

	handler := NewListTransactionsHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected application/json, got %q", contentType)
	}

	

	var got []transaction.Transaction
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %+v, got %+v", expected, got)
	}

	if !mock.called {
		t.Fatal("expected use case to be called")
	}
}

func TestListTransactionsHandler_ParsesFilters(t *testing.T) {
	mock := &listTransactionsUseCaseMock{}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions?type=EXPENSE&minAmount=1000&maxAmount=5000&dateFrom=2026-10-01&dateTo=2026-10-14&limit=10&offset=5",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	filter := mock.gotFilter

	if filter.Type == nil || *filter.Type != transaction.Expense {
		t.Errorf("expected type EXPENSE, got %v", filter.Type)
	}

	if filter.MinAmount == nil || *filter.MinAmount != 1000 {
		t.Errorf("expected minAmount 1000, got %v", filter.MinAmount)
	}

	if filter.MaxAmount == nil || *filter.MaxAmount != 5000 {
		t.Errorf("expected maxAmount 5000, got %v", filter.MaxAmount)
	}

	if filter.DateFrom == nil || filter.DateFrom.Format("2006-01-02") != "2026-10-01" {
		t.Errorf("unexpected dateFrom: %v", filter.DateFrom)
	}

	if filter.DateTo == nil || filter.DateTo.Format("2006-01-02") != "2026-10-14" {
		t.Errorf("unexpected dateTo: %v", filter.DateTo)
	}

	if filter.Limit != 10 {
		t.Errorf("expected limit 10, got %d", filter.Limit)
	}

	if filter.Offset != 5 {
		t.Errorf("expected offset 5, got %d", filter.Offset)
	}
}

func TestListTransactionsHandler_InvalidType(t *testing.T) {
	mock := &listTransactionsUseCaseMock{}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/transactions?type=UNKNOWN", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if mock.called {
		t.Error("use case should not be called for invalid filters")
	}
}

func TestListTransactionsHandler_InvalidAmount(t *testing.T) {
	mock := &listTransactionsUseCaseMock{}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/transactions?minAmount=abc", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if mock.called {
		t.Error("use case should not be called for invalid filters")
	}
}

func TestListTransactionsHandler_InvalidDateRange(t *testing.T) {
	mock := &listTransactionsUseCaseMock{}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions?dateFrom=2026-10-15&dateTo=2026-10-01",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if mock.called {
		t.Error("use case should not be called for invalid filters")
	}
}

func TestListTransactionsHandler_InvalidAmountRange(t *testing.T) {
	mock := &listTransactionsUseCaseMock{}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions?minAmount=5000&maxAmount=1000",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if mock.called {
		t.Error("use case should not be called for invalid filters")
	}
}

func TestListTransactionsHandler_UseCaseError(t *testing.T) {
	mock := &listTransactionsUseCaseMock{
		err: errors.New("database unavailable"),
	}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestListTransactionsHandler_EmptyResult(t *testing.T) {
	mock := &listTransactionsUseCaseMock{
		transactions: []transaction.Transaction{},
	}
	handler := NewListTransactionsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []transaction.Transaction
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty result, got %d transactions", len(got))
	}
}