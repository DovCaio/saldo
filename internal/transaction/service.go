package transaction

import "context"

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