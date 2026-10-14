package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DovCaio/saldo/internal/transaction"
)

type TransactionRepository struct {
    db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
    return &TransactionRepository{
        db: db,
    }
}

var _ transaction.Repository = (*TransactionRepository)(nil)

func (r *TransactionRepository) SaveBatch(
    ctx context.Context,
    transactions []transaction.Transaction,
) error {
    if len(transactions) == 0 {
        return transaction.ErrEmptyImport
    }

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("beginning transaction batch: %w", err)
    }
    defer tx.Rollback()

    query := `
        INSERT INTO transactions (
            transaction_date,
            description,
            amount_cents,
            type,
            external_id
        )
        VALUES ($1, $2, $3, $4, NULLIF($5, ''))
    `

    for i, item := range transactions {
        _, err := tx.ExecContext(
            ctx,
            query,
            item.Date.Format("2006-01-02"),
            item.Description,
            item.AmountCents,
            string(item.Type),
            item.ExternalID,
        )
        if err != nil {
            return fmt.Errorf(
                "saving transaction %d: %w",
                i+1,
                err,
            )
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("committing transaction batch: %w", err)
    }

    return nil
}


func (r *TransactionRepository) FindAll(
	ctx context.Context,
	filter transaction.TransactionFilter,
) ([]transaction.Transaction, error) {
	const (
		defaultLimit = 20
		maxLimit     = 100
	)

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id::text,
			transaction_date,
			description,
			amount_cents,
			type,
			COALESCE(external_id, '')
		FROM transactions
		WHERE 1 = 1
	`

	args := make([]any, 0, 7)

	if filter.DateFrom != nil {
		args = append(args, *filter.DateFrom)
		query += fmt.Sprintf(
			" AND transaction_date >= $%d",
			len(args),
		)
	}

	if filter.DateTo != nil {
		args = append(args, *filter.DateTo)
		query += fmt.Sprintf(
			" AND transaction_date <= $%d",
			len(args),
		)
	}

	if filter.MinAmount != nil {
		args = append(args, *filter.MinAmount)
		query += fmt.Sprintf(
			" AND amount_cents >= $%d",
			len(args),
		)
	}

	if filter.MaxAmount != nil {
		args = append(args, *filter.MaxAmount)
		query += fmt.Sprintf(
			" AND amount_cents <= $%d",
			len(args),
		)
	}

	if filter.Type != nil {
		args = append(args, string(*filter.Type))
		query += fmt.Sprintf(
			" AND type = $%d",
			len(args),
		)
	}

	query += " ORDER BY transaction_date DESC, id DESC"

	args = append(args, limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	args = append(args, offset)
	query += fmt.Sprintf(" OFFSET $%d", len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]transaction.Transaction, 0)

	for rows.Next() {
		var item transaction.Transaction
		var txType string

		err := rows.Scan(
			&item.ID,
			&item.Date,
			&item.Description,
			&item.AmountCents,
			&txType,
			&item.ExternalID,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction: %w", err)
		}

		item.Type = transaction.TransactionType(txType)
		transactions = append(transactions, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transactions: %w", err)
	}

	return transactions, nil
}
