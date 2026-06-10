package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type contextKey	string

const TxManagerKey  contextKey = "tx_manager"

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxManagerKey, tx)
}

func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(TxManagerKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (tm *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		panic(err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) 
		}
	}()

	txCtx := injectTx(ctx, tx)
	err = fn(txCtx)

	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	return nil
}
