package extensions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExtensionCollector struct {
	pool *pgxpool.Pool
}

func NewExtensionCollector(pool *pgxpool.Pool) *ExtensionCollector {
	return &ExtensionCollector{
		pool: pool,
	}
}

// IsPgStatStatementsEnabled checks whether the pg_stat_statements extension
// is installed in the current database.
func (c *ExtensionCollector) IsPgStatStatementsEnabled(ctx context.Context) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements');`

	var enabled bool

	if err := c.pool.QueryRow(ctx, query).Scan(&enabled); err != nil {
		return false, err
	}

	return enabled, nil
}
