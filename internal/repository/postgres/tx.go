package postgres

import (
	"context"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey	string

const TxManagerKey  contextKey = "tx_manager"

type TxQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxManagerKey, tx)
}

func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(TxManagerKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func getExecutor(ctx context.Context, defaultPool *pgxpool.Pool) TxQuerier {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return defaultPool
}

func (tm *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return domain.ErrInternal
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			return
		}
	}()

	txCtx := injectTx(ctx, tx)
	err = fn(txCtx)

	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ErrInternal
	}

	return nil
}
