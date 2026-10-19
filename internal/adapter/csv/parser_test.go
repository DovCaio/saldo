package csv

import (
	"strings"
	"testing"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCount   int
		wantErr     bool
	}{
		{
			name: "parses valid transactions",
			input: `date,description,amount,type
2026-10-01,Salario,5000.00,INCOME
2026-10-02,Mercado,150.50,EXPENSE`,
			wantCount: 2,
		},
		{
			name:    "rejects empty CSV",
			input:   "",
			wantErr: true,
		},
		{
			name: "rejects invalid header",
			input: `date,title,amount,type
2026-10-01,Salario,5000.00,INCOME`,
			wantErr: true,
		},
		{
			name: "rejects invalid date",
			input: `date,description,amount,type
01-10-2026,Salario,5000.00,INCOME`,
			wantErr: true,
		},
		{
			name: "rejects invalid amount",
			input: `date,description,amount,type
2026-10-01,Salario,abc,INCOME`,
			wantErr: true,
		},
		{
			name: "rejects invalid transaction type",
			input: `date,description,amount,type
2026-10-01,Salario,5000.00,TRANSFER`,
			wantErr: true,
		},
		{
			name: "rejects incomplete row",
			input: `date,description,amount,type
2026-10-01,Salario,5000.00`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCSV(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCSV() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if len(got) != tt.wantCount {
				t.Fatalf("got %d transactions, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestParseCSVTransactionFields(t *testing.T) {
	input := `date,description,amount,type
2026-10-01,Salario,5000.00,INCOME`

	got, err := ParseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d transactions, want 1", len(got))
	}

	tx := got[0]

	if !tx.Date.Equal(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected date: %v", tx.Date)
	}

	if tx.Description != "Salario" {
		t.Errorf("got description %q, want %q", tx.Description, "Salario")
	}

	if tx.AmountCents != 500000 {
		t.Errorf("got amount %d cents, want 500000", tx.AmountCents)
	}

	if tx.Type != transaction.Income {
		t.Errorf("got type %q, want %q", tx.Type, transaction.Income)
	}
}



func TestParseAmountCents(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"5000.00", 500000, false},
		{"150.50", 15050, false},
		{"10", 1000, false},
		{"1.5", 150, false},
		{"0.01", 1, false},
		{"1.999", 0, true},
		{"abc", 0, true},
		{"", 0, true},
		{"-10.00", 0, true},
		{"1,50", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseAmountCents(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}

			if err == nil && got != tt.want {
				t.Errorf("got %d cents, want %d", got, tt.want)
			}
		})
	}
}
