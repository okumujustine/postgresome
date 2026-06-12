package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a connection pool to the Postgresome application database
// and verifies connectivity with a ping.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
