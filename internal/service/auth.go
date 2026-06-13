package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/SternKater/carsharing/internal/domain"
)

type AuthRepositoryInterface interface {
	CreateUser(ctx context.Context, login string, hashedPwd string) (int64, error)
	// id, hashed_pwd
	LoginUser(ctx context.Context, login string) (int64, string, error)

	AddToken(ctx context.Context, token *domain.AuthRefreshToken) error
	UpdateToken(ctx context.Context, userId int64, refreshToken string, updateAt time.Time) error
	RevokeTokens(ctx context.Context, userId int64) error
	RevokeToken(ctx context.Context, userId int64, userIdentify string) error
	GetLastActiveToken(ctx context.Context, userId int64, userIdentify string) (string, error)
	GetToken(ctx context.Context, token string) (*domain.AuthRefreshToken, error)
}

type AuthTokenInterface interface {
	GenerateRefreshToken() string
	JWTByUserID(id int64) (string, error)
}

type AuthService struct {
	repo  AuthRepositoryInterface
	tkMgr AuthTokenInterface
	txMgr domain.TransactionManager
}

func NewAuthService(r AuthRepositoryInterface, t AuthTokenInterface, tx domain.TransactionManager) *AuthService {
	return &AuthService{
		repo:  r,
		tkMgr: t,
		txMgr: tx,
	}
}

func (a *AuthService) CreateUser(ctx context.Context, login string, password string, userIdentify string) (*domain.Auth, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	var userID int64
	var refresh string

	err = a.txMgr.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error

		userID, err = a.repo.CreateUser(txCtx, login, string(hashed))
		if err != nil {
			return err
		}

		refresh = a.tkMgr.GenerateRefreshToken()
		refreshToken := &domain.AuthRefreshToken{
			UserID:             userID,
			TokenValue:         refresh,
			UserIdentifyString: userIdentify,
			ExpiresAt:          time.Now().AddDate(1, 0, 0),
			UsedAt:             nil,
		}

		if err := a.repo.AddToken(txCtx, refreshToken); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	jwtToken, err := a.tkMgr.JWTByUserID(userID)
	if err != nil {
		return nil, domain.ErrInternal
	}
	return &domain.Auth{
		RefreshToken: refresh,
		AccessToken:  jwtToken,
	}, nil
}

func (a *AuthService) LoginUser(ctx context.Context, login string, password string, userIdentify string) (*domain.Auth, error) {

	id, hashedPwd, err := a.repo.LoginUser(ctx, login)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) ||
			errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// compare with incoming password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	refresh := a.tkMgr.GenerateRefreshToken()
	refreshToken := &domain.AuthRefreshToken{
		UserID:             id,
		TokenValue:         refresh,
		UserIdentifyString: userIdentify,
		ExpiresAt:          time.Now().AddDate(1, 0, 0),
		UsedAt:             nil,
	}
	if err := a.repo.AddToken(ctx, refreshToken); err != nil {
		return nil, domain.ErrInternal
	}
	jwtToken, err := a.tkMgr.JWTByUserID(id)
	if err != nil {
		return nil, domain.ErrInternal
	}
	return &domain.Auth{
		AccessToken:  jwtToken,
		RefreshToken: refresh,
	}, nil
}

func (a *AuthService) LogoutUser(ctx context.Context, userId int64, userIdentify string) error {
	err := a.repo.RevokeToken(ctx, userId, userIdentify)
	return err
}

func (a *AuthService) NewTokenUser(ctx context.Context, refreshToken string, userIdentify string) (*domain.Auth, error) {

	var authEntity *domain.Auth
	err := a.txMgr.WithinTransaction(ctx, func(txCtx context.Context) error {
		refreshEntity, err := a.repo.GetToken(txCtx, refreshToken)
		// not found
		if errors.Is(err, domain.ErrInvalidCredentials) || 
		errors.Is(err, domain.ErrUserNotFound) {
			return err
		}
		// panic! refresh token has stolen!
		if refreshEntity.UserIdentifyString != userIdentify {
			a.repo.RevokeTokens(txCtx, refreshEntity.UserID)
			return domain.ErrStolenRefreshToken
		}
		// refresh token has expired
		if refreshEntity.ExpiresAt.Before(time.Now()) {
			return domain.ErrInvalidCredentials
		}
		usedAt := refreshEntity.UsedAt
		// didnt use before
		if usedAt == nil {
			if err := a.repo.UpdateToken(txCtx, refreshEntity.UserID, refreshToken, time.Now()); err != nil {
				return domain.ErrInternal
			}
			newRefresh := a.tkMgr.GenerateRefreshToken()
			newEntity := &domain.AuthRefreshToken{
				UserID:             refreshEntity.UserID,
				TokenValue:         newRefresh,
				UserIdentifyString: userIdentify,
				UsedAt:             nil,
				ExpiresAt:          time.Now().AddDate(1, 0, 0),
			}
			if err := a.repo.AddToken(txCtx, newEntity); err != nil {
				return domain.ErrInternal
			}
			jwtToken, err := a.tkMgr.JWTByUserID(refreshEntity.UserID)
			if err != nil {
				return domain.ErrInternal
			}
			authEntity = &domain.Auth{
				ID:           refreshEntity.UserID,
				Name:         "",
				RefreshToken: newRefresh,
				AccessToken:  jwtToken,
			}
			return nil
		}
		// token is used but it's still grace period
		gracePeriod := usedAt.Add(5 * time.Minute)
		if gracePeriod.After(time.Now()) {
			activeRefresh, err := a.repo.GetLastActiveToken(ctx, refreshEntity.UserID, refreshEntity.UserIdentifyString)
			if err != nil {
				return domain.ErrInternal
			}
			jwtToken, err := a.tkMgr.JWTByUserID(refreshEntity.UserID)
			if err != nil {
				return domain.ErrInternal
			}

			authEntity = &domain.Auth{
				ID:           refreshEntity.UserID,
				Name:         "",
				RefreshToken: activeRefresh,
				AccessToken:  jwtToken,
			}
			return nil
		}
		return domain.ErrInvalidRefreshToken
	})

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) ||
			errors.Is(err, domain.ErrStolenRefreshToken) ||
			errors.Is(err, domain.ErrTokenExpired) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}
	return authEntity, nil
}
