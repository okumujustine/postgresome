package collector

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/model"
)

type PostgresCollector struct {
	pool *pgxpool.Pool
}

func NewPostgresCollector(pool *pgxpool.Pool) *PostgresCollector {
	return &PostgresCollector{
		pool: pool,
	}
}

func (c *PostgresCollector) GetDatabaseInfo(ctx context.Context) (*model.DatabaseInfo, error) {
	var version string

	err := c.pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return nil, err
	}

	return &model.DatabaseInfo{
		Version: version,
	}, nil
}

func (c *PostgresCollector) GetDatabaseStats(ctx context.Context) (*model.DatabaseStats, error) {
	query := `
		SELECT
			datname,
			numbackends,
			xact_commit,
			xact_rollback,
			blks_read,
			blks_hit,
			tup_returned,
			tup_fetched,
			tup_inserted,
			tup_updated,
			tup_deleted
		FROM pg_stat_database
		WHERE datname = current_database();
	`

	var stats model.DatabaseStats

	err := c.pool.QueryRow(ctx, query).Scan(
		&stats.DatabaseName,
		&stats.NumBackends,
		&stats.XactCommit,
		&stats.XactRollback,
		&stats.BlksRead,
		&stats.BlksHit,
		&stats.TupReturned,
		&stats.TupFetched,
		&stats.TupInserted,
		&stats.TupUpdated,
		&stats.TupDeleted,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
