package transaction

import (
	"context"

	"github.com/DovCaio/saldo/internal/adapter/dto"
)

type ServiceTransaction interface {

    ImportTransactions(
    ctx context.Context,
    transactions []Transaction,
    ) error 

    FindAll(ctx context.Context, filter TransactionFilter) ([]dto.TransactionResponse, error)

    Summary(ctx context.Context, filter TransactionSummaryFilter) (*dto.SummaryResponse, error)
}

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

func (s *Service) FindAll(ctx context.Context, filter TransactionFilter) ([]dto.TransactionResponse, error){

    response, err := s.repository.FindAll(ctx, filter)
    if err != nil {
        return nil, err
    }

    result := make([]dto.TransactionResponse, 0, len(response))

    for _, value := range response {
        res, err := dto.NewTransactionResponse(value.ID, value.Date, value.Description, value.AmountCents, string(value.Type))
        if err != nil {
            return nil, err
        }
        result = append(result, *res)
    }

    return result, nil

}

func (s *Service) Summary(ctx context.Context, filter TransactionSummaryFilter) (*dto.SummaryResponse, error) {
    return s.repository.Summary(ctx, filter)
}