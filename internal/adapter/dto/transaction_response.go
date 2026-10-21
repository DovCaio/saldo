package dto

import (
	"time"
)

type TransactionTypeResponse string 

const (
	Income  TransactionTypeResponse = "INCOME"
	Expense TransactionTypeResponse = "EXPENSE"
)

type TransactionResponse struct {
	ID          string
	Date        time.Time
	Description string
	AmountCents int64
	Type        TransactionTypeResponse
}

func NewTransactionResponse(id string, date time.Time, description string, amountCents int64, type_ string) (*TransactionResponse, error) {

	var responseType TransactionTypeResponse

	switch type_ {	
	case "INCOME", "EXPENSE":
		responseType = TransactionTypeResponse(type_)
	default:
		return nil, InvalidResponse
	}

	return &TransactionResponse{
		ID: id, Date: date, Description: description, AmountCents: amountCents, Type: responseType,
	}, nil


}