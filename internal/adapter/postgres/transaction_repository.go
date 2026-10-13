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