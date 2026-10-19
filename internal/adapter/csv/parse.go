package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/DovCaio/saldo/internal/transaction"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(
	r io.Reader,
) ([]transaction.Transaction, error) {
	return ParseCSV(r)
}

func ParseCSV(r io.Reader) ([]transaction.Transaction, error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("CSV is empty")
		}
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	expectedHeader := []string{"date", "description", "amount", "type"}

	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header")
	}

	for i, expected := range expectedHeader {
		if strings.TrimSpace(header[i]) != expected {
			return nil, fmt.Errorf("invalid CSV header")
		}
	}

	reader.FieldsPerRecord = len(expectedHeader)

	var transactions []transaction.Transaction

	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV record %d: %w", line, err)
		}

		date, err := time.Parse(
			"2006-01-02",
			strings.TrimSpace(record[0]),
		)
		if err != nil {
			return nil, fmt.Errorf("invalid date on record %d: %w", line, err)
		}

		amountCents, err := parseAmountCents(record[2])
		if err != nil {
			return nil, fmt.Errorf("invalid amount on record %d: %w", line, err)
		}

		tx, err := transaction.NewTransaction(
			date,
			record[1],
			amountCents,
			transaction.TransactionType(strings.TrimSpace(record[3])),
		)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction on record %d: %w", line, err)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func parseAmountCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("amount cannot be empty")
	}

	if strings.HasPrefix(value, "-") {
		return 0, fmt.Errorf("amount cannot be negative")
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
