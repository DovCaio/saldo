package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	application "github.com/DovCaio/saldo/internal/application/import_transaction"
	"github.com/DovCaio/saldo/internal/transaction"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(
	parser application.BankParser,
	csvReader *csv.Reader,
) ([]transaction.Transaction, error) {
	return ParseCSV(parser, csvReader)
}

func ParseCSV(parser application.BankParser, csvReader *csv.Reader) ([]transaction.Transaction, error) {

	var transactions []transaction.Transaction

	for line := 2; ; line++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV record %d: %w", line, err)
		}

		dateIndex, exists := parser.GetColumnIndex("date")
		if !exists {
			return nil, fmt.Errorf("date column not found in header")
		}

		date, err := time.Parse(
			"02/01/2006",
			strings.TrimSpace(record[dateIndex]),
		)
		if err != nil {
			return nil, fmt.Errorf("invalid date on record %d: %w", line, err)
		}

		amountIndex, exists := parser.GetColumnIndex("amount")
		if !exists {
			return nil, fmt.Errorf("amount column not found in header")
		}

		amountCents, err := parseAmountCents(record[amountIndex])
		if err != nil {
			return nil, fmt.Errorf("invalid amount on record %d: %w", line, err)
		}

		var transactionType transaction.TransactionType
		if record[amountIndex][0] == '-' {
			transactionType = transaction.Expense
		} else {
			transactionType = transaction.Income
		}

		descriptionIndex, exists := parser.GetColumnIndex("description")
		if !exists {
			return nil, fmt.Errorf("description column not found in header")
		}

		externalIDIndex, exists := parser.GetColumnIndex("external_id")
		if !exists {
			return nil, fmt.Errorf("external_id column not found in header")
		}

		externalID := strings.TrimSpace(record[externalIDIndex])

		tx, err := transaction.NewTransaction(
			date,
			record[descriptionIndex],
			amountCents,
			transactionType,
			externalID)

		if err != nil {
			return nil, fmt.Errorf("invalid transaction on record %d: %w", line, err)
		}

		transactions = append(transactions, tx)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in CSV")
	}

	return transactions, nil
}


func parseAmountCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("amount cannot be empty")
	}

	if strings.HasPrefix(value, "-") {
		value = value[1:]
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid decimal amount")
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid whole amount")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) == 0 || len(fraction) > 2 {
			return 0, fmt.Errorf("amount must have at most two decimal places")
		}
	}

	for len(fraction) < 2 {
		fraction += "0"
	}

	cents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal fraction")
	}

	if whole > (int64(^uint64(0)>>1)-cents)/100 {
		return 0, fmt.Errorf("amount is too large")
	}

	return whole*100 + cents, nil
}