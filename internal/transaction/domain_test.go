package transaction

import (
	"errors"
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	date := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		date        time.Time
		description string
		amountCents int64
		kind        TransactionType
		wantErr     error
	}{
		{
			name:        "valid expense",
			date:        date,
			description: " Supermercado ",
			amountCents: 2590,
			kind:        Expense,
		},
		{
			name:        "valid income",
			date:        date,
			description: "Salário",
			amountCents: 350000,
			kind:        Income,
		},
		{
			name:        "zero date",
			description: "Compra",
			amountCents: 1000,
			kind:        Expense,
			wantErr:     ErrInvalidDate,
		},
		{
			name:        "empty description",
			date:        date,
			description: "   ",
			amountCents: 1000,
			kind:        Expense,
			wantErr:     ErrEmptyDescription,
		},
		{
			name:        "zero amount",
			date:        date,
			description: "Compra",
			amountCents: 0,
			kind:        Expense,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "negative amount",
			date:        date,
			description: "Compra",
			amountCents: -100,
			kind:        Expense,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "invalid type",
			date:        date,
			description: "Compra",
			amountCents: 1000,
			kind:        "TRANSFER",
			wantErr:     ErrInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTransaction(
				tt.date,
				tt.description,
				tt.amountCents,
				tt.kind,
			)

			if !errors.Is(err, tt.wantErr)	 {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			wantDescription := tt.description
			if tt.name == "valid expense" {
				wantDescription = "Supermercado"
			}

			if got.Description != wantDescription {
				t.Errorf("description = %q, want %q",
					got.Description, wantDescription)
			}
			
			if got.AmountCents != tt.amountCents {
				t.Errorf(
					"amount = %d, want %d",
					got.AmountCents,
					tt.amountCents,
				)
			}

			if got.Type != tt.kind {
				t.Errorf("type = %q, want %q", got.Type, tt.kind)
			}
		})
	}
}
