package transaction

import (
	"errors"
	"strings"
	"time"
)


type TransactionType string

const (
	Income  TransactionType = "INCOME"
	Expense TransactionType = "EXPENSE"
)

type TransactionGroupBy string



var (
	ErrInvalidDate        = errors.New("invalid transaction date")
	ErrEmptyDescription   = errors.New("description cannot be empty")
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
	ErrInvalidType        = errors.New("invalid transaction type")
)

type Transaction struct {
	ID          string
	Date        time.Time
	Description string
	AmountCents int64
	Type        TransactionType
	ExternalID  string
}

func NewTransaction(
	date time.Time,
	description string,
	amountCents int64,
	transactionType TransactionType,
) (Transaction, error) {
	if date.IsZero() {
		return Transaction{}, ErrInvalidDate
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return Transaction{}, ErrEmptyDescription
	}
	if amountCents <= 0 {
		return Transaction{}, ErrInvalidAmount
	}

	if transactionType != Income && transactionType != Expense {
		return Transaction{}, ErrInvalidType
	}

	return Transaction{
		Date:        date,
		Description: description,
		AmountCents: amountCents,
		Type:        transactionType,
	}, nil
}