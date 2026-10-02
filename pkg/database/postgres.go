package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool initializes a connection pool to PostgreSQL using pgx
func NewPostgresPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	var pingErr error
	for i := 0; i < 15; i++ {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			break
		}
		slog.Warn("PostgreSQL ping failed, retrying in 2 seconds...", "attempt", i+1, "error", pingErr)
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping postgres after 15 attempts: %w", pingErr)
	}

	return pool, nil
}
