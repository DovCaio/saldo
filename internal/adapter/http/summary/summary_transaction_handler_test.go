package summary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DovCaio/saldo/internal/adapter/dto"
	"github.com/DovCaio/saldo/internal/transaction"
)

type summaryUseCaseMock struct{

	result dto.SummaryResponse
	err error
	called bool
	filter transaction.TransactionSummaryFilter

}

func (s *summaryUseCaseMock) Execute(
	ctx context.Context, filter transaction.TransactionSummaryFilter,
) (*dto.SummaryResponse, error){

	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	s.filter = filter
	return &s.result, nil

}

func TestExecuteSucessfulWithoutFilter(t *testing.T){
	expected := dto.SummaryResponse{
		TotalIncome: 1000,
		TotalExpense: 800,
		Balance: 200,
		Series: []dto.Serie{
			{Date: "2099-12-20", Income: 1000, Expense: 800},
		},
	}

	mock := &summaryUseCaseMock{
		called: false,
		result: expected,
	}

	handler := NewSummaryTransactionHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/transactions/summary", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req )


	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code )
	}

	if !mock.called {
		t.Error("use case was not called")
	}

	if mock.filter.DateFrom != nil {
		t.Error("dateFrom must not exists")
	}

	if mock.filter.DateTo != nil {
		t.Error("dateTo must not exists")
	}

		var got dto.SummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&got) ; err != nil {
		t.Fatalf("error deconding the rec value: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v actual: %v", expected, got)
	}


}

func TestExecuteSucessfulWithFilter(t *testing.T){
	expected := dto.SummaryResponse{
		TotalIncome: 1000,
		TotalExpense: 800,
		Balance: 200,
		Series: []dto.Serie{
			{Date: "2099-12-20", Income: 1000, Expense: 800},
		},
	}

	mock := &summaryUseCaseMock{
		called: false,
		result: expected,
	}

	handler := NewSummaryTransactionHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/transactions/summary?date_from=2099-12-01&date_to=2099-12-30&group_by=DAY", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req )


	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code )
	}

	if !mock.called {
		t.Error("use case was not called")
	}

	if mock.filter.DateFrom == nil {
		t.Error("dateFrom must exists")
	}

	if mock.filter.DateTo == nil {
		t.Error("dateTo must exists")
	}

	var got dto.SummaryResponse

	if err := json.NewDecoder(rec.Body).Decode(&got) ; err != nil {
		t.Fatalf("error deconding the rec value: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v actual: %v", expected, got)
	}


}

func TestExecuteWithValidGroupBy(t *testing.T) {
	tests := []struct {
		name    string
		groupBy string
	}{
		{name: "day", groupBy: "DAY"},
		{name: "month", groupBy: "MONTH"},
		{name: "year", groupBy: "YEAR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &summaryUseCaseMock{}
			handler := NewSummaryTransactionHandler(mock)

			req := httptest.NewRequest(
				http.MethodGet,
				"/transactions/summary?group_by="+tt.groupBy,
				nil,
			)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf(
					"expected status %d, got %d: %s",
					http.StatusOK,
					rec.Code,
					rec.Body.String(),
				)
			}

			if !mock.called {
				t.Fatal("use case was not called")
			}

			if mock.filter.GroupBy != transaction.SummaryGroupBy(tt.groupBy) {
				t.Errorf(
					"expected group_by %q, got %q",
					tt.groupBy,
					mock.filter.GroupBy,
				)
			}
		})
	}
}

func TestExecuteWithInvalidGroupBy(t *testing.T) {
	tests := []string{
		"INVALID",
		"WEEK",
		"DAYY",
	}

	for _, groupBy := range tests {
		t.Run(groupBy, func(t *testing.T) {
			mock := &summaryUseCaseMock{}
			handler := NewSummaryTransactionHandler(mock)

			req := httptest.NewRequest(
				http.MethodGet,
				"/transactions/summary?group_by="+groupBy,
				nil,
			)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					rec.Code,
				)
			}

			if mock.called {
				t.Error("use case should not be called for invalid group_by")
			}
		})
	}
}

func TestExecuteWithInvalidDates(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "invalid date_from",
			query: "?date_from=01-12-2099",
		},
		{
			name:  "invalid date_to",
			query: "?date_to=not-a-date",
		},
		{
			name:  "invalid date_from and date_to",
			query: "?date_from=abc&date_to=xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &summaryUseCaseMock{}
			handler := NewSummaryTransactionHandler(mock)

			req := httptest.NewRequest(
				http.MethodGet,
				"/transactions/summary"+tt.query,
				nil,
			)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d: %s",
					http.StatusBadRequest,
					rec.Code,
					rec.Body.String(),
				)
			}

			if mock.called {
				t.Error("use case should not be called with invalid dates")
			}
		})
	}
}

func TestExecuteWithUseCaseError(t *testing.T) {
	mock := &summaryUseCaseMock{
		err: errors.New("database error"),
	}

	handler := NewSummaryTransactionHandler(mock)

	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions/summary",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	if !mock.called {
		t.Error("use case was not called")
	}
}

func TestExecutePassesFiltersToUseCase(t *testing.T) {
	mock := &summaryUseCaseMock{}
	handler := NewSummaryTransactionHandler(mock)

	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions/summary?date_from=2099-12-01&date_to=2099-12-30&group_by=MONTH",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if mock.filter.DateFrom == nil {
		t.Fatal("expected date_from to be set")
	}

	if mock.filter.DateTo == nil {
		t.Fatal("expected date_to to be set")
	}

	expectedFrom := "2099-12-01"
	expectedTo := "2099-12-30"

	if got := mock.filter.DateFrom.Format("2006-01-02"); got != expectedFrom {
		t.Errorf("expected date_from %s, got %s", expectedFrom, got)
	}

	if got := mock.filter.DateTo.Format("2006-01-02"); got != expectedTo {
		t.Errorf("expected date_to %s, got %s", expectedTo, got)
	}

	if mock.filter.GroupBy != transaction.MONTH {
		t.Errorf(
			"expected group_by %q, got %q",
			transaction.MONTH,
			mock.filter.GroupBy,
		)
	}
}