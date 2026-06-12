package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseInstance is a monitored Postgres database registered by an agent.
type DatabaseInstance struct {
	ID        string
	AgentID   string
	Name      string
	Host      string
	CreatedAt time.Time
}

const getDatabaseInstanceSQL = `
	SELECT id, agent_id, name, host, created_at
	FROM database_instances
	WHERE id = $1
`

// GetDatabaseInstance fetches a database instance by id, returning nil if no
// instance with that id is registered.
func GetDatabaseInstance(ctx context.Context, pool *pgxpool.Pool, id string) (*DatabaseInstance, error) {
	var instance DatabaseInstance

	err := pool.QueryRow(ctx, getDatabaseInstanceSQL, id).Scan(
		&instance.ID, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance %q: %w", id, err)
	}

	return &instance, nil
}

const getMostRecentDatabaseInstanceSQL = `
	SELECT id, agent_id, name, host, created_at
	FROM database_instances
	ORDER BY created_at DESC
	LIMIT 1
`

// GetMostRecentDatabaseInstance returns the most recently registered database
// instance, or nil if none are registered yet.
func GetMostRecentDatabaseInstance(ctx context.Context, pool *pgxpool.Pool) (*DatabaseInstance, error) {
	var instance DatabaseInstance

	err := pool.QueryRow(ctx, getMostRecentDatabaseInstanceSQL).Scan(
		&instance.ID, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get most recent database instance: %w", err)
	}

	return &instance, nil
}
