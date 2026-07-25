package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is satisfied by *pgxpool.Pool and pgx.Tx so repositories stay transaction-aware.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type txKey struct{}

// WithinTransaction runs fn inside a single database transaction.
// Repositories must not begin or commit transactions themselves.
func WithinTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if pool == nil {
		return fmt.Errorf("persistence: nil pool")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistence: begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistence: commit tx: %w", err)
	}

	return nil
}

// TxFromContext returns the active transaction if present.
func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// Conn returns the transaction from context when available, otherwise the pool.
func Conn(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return pool
}
