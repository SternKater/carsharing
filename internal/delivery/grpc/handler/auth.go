package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/SternKater/carsharing/pkg/auth"
)

type AuthServiceInterface interface {
	CreateUser(ctx context.Context, login string, password string, userIdentify string) (*domain.Auth, error)
	LoginUser(ctx context.Context, login string, password string, userIdentify string) (*domain.Auth, error)
	LogoutUser(ctx context.Context, userId int64, userIdentify string) (error)
	NewTokenUser(ctx context.Context, refreshToken string, userIdentify string) (*domain.Auth, error)
}

type AuthHandler struct {
	service AuthServiceInterface
	auth.UnimplementedAuthServiceServer
}

func NewAuthHandler(s AuthServiceInterface) (*AuthHandler) {
	return &AuthHandler {
		service: s,
	}
}

func (a *AuthHandler) SignUp(ctx context.Context, req *auth.SignUpRequest) (*auth.SignUpResponse, error) {
	
	user, err := a.service.CreateUser(ctx, req.UserName, req.UserPwd, req.UserIdentifyString)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "Wrong login or password")
		case errors.Is(err, domain.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "Already exists")
		default:
			return nil, status.Error(codes.Internal, "Something went wrong")
		}
	}

	return &auth.SignUpResponse{
		Success: true,
		RefreshToken: user.RefreshToken,
		AccessToken: user.AccessToken,
		Message: "User has been created",
	}, nil
}

func (a *AuthHandler) SignIn(ctx context.Context, req *auth.SignInRequest) (*auth.SignInResponse, error) {

	user, err := a.service.LoginUser(ctx, req.UserName, req.UserPassword, req.UserIdentifyString)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "Wrong login or password")
		default:
			return nil, status.Error(codes.Internal, "Something went wrong")
		}
	}

	return &auth.SignInResponse{
		Success:      true,
		Message:      "Welcome aboard, sir!",
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}, nil
}

func (a *AuthHandler) SignOut(ctx context.Context, req *auth.SignOutRequest) (*auth.SignOutResponse, error) {
	userID, _ := domain.GetUserID(ctx)
	err := a.service.LogoutUser(ctx, userID, req.UserIdentify)
	if err != nil {
		return nil, status.Error(codes.Internal, "Cannot log out")
	}
	return &auth.SignOutResponse{
		Success: true,
		Message: "Logged out",
	}, nil
}

func (a *AuthHandler) AuthRefresh(ctx context.Context, req *auth.AuthRefreshRequest) (*auth.AuthRefreshResponse, error) {
	user, err := a.service.NewTokenUser(ctx, req.RefreshToken, req.UserIdentifyString)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRefreshToken):
			return nil, status.Error(codes.Unauthenticated, "Refresh token is expired or invalid")
		default:
			return nil, status.Error(codes.Internal, "Something went wrong")
		}
	}
	return &auth.AuthRefreshResponse {
		Success: true,
		Message: "New refresh token has been generated",
		RefreshToken: user.RefreshToken,
		AccessToken: user.AccessToken,
	}, nil
}
