package handler

import (
	"context"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/SternKater/carsharing/pkg/billing/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BalanceServiceInterface interface {
	GetUserBalance(ctx context.Context, userID int64) (int64, error)
}

type BalanceHandler struct {
	service BalanceServiceInterface
	wallet.UnimplementedBalanceServiceServer
}

func NewBalanceHandler(s BalanceServiceInterface) *BalanceHandler {
	return &BalanceHandler{service: s}
}

func (h *BalanceHandler) Balance(ctx context.Context, req *wallet.BalanceRequest) (*wallet.BalanceResponse, error) {
	userID, ok := domain.GetUserID(ctx)
	if !ok || userID == domain.ErrNoUserID {
		return nil, status.Error(codes.Unauthenticated, "Who are you? No user_id detected!")
	}

	amount, err := h.service.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch balance")
	}

	return &wallet.BalanceResponse{
		Success: true,
		Amount:  amount,
		Message: "Balance fetched successfully",
	}, nil
}
