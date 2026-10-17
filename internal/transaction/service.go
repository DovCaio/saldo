package transaction

import (
	"context"

	"github.com/DovCaio/saldo/internal/adapter/dto"
)

type Service struct {
    repository Repository
}

func NewService(repository Repository) *Service {
    return &Service{
        repository: repository,
    }
}

func (s *Service) Register(
    ctx context.Context,
    tx Transaction,
) error {
    return s.repository.SaveBatch(ctx, []Transaction{tx})
}

func (s *Service) ImportTransactions(
    ctx context.Context,
    transactions []Transaction,
) error {
    if len(transactions) == 0 {
        return ErrEmptyImport
    }

    return s.repository.SaveBatch(ctx, transactions)
}

func (s *Service) FindAll(ctx context.Context, filter TransactionFilter) ([]Transaction, error){

    return s.repository.FindAll(ctx, filter)

}

func (s *Service) Summary(ctx context.Context, filter TransactionSummaryFilter) (*dto.SummaryResponse, error) {
    return s.repository.Summary(ctx, filter)
}