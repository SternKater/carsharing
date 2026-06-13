package handler

import (
	"context"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/SternKater/carsharing/pkg/cars"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CarsServiceInterface interface {
	ListAvailableCars(ctx context.Context) ([]*domain.Car, error)
}

type CarsHandler struct {
	service CarsServiceInterface
	cars.UnimplementedCarsServiceServer
}

func NewCarsHandler(s CarsServiceInterface) *CarsHandler {
	return &CarsHandler{service: s}
}

func (h *CarsHandler) Cars(ctx context.Context, req *cars.CarsRequest) (*cars.CarsResponse, error) {
	domainCars, err := h.service.ListAvailableCars(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch cars catalog")
	}

	// carsItem2proto
	var pbCars []*cars.CarsItem
	for _, c := range domainCars {
		pbCars = append(pbCars, &cars.CarsItem{
			Id:     c.ID,
			Name:   c.CarName,
			Number: c.CarNumber,
			Ppm:    c.PricePerMinute,
			Status: c.Status,
		})
	}

	return &cars.CarsResponse{
		Success: true,
		Message: "Available cars fetched successfully(weeha!)",
		Cars:    pbCars,
	}, nil
}
