package postgres

import (
	"context"
	"errors"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceRepository(p *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: p}
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	query := `SELECT user_id, amount_penny FROM user_balances WHERE user_id = $1`

	var b domain.Balance
	err := getExecutor(ctx, r.pool).QueryRow(ctx, query, userID).Scan(&b.UserID, &b.AmountPenny)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBalanceNotFound
		}
		return nil, domain.ErrInternal
	}

	return &b, nil
}
