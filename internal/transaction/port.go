package transaction

import "context"

type Repository interface {
    SaveBatch(ctx context.Context, transactions []Transaction) error
}
