package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UpsertAgent inserts a new agent or updates an existing one with the same id.
func UpsertAgent(ctx context.Context, pool *pgxpool.Pool, id, name, environment string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO agents (id, name, environment)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			environment = EXCLUDED.environment
	`, id, name, environment)

	return err
}

// UpsertDatabaseInstance inserts a new database instance or updates an
// existing one with the same id.
func UpsertDatabaseInstance(ctx context.Context, pool *pgxpool.Pool, id, agentID, name, host string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO database_instances (id, agent_id, name, host)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			agent_id = EXCLUDED.agent_id,
			name = EXCLUDED.name,
			host = EXCLUDED.host
	`, id, agentID, name, host)

	return err
}
