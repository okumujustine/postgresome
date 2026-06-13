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

const listDatabaseInstancesSQL = `
	SELECT id, agent_id, name, host, created_at
	FROM database_instances
	ORDER BY name ASC
`

// ListDatabaseInstances returns every registered database instance, ordered
// by name, for use in a database selector UI.
func ListDatabaseInstances(ctx context.Context, pool *pgxpool.Pool) ([]DatabaseInstance, error) {
	rows, err := pool.Query(ctx, listDatabaseInstancesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to list database instances: %w", err)
	}
	defer rows.Close()

	instances := make([]DatabaseInstance, 0)

	for rows.Next() {
		var instance DatabaseInstance
		if err := rows.Scan(&instance.ID, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan database instance: %w", err)
		}
		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read database instances: %w", err)
	}

	return instances, nil
}
