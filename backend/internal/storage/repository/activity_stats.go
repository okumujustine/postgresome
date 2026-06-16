package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

type ActivityStatRow struct {
	DatabaseName string

	ProcessID int

	UserName        string
	ApplicationName string
	ClientAddress   string

	State string
	Query string

	WaitEventType string
	WaitEvent     string

	BackendStartedAt time.Time
	QueryStartedAt   *time.Time
	StateChangedAt   *time.Time
}

type ActivityStatsSnapshot struct {
	CollectedAt time.Time
	Activities  []ActivityStatRow
}

const deleteActivityStatsSQL = `
	DELETE FROM activity_stats WHERE database_instance_id = $1
`

const insertActivityStatSQL = `
	INSERT INTO activity_stats (
		database_instance_id, agent_id, collected_at, database_name, process_id,
		user_name, application_name, client_address, state, query,
		wait_event_type, wait_event, backend_started_at, query_started_at, state_changed_at
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15
	)
`

func ReplaceActivityStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, activities []ActivityStatRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteActivityStatsSQL, databaseInstanceID); err != nil {
		return fmt.Errorf("failed to delete existing activity stats: %w", err)
	}

	for _, activity := range activities {
		_, err := tx.Exec(ctx, insertActivityStatSQL,
			databaseInstanceID, agentID, collectedAt, activity.DatabaseName, activity.ProcessID,
			activity.UserName, activity.ApplicationName, activity.ClientAddress, activity.State, activity.Query,
			activity.WaitEventType, activity.WaitEvent, activity.BackendStartedAt, activity.QueryStartedAt, activity.StateChangedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert activity stat for pid %d: %w", activity.ProcessID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit activity stats transaction: %w", err)
	}

	return nil
}

const listActivityStatsSQL = `
	SELECT collected_at, database_name, process_id, user_name, application_name, client_address,
	       state, query, wait_event_type, wait_event, backend_started_at, query_started_at, state_changed_at
	FROM activity_stats
	WHERE database_instance_id = $1
	ORDER BY process_id ASC
`

const listActivityStatsAtSQL = `
	SELECT collected_at, database_name, process_id, user_name, application_name, client_address,
	       state, query, wait_event_type, wait_event, backend_started_at, query_started_at, state_changed_at
	FROM activity_stats
	WHERE database_instance_id = $1
	  AND collected_at = $2
	ORDER BY process_id ASC
`

func ListActivityStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (ActivityStatsSnapshot, error) {
	return listActivityStatsSnapshot(ctx, pool, listActivityStatsSQL, databaseInstanceID)
}

func ListActivityStatsAt(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string, collectedAt time.Time) (ActivityStatsSnapshot, error) {
	return listActivityStatsSnapshot(ctx, pool, listActivityStatsAtSQL, databaseInstanceID, collectedAt)
}

func listActivityStatsSnapshot(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (ActivityStatsSnapshot, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return ActivityStatsSnapshot{}, fmt.Errorf("failed to query activity stats: %w", err)
	}
	defer rows.Close()

	snapshot := ActivityStatsSnapshot{Activities: make([]ActivityStatRow, 0)}

	for rows.Next() {
		var (
			activity    ActivityStatRow
			collectedAt time.Time
		)

		if err := rows.Scan(
			&collectedAt, &activity.DatabaseName, &activity.ProcessID, &activity.UserName, &activity.ApplicationName, &activity.ClientAddress,
			&activity.State, &activity.Query, &activity.WaitEventType, &activity.WaitEvent, &activity.BackendStartedAt, &activity.QueryStartedAt, &activity.StateChangedAt,
		); err != nil {
			return ActivityStatsSnapshot{}, fmt.Errorf("failed to scan activity stat: %w", err)
		}

		snapshot.CollectedAt = collectedAt
		snapshot.Activities = append(snapshot.Activities, activity)
	}

	if err := rows.Err(); err != nil {
		return ActivityStatsSnapshot{}, fmt.Errorf("failed to read activity stats: %w", err)
	}

	return snapshot, nil
}

func (s ActivityStatsSnapshot) ToMetrics(databaseInstanceID string) *metrics.DatabaseActivitySnapshot {
	activities := make([]metrics.DatabaseActivity, len(s.Activities))
	for i, activity := range s.Activities {
		activities[i] = metrics.DatabaseActivity{
			DatabaseName:     activity.DatabaseName,
			ProcessID:        activity.ProcessID,
			UserName:         activity.UserName,
			ApplicationName:  activity.ApplicationName,
			ClientAddress:    activity.ClientAddress,
			State:            activity.State,
			Query:            activity.Query,
			WaitEventType:    activity.WaitEventType,
			WaitEvent:        activity.WaitEvent,
			BackendStartedAt: activity.BackendStartedAt,
			QueryStartedAt:   activity.QueryStartedAt,
			StateChangedAt:   activity.StateChangedAt,
		}
	}

	return &metrics.DatabaseActivitySnapshot{
		CollectedAt:        s.CollectedAt,
		DatabaseInstanceID: databaseInstanceID,
		Activities:         activities,
	}
}
