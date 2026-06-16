package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UpsertDatabaseInstance inserts a new database instance or updates an
// existing one with the same id.
func UpsertDatabaseInstance(ctx context.Context, pool *pgxpool.Pool, id, sourceID, agentID, name, host string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO database_instances (id, source_id, agent_id, name, host)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			agent_id = EXCLUDED.agent_id,
			name = EXCLUDED.name,
			host = EXCLUDED.host
	`, id, sourceID, nullIfEmpty(agentID), name, host)

	return err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
