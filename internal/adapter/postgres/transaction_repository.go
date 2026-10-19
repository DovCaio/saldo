package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DovCaio/saldo/internal/adapter/dto"
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

func (r *TransactionRepository) Summary(
    ctx context.Context,
    filter transaction.TransactionSummaryFilter,
) (*dto.SummaryResponse, error) {
    groupBy := filter.GroupBy

    if groupBy == "" {
        groupBy = transaction.MONTH
    }

    switch groupBy {
    case transaction.DAY, transaction.MONTH, transaction.YEAR:
    default:
        return nil, fmt.Errorf(
            "invalid summary grouping: %q",
            groupBy,
        )
    }

    query := `
        SELECT
            DATE_TRUNC($1, transaction_date)::date AS date,
            type,
            SUM(amount_cents) AS total
        FROM transactions
    `

    args := []any{string(groupBy)}
    conditions := make([]string, 0, 2)

    if filter.DateFrom != nil {
        args = append(args, *filter.DateFrom)
        conditions = append(
            conditions,
            fmt.Sprintf("transaction_date >= $%d", len(args)),
        )
    }

    if filter.DateTo != nil {
        args = append(args, *filter.DateTo)
        conditions = append(
            conditions,
            fmt.Sprintf("transaction_date <= $%d", len(args)),
        )
    }

    if len(conditions) > 0 {
        query += " WHERE " + strings.Join(conditions, " AND ")
    }

    query += `
        GROUP BY DATE_TRUNC($1, transaction_date)::date, type
        ORDER BY date, type
    `

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("querying transaction summary: %w", err)
    }
    defer rows.Close()

    response := &dto.SummaryResponse{
        Series: make([]dto.Serie, 0),
    }

    seriesIndex := make(map[string]int)

    for rows.Next() {
        var (
            date  time.Time
            typ   transaction.TransactionType
            total int64
        )

        if err := rows.Scan(&date, &typ, &total); err != nil {
            return nil, fmt.Errorf(
                "scanning transaction summary: %w",
                err,
            )
        }

        var dateKey string

        switch groupBy {
        case transaction.DAY:
            dateKey = date.Format("2006-01-02")
        case transaction.MONTH:
            dateKey = date.Format("2006-01")
        case transaction.YEAR:
            dateKey = date.Format("2006")
        }

        index, exists := seriesIndex[dateKey]
        if !exists {
            index = len(response.Series)
            seriesIndex[dateKey] = index

            response.Series = append(response.Series, dto.Serie{
                Date: dateKey,
            })
        }

        switch typ {
        case transaction.Income:
            response.TotalIncome += total
            response.Series[index].Income += total

        case transaction.Expense:
            response.TotalExpense += total
            response.Series[index].Expense += total

        default:
            return nil, fmt.Errorf(
                "unexpected transaction type in summary: %q",
                typ,
            )
        }
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "iterating transaction summary: %w",
            err,
        )
    }

    response.Balance = response.TotalIncome - response.TotalExpense

    return response, nil
}