package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Source struct {
	ID        string
	Kind      string
	Provider  string
	Name      string
	AgentID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const upsertSourceSQL = `
	INSERT INTO sources (id, kind, provider, name, agent_id)
	VALUES ($1, $2, $3, $4, NULLIF($5, ''))
	ON CONFLICT (id) DO UPDATE SET
		kind = EXCLUDED.kind,
		provider = EXCLUDED.provider,
		name = EXCLUDED.name,
		agent_id = EXCLUDED.agent_id,
		updated_at = NOW()
`

func UpsertSource(ctx context.Context, pool *pgxpool.Pool, id, kind, provider, name, agentID string) error {
	_, err := pool.Exec(ctx, upsertSourceSQL, id, kind, provider, name, agentID)
	if err != nil {
		return fmt.Errorf("failed to upsert source %q: %w", id, err)
	}
	return nil
}

const getSourceSQL = `
	SELECT id, kind, provider, name, COALESCE(agent_id, ''), created_at, updated_at
	FROM sources
	WHERE id = $1
`

func GetSource(ctx context.Context, pool *pgxpool.Pool, id string) (*Source, error) {
	var source Source

	err := pool.QueryRow(ctx, getSourceSQL, id).Scan(
		&source.ID,
		&source.Kind,
		&source.Provider,
		&source.Name,
		&source.AgentID,
		&source.CreatedAt,
		&source.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get source %q: %w", id, err)
	}

	return &source, nil
}

const listSourcesSQL = `
	SELECT id, kind, provider, name, COALESCE(agent_id, ''), created_at, updated_at
	FROM sources
	ORDER BY created_at ASC, name ASC
`

func ListSources(ctx context.Context, pool *pgxpool.Pool) ([]Source, error) {
	rows, err := pool.Query(ctx, listSourcesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}
	defer rows.Close()

	sources := make([]Source, 0)
	for rows.Next() {
		var source Source
		if err := rows.Scan(
			&source.ID,
			&source.Kind,
			&source.Provider,
			&source.Name,
			&source.AgentID,
			&source.CreatedAt,
			&source.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read sources: %w", err)
	}

	return sources, nil
}
