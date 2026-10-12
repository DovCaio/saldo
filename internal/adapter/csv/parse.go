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

func ParseCSV(r io.Reader) ([]transaction.Transaction, error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("CSV is empty")
		}
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	expectedHeader := []string{
		"date",
		"description",
		"amount",
		"type",
	}

	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header")
	}

	for i := range expectedHeader {
		if strings.TrimSpace(header[i]) != expectedHeader[i] {
			return nil, fmt.Errorf("invalid CSV header")
		}
	}

	var transactions []transaction.Transaction

	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV line %d: %w", line, err)
		}

		date, err := time.Parse("2006-01-02", strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid date on line %d: %w", line, err)
		}

		description := strings.TrimSpace(record[1])

		amount, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount on line %d: %w", line, err)
		}

		amountCents := int64(amount * 100)

		transactionType := transaction.TransactionType(
			strings.TrimSpace(record[3]),
		)

		tx, err := transaction.NewTransaction(
			date,
			description,
			amountCents,
			transactionType,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction on line %d: %w", line, err)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}
