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
	ID                   string
	Kind                 string
	Provider             string
	Name                 string
	AgentID              string
	Configured           bool
	LastCheckStartedAt   *time.Time
	LastCheckCompletedAt *time.Time
	LastCheckStatus      string
	LastCheckError       string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type SourceConnectionProfile struct {
	SourceID               string
	ConnectionURI          string
	ConnectionURIEncrypted string
	CreatedAt              time.Time
	UpdatedAt              time.Time
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
	SELECT
		s.id,
		s.kind,
		s.provider,
		s.name,
		COALESCE(s.agent_id, ''),
		EXISTS (
			SELECT 1
			FROM source_connection_profiles scp
			WHERE scp.source_id = s.id
			  AND (scp.connection_uri <> '' OR scp.connection_uri_encrypted <> '')
		),
		s.last_check_started_at,
		s.last_check_completed_at,
		s.last_check_status,
		s.last_check_error,
		s.created_at,
		s.updated_at
	FROM sources
	WHERE s.id = $1
`

func GetSource(ctx context.Context, pool *pgxpool.Pool, id string) (*Source, error) {
	var source Source

	err := pool.QueryRow(ctx, getSourceSQL, id).Scan(
		&source.ID,
		&source.Kind,
		&source.Provider,
		&source.Name,
		&source.AgentID,
		&source.Configured,
		&source.LastCheckStartedAt,
		&source.LastCheckCompletedAt,
		&source.LastCheckStatus,
		&source.LastCheckError,
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
	SELECT
		s.id,
		s.kind,
		s.provider,
		s.name,
		COALESCE(s.agent_id, ''),
		EXISTS (
			SELECT 1
			FROM source_connection_profiles scp
			WHERE scp.source_id = s.id
			  AND (scp.connection_uri <> '' OR scp.connection_uri_encrypted <> '')
		),
		s.last_check_started_at,
		s.last_check_completed_at,
		s.last_check_status,
		s.last_check_error,
		s.created_at,
		s.updated_at
	FROM sources s
	ORDER BY s.created_at ASC, s.name ASC
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
			&source.Configured,
			&source.LastCheckStartedAt,
			&source.LastCheckCompletedAt,
			&source.LastCheckStatus,
			&source.LastCheckError,
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

const upsertSourceConnectionProfileSQL = `
	INSERT INTO source_connection_profiles (source_id, connection_uri, connection_uri_encrypted)
	VALUES ($1, '', $2)
	ON CONFLICT (source_id) DO UPDATE SET
		connection_uri = EXCLUDED.connection_uri,
		connection_uri_encrypted = EXCLUDED.connection_uri_encrypted,
		updated_at = NOW()
`

func UpsertSourceConnectionProfile(ctx context.Context, pool *pgxpool.Pool, sourceID, connectionURIEncrypted string) error {
	_, err := pool.Exec(ctx, upsertSourceConnectionProfileSQL, sourceID, connectionURIEncrypted)
	if err != nil {
		return fmt.Errorf("failed to upsert source connection profile for %q: %w", sourceID, err)
	}
	return nil
}

const getSourceConnectionProfileSQL = `
	SELECT source_id, connection_uri, connection_uri_encrypted, created_at, updated_at
	FROM source_connection_profiles
	WHERE source_id = $1
`

func GetSourceConnectionProfile(ctx context.Context, pool *pgxpool.Pool, sourceID string) (*SourceConnectionProfile, error) {
	var profile SourceConnectionProfile

	err := pool.QueryRow(ctx, getSourceConnectionProfileSQL, sourceID).Scan(
		&profile.SourceID,
		&profile.ConnectionURI,
		&profile.ConnectionURIEncrypted,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get source connection profile for %q: %w", sourceID, err)
	}

	return &profile, nil
}

const markSourceCheckRunningSQL = `
	UPDATE sources
	SET
		last_check_started_at = $2,
		last_check_status = 'running',
		last_check_error = '',
		updated_at = NOW()
	WHERE id = $1
`

func MarkSourceCheckRunning(ctx context.Context, pool *pgxpool.Pool, sourceID string, startedAt time.Time) error {
	_, err := pool.Exec(ctx, markSourceCheckRunningSQL, sourceID, startedAt)
	if err != nil {
		return fmt.Errorf("failed to mark source %q check as running: %w", sourceID, err)
	}
	return nil
}

const markSourceCheckFinishedSQL = `
	UPDATE sources
	SET
		last_check_completed_at = $2,
		last_check_status = $3,
		last_check_error = $4,
		updated_at = NOW()
	WHERE id = $1
`

func MarkSourceCheckFinished(ctx context.Context, pool *pgxpool.Pool, sourceID string, completedAt time.Time, status, lastCheckError string) error {
	_, err := pool.Exec(ctx, markSourceCheckFinishedSQL, sourceID, completedAt, status, lastCheckError)
	if err != nil {
		return fmt.Errorf("failed to mark source %q check as %s: %w", sourceID, status, err)
	}
	return nil
}
