package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseInstance is a monitored Postgres database linked to a source.
type DatabaseInstance struct {
	ID         string
	SourceID   string
	SourceKind string
	Provider   string
	AgentID    string
	Name       string
	Host       string
	CreatedAt  time.Time
}

const getDatabaseInstanceSQL = `
	SELECT di.id, di.source_id, s.kind, s.provider, COALESCE(di.agent_id, ''), di.name, di.host, di.created_at
	FROM database_instances di
	JOIN sources s ON s.id = di.source_id
	WHERE di.id = $1
`

// GetDatabaseInstance fetches a database instance by id, returning nil if no
// instance with that id is registered.
func GetDatabaseInstance(ctx context.Context, pool *pgxpool.Pool, id string) (*DatabaseInstance, error) {
	var instance DatabaseInstance

	err := pool.QueryRow(ctx, getDatabaseInstanceSQL, id).Scan(
		&instance.ID, &instance.SourceID, &instance.SourceKind, &instance.Provider, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance %q: %w", id, err)
	}

	return &instance, nil
}

const getDatabaseInstanceBySourceIDSQL = `
	SELECT di.id, di.source_id, s.kind, s.provider, COALESCE(di.agent_id, ''), di.name, di.host, di.created_at
	FROM database_instances di
	JOIN sources s ON s.id = di.source_id
	WHERE di.source_id = $1
	ORDER BY di.created_at ASC
	LIMIT 1
`

// GetDatabaseInstanceBySourceID returns the first database instance registered
// for a source. The current product model uses one primary database instance
// per source.
func GetDatabaseInstanceBySourceID(ctx context.Context, pool *pgxpool.Pool, sourceID string) (*DatabaseInstance, error) {
	var instance DatabaseInstance

	err := pool.QueryRow(ctx, getDatabaseInstanceBySourceIDSQL, sourceID).Scan(
		&instance.ID, &instance.SourceID, &instance.SourceKind, &instance.Provider, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance for source %q: %w", sourceID, err)
	}

	return &instance, nil
}

const listDatabaseInstancesSQL = `
	SELECT di.id, di.source_id, s.kind, s.provider, COALESCE(di.agent_id, ''), di.name, di.host, di.created_at
	FROM database_instances di
	JOIN sources s ON s.id = di.source_id
	ORDER BY di.name ASC
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
		if err := rows.Scan(&instance.ID, &instance.SourceID, &instance.SourceKind, &instance.Provider, &instance.AgentID, &instance.Name, &instance.Host, &instance.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan database instance: %w", err)
		}
		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read database instances: %w", err)
	}

	return instances, nil
}
