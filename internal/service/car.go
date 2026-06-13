package service

import (
	"context"

	"github.com/SternKater/carsharing/internal/domain"
)

type CarsRepositoryInterface interface {
	GetAvailableCars(ctx context.Context) ([]*domain.Car, error)
}

type CarsService struct {
	repo CarsRepositoryInterface
}

func NewCarsService(r CarsRepositoryInterface) *CarsService {
	return &CarsService{repo: r}
}

func (s *CarsService) ListAvailableCars(ctx context.Context) ([]*domain.Car, error) {
// just plain query(no filters)	
	return s.repo.GetAvailableCars(ctx)
}
