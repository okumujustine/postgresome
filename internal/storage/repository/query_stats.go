package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryStatRow is one query's statistics within a QueryStatsSnapshot.
type QueryStatRow struct {
	QueryID      string
	DatabaseName string
	UserName     string
	Query        string

	Calls int64

	TotalExecTimeMs float64
	MeanExecTimeMs  float64
	MinExecTimeMs   float64
	MaxExecTimeMs   float64

	RowsReturned int64

	SharedBlocksHit     int64
	SharedBlocksRead    int64
	SharedBlocksDirtied int64
	SharedBlocksWritten int64

	TempBlocksRead    int64
	TempBlocksWritten int64
}

// QueryStatsSnapshot is the latest set of query statistics for a database
// instance.
type QueryStatsSnapshot struct {
	CollectedAt time.Time
	Queries     []QueryStatRow
}

const deleteQueryStatsSQL = `
	DELETE FROM query_stats WHERE database_instance_id = $1
`

const insertQueryStatSQL = `
	INSERT INTO query_stats (
		database_instance_id, agent_id, collected_at, query_id, database_name, user_name, query,
		calls, total_exec_time_ms, mean_exec_time_ms, min_exec_time_ms, max_exec_time_ms, rows_returned,
		shared_blocks_hit, shared_blocks_read, shared_blocks_dirtied, shared_blocks_written,
		temp_blocks_read, temp_blocks_written
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13,
		$14, $15, $16, $17,
		$18, $19
	)
`

// ReplaceQueryStats replaces the stored query statistics snapshot for a
// database instance with the given rows, in a single transaction.
func ReplaceQueryStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, queries []QueryStatRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteQueryStatsSQL, databaseInstanceID); err != nil {
		return fmt.Errorf("failed to delete existing query stats: %w", err)
	}

	for _, query := range queries {
		_, err := tx.Exec(ctx, insertQueryStatSQL,
			databaseInstanceID, agentID, collectedAt, query.QueryID, query.DatabaseName, query.UserName, query.Query,
			query.Calls, query.TotalExecTimeMs, query.MeanExecTimeMs, query.MinExecTimeMs, query.MaxExecTimeMs, query.RowsReturned,
			query.SharedBlocksHit, query.SharedBlocksRead, query.SharedBlocksDirtied, query.SharedBlocksWritten,
			query.TempBlocksRead, query.TempBlocksWritten,
		)
		if err != nil {
			return fmt.Errorf("failed to insert query stat for %q: %w", query.QueryID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit query stats transaction: %w", err)
	}

	return nil
}

const listQueryStatsSQL = `
	SELECT collected_at, query_id, database_name, user_name, query,
	       calls, total_exec_time_ms, mean_exec_time_ms, min_exec_time_ms, max_exec_time_ms, rows_returned,
	       shared_blocks_hit, shared_blocks_read, shared_blocks_dirtied, shared_blocks_written,
	       temp_blocks_read, temp_blocks_written
	FROM query_stats
	WHERE database_instance_id = $1
	ORDER BY total_exec_time_ms DESC
`

// ListQueryStats returns the latest query statistics snapshot for a database
// instance. If no snapshot has been stored yet, it returns a zero-value
// snapshot (zero CollectedAt, empty Queries).
func ListQueryStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (QueryStatsSnapshot, error) {
	rows, err := pool.Query(ctx, listQueryStatsSQL, databaseInstanceID)
	if err != nil {
		return QueryStatsSnapshot{}, fmt.Errorf("failed to query query stats: %w", err)
	}
	defer rows.Close()

	snapshot := QueryStatsSnapshot{Queries: make([]QueryStatRow, 0)}

	for rows.Next() {
		var (
			query       QueryStatRow
			collectedAt time.Time
		)

		if err := rows.Scan(
			&collectedAt, &query.QueryID, &query.DatabaseName, &query.UserName, &query.Query,
			&query.Calls, &query.TotalExecTimeMs, &query.MeanExecTimeMs, &query.MinExecTimeMs, &query.MaxExecTimeMs, &query.RowsReturned,
			&query.SharedBlocksHit, &query.SharedBlocksRead, &query.SharedBlocksDirtied, &query.SharedBlocksWritten,
			&query.TempBlocksRead, &query.TempBlocksWritten,
		); err != nil {
			return QueryStatsSnapshot{}, fmt.Errorf("failed to scan query stat: %w", err)
		}

		snapshot.CollectedAt = collectedAt
		snapshot.Queries = append(snapshot.Queries, query)
	}

	if err := rows.Err(); err != nil {
		return QueryStatsSnapshot{}, fmt.Errorf("failed to read query stats: %w", err)
	}

	return snapshot, nil
}
