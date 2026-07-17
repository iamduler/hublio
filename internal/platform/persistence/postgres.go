package persistence

import (
	"context"
	"fmt"
	"time"

	"hublio/internal/platform/config"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type Database struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := cfg.DNS()

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing DB config: %w", err)
	}

	sqlLogger := logging.NewLoggerWithPath("sql.log", "info")
	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: &PgxZerologTracer{
			Logger:         *sqlLogger,
			SlowQueryLimit: 500 * time.Millisecond,
		},
		LogLevel: tracelog.LogLevelDebug,
	}

	poolConfig.MaxConns = 50
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating DB pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping error: %w", err)
	}

	if logging.Log != nil {
		logging.Log.Info().Msg("PostgreSQL connected")
	}

	return &Database{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

func (d *Database) Close() {
	if d != nil && d.Pool != nil {
		d.Pool.Close()
	}
}

// Querier returns sqlc queries bound to the active transaction when present.
func (d *Database) Querier(ctx context.Context) *sqlc.Queries {
	if tx, ok := TxFromContext(ctx); ok {
		return d.Queries.WithTx(tx)
	}
	return d.Queries
}

// Ping verifies the database connection.
func (d *Database) Ping(ctx context.Context) error {
	if d == nil || d.Pool == nil {
		return fmt.Errorf("persistence: database not initialized")
	}
	return d.Pool.Ping(ctx)
}
