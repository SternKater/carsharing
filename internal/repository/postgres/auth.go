package postgres

import (
	"context"
	"time"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(p *pgxpool.Pool) (*AuthRepository) {
	return &AuthRepository {
		pool: p,
	}
}

// TODO: check error whether user exists
func (a *AuthRepository) CreateUser(ctx context.Context, login string, hashedPwd string) (int64, error) {
	query := `
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`	
	tx := extractTx(ctx)
	var userId int64
	if tx != nil {
		if err := tx.QueryRow(ctx, query, login, hashedPwd).Scan(&userId); err != nil {
			return int64(domain.ErrNoUserID), err
		}

		return userId, nil
	}
	
	if err := a.pool.QueryRow(ctx, query, login, hashedPwd).Scan(&userId); err != nil {
		return int64(domain.ErrNoUserID), err
	}
	return userId, nil

}

func (a *AuthRepository) UpdateToken(ctx context.Context, userID int64, token string, updatedAt time.Time) error {
	query := `UPDATE refresh_tokens SET used_at = $1 WHERE user_id = $2 AND token_value = $3`
	tx := extractTx(ctx)
	
// just using pool
	if tx != nil {
		_, err := tx.Exec(ctx, query, updatedAt, userID, token)
		return err
	}
	
	_, err := a.pool.Exec(ctx, query, updatedAt, userID, token)
	return err
}

func (a *AuthRepository) RevokeTokens(ctx context.Context, userId int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id=$1`
	tx := extractTx(ctx)
	
	if tx != nil {
		_, err := tx.Exec(ctx, query, userId)
		return err
	}
	
	_, err := a.pool.Exec(ctx, query, userId)
	return err

}

func (a *AuthRepository) LoginUser(ctx context.Context, login string) (int64, string, error) {
	return int64(domain.ErrNoUserID), "", nil
}
	
func (a *AuthRepository) AddToken(ctx context.Context, token *domain.AuthRefreshToken) error {
	return nil
}

func (a *AuthRepository) RevokeToken(ctx context.Context, userId int64, userIdentify string) error {
	return nil
}

func (a *AuthRepository) GetToken(ctx context.Context, token string) (*domain.AuthRefreshToken, error) {
	return nil, nil
}


