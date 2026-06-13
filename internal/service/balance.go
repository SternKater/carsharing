package service

import (
	"context"
	"errors"

	"github.com/SternKater/carsharing/internal/domain"
)

type BalanceRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
}

type BalanceService struct {
	repo BalanceRepositoryInterface
}

func NewBalanceService(r BalanceRepositoryInterface) *BalanceService {
	return &BalanceService{repo: r}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID int64) (int64, error) {
	b, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrBalanceNotFound) {
			return 0, nil
		}
		return 0, domain.ErrInternal
	}

	return b.AmountPenny, nil
}
