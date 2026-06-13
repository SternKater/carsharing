package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(p *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: p,
	}
}

func (a *AuthRepository) CreateUser(ctx context.Context, login string, hashedPwd string) (int64, error) {
	query := `
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`
	var userId int64

	if err := getExecutor(ctx, a.pool).QueryRow(ctx, query, login, hashedPwd).Scan(&userId); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrNoUserID, domain.ErrUserAlreadyExists
		}
		return domain.ErrNoUserID, domain.ErrInternal
	}

	return userId, nil
}

func (a *AuthRepository) UpdateToken(ctx context.Context, userID int64, token string, updatedAt time.Time) error {
	query := `UPDATE refresh_tokens SET used_at = $1 WHERE user_id = $2 AND token_value = $3`
	if _, err := getExecutor(ctx, a.pool).Exec(ctx, query, updatedAt, userID, token); err != nil {
		return domain.ErrInternal
	}

	return nil
}

func (a *AuthRepository) RevokeTokens(ctx context.Context, userId int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id=$1`
	if _, err := getExecutor(ctx, a.pool).Exec(ctx, query, userId); err != nil {
		return domain.ErrInternal
	}
	return nil
}

func (a *AuthRepository) LoginUser(ctx context.Context, login string) (int64, string, error) {
	query := `SELECT id, password_hash FROM users WHERE login=$1`

	var hashedPwd string
	var userId int64
	if err := getExecutor(ctx, a.pool).QueryRow(ctx, query, login).Scan(&userId, &hashedPwd); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNoUserID, "", domain.ErrUserNotFound
		}
		return domain.ErrNoUserID, "", domain.ErrInternal
	}

	return userId, hashedPwd, nil
}

func (a *AuthRepository) AddToken(ctx context.Context, token *domain.AuthRefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_value, expires_at, user_identify_string) 
		VALUES ($1, $2, $3, $4) 
	`
	if _, err := getExecutor(ctx, a.pool).Exec(ctx, query, token.UserID, token.TokenValue, token.ExpiresAt, token.UserIdentifyString); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrInvalidRefreshToken
		}
		return domain.ErrInternal
	}

	return nil
}

func (a *AuthRepository) RevokeToken(ctx context.Context, userId int64, tokenValue string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1 AND user_identify_string = $2`
	if _, err := getExecutor(ctx, a.pool).Exec(ctx, query, userId, tokenValue); err != nil {
		return domain.ErrInternal
	}

	return nil
}

func (a *AuthRepository) GetLastActiveToken(ctx context.Context, userId int64, userIdentify string) (string, error) {
	query := `SELECT token_value FROM refresh_tokens WHERE used_at IS NULL AND user_id=$1 AND user_identify_string=$2 FOR UPDATE`
	var activeToken = ""
	if err := getExecutor(ctx, a.pool).QueryRow(ctx, query, userId, userIdentify).Scan(&activeToken); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return activeToken, domain.ErrUserNotFound
		}
		return activeToken, domain.ErrInternal
	}

	return activeToken, nil
}

func (a *AuthRepository) GetToken(ctx context.Context, token string) (*domain.AuthRefreshToken, error) {
	var refreshToken = &domain.AuthRefreshToken{}
	query := `SELECT user_id, token_value, expires_at, used_at, user_identify_string FROM refresh_tokens WHERE token_value=$1 FOR UPDATE`

	if err := getExecutor(ctx, a.pool).QueryRow(ctx, query, token).Scan(
		&refreshToken.UserID,
		&refreshToken.TokenValue,
		&refreshToken.ExpiresAt,
		&refreshToken.UsedAt,
		&refreshToken.UserIdentifyString,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.ErrInternal
	}

	return refreshToken, nil
}
