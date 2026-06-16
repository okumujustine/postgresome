package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
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

type LatestQueryStat struct {
	CollectedAt time.Time
	Query       QueryStatRow
}

const upsertQueryStatSQL = `
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
	ON CONFLICT (database_instance_id, collected_at, query_id) DO UPDATE SET
		agent_id = EXCLUDED.agent_id,
		database_name = EXCLUDED.database_name,
		user_name = EXCLUDED.user_name,
		query = EXCLUDED.query,
		calls = EXCLUDED.calls,
		total_exec_time_ms = EXCLUDED.total_exec_time_ms,
		mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
		min_exec_time_ms = EXCLUDED.min_exec_time_ms,
		max_exec_time_ms = EXCLUDED.max_exec_time_ms,
		rows_returned = EXCLUDED.rows_returned,
		shared_blocks_hit = EXCLUDED.shared_blocks_hit,
		shared_blocks_read = EXCLUDED.shared_blocks_read,
		shared_blocks_dirtied = EXCLUDED.shared_blocks_dirtied,
		shared_blocks_written = EXCLUDED.shared_blocks_written,
		temp_blocks_read = EXCLUDED.temp_blocks_read,
		temp_blocks_written = EXCLUDED.temp_blocks_written
`

// ReplaceQueryStats stores a query statistics snapshot for a database
// instance. Historical snapshots are preserved; only rows for the same
// database instance, timestamp, and query id are replaced.
func ReplaceQueryStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, queries []QueryStatRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, query := range queries {
		_, err := tx.Exec(ctx, upsertQueryStatSQL,
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
	  AND collected_at = (
		  SELECT max(collected_at)
		  FROM query_stats
		  WHERE database_instance_id = $1
	  )
	ORDER BY total_exec_time_ms DESC
`

const listQueryStatsAtSQL = `
	SELECT collected_at, query_id, database_name, user_name, query,
	       calls, total_exec_time_ms, mean_exec_time_ms, min_exec_time_ms, max_exec_time_ms, rows_returned,
	       shared_blocks_hit, shared_blocks_read, shared_blocks_dirtied, shared_blocks_written,
	       temp_blocks_read, temp_blocks_written
	FROM query_stats
	WHERE database_instance_id = $1
	  AND collected_at = $2
	ORDER BY total_exec_time_ms DESC
`

// ListQueryStats returns the latest query statistics snapshot for a database
// instance. If no snapshot has been stored yet, it returns a zero-value
// snapshot (zero CollectedAt, empty Queries).
func ListQueryStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (QueryStatsSnapshot, error) {
	return listQueryStatsSnapshot(ctx, pool, listQueryStatsSQL, databaseInstanceID)
}

func ListQueryStatsAt(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string, collectedAt time.Time) (QueryStatsSnapshot, error) {
	return listQueryStatsSnapshot(ctx, pool, listQueryStatsAtSQL, databaseInstanceID, collectedAt)
}

func listQueryStatsSnapshot(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (QueryStatsSnapshot, error) {
	rows, err := pool.Query(ctx, sql, args...)
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

func (s QueryStatsSnapshot) ToMetrics(databaseInstanceID string) metrics.QueryStatsSnapshot {
	queries := make([]metrics.QueryStats, len(s.Queries))
	for i, query := range s.Queries {
		queries[i] = metrics.QueryStats{
			QueryID:              query.QueryID,
			DatabaseName:         query.DatabaseName,
			UserName:             query.UserName,
			Query:                query.Query,
			Calls:                query.Calls,
			TotalExecutionTimeMs: query.TotalExecTimeMs,
			MeanExecutionTimeMs:  query.MeanExecTimeMs,
			MinExecutionTimeMs:   query.MinExecTimeMs,
			MaxExecutionTimeMs:   query.MaxExecTimeMs,
			RowsReturned:         query.RowsReturned,
			SharedBlocksHit:      query.SharedBlocksHit,
			SharedBlocksRead:     query.SharedBlocksRead,
			SharedBlocksDirtied:  query.SharedBlocksDirtied,
			SharedBlocksWritten:  query.SharedBlocksWritten,
			TempBlocksRead:       query.TempBlocksRead,
			TempBlocksWritten:    query.TempBlocksWritten,
		}
	}

	return metrics.QueryStatsSnapshot{
		CollectedAt:        s.CollectedAt,
		DatabaseInstanceID: databaseInstanceID,
		Queries:            queries,
	}
}

const getQueryStatAtOrBeforeSQL = `
	SELECT
		collected_at, query_id, database_name, user_name, query,
		calls, total_exec_time_ms, mean_exec_time_ms, min_exec_time_ms, max_exec_time_ms, rows_returned,
		shared_blocks_hit, shared_blocks_read, shared_blocks_dirtied, shared_blocks_written,
		temp_blocks_read, temp_blocks_written
	FROM query_stats
	WHERE database_instance_id = $1
	  AND query_id = $2
	  AND collected_at <= $3
	ORDER BY collected_at DESC
	LIMIT 1
`

func GetQueryStatAtOrBefore(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, queryID string, observedAt time.Time) (*LatestQueryStat, error) {
	var result LatestQueryStat
	err := pool.QueryRow(ctx, getQueryStatAtOrBeforeSQL, databaseInstanceID, queryID, observedAt).Scan(
		&result.CollectedAt,
		&result.Query.QueryID,
		&result.Query.DatabaseName,
		&result.Query.UserName,
		&result.Query.Query,
		&result.Query.Calls,
		&result.Query.TotalExecTimeMs,
		&result.Query.MeanExecTimeMs,
		&result.Query.MinExecTimeMs,
		&result.Query.MaxExecTimeMs,
		&result.Query.RowsReturned,
		&result.Query.SharedBlocksHit,
		&result.Query.SharedBlocksRead,
		&result.Query.SharedBlocksDirtied,
		&result.Query.SharedBlocksWritten,
		&result.Query.TempBlocksRead,
		&result.Query.TempBlocksWritten,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get query stat for %s at %s: %w", queryID, observedAt.Format(time.RFC3339Nano), err)
	}

	return &result, nil
}

type QueryStatHistoryPoint struct {
	CollectedAt      time.Time
	Calls            int64
	MeanExecTimeMs   float64
	TotalExecTimeMs  float64
	SharedBlocksRead int64
	SharedBlocksHit  int64
}

type QueryStatHistoricalContext struct {
	Current            *QueryStatHistoryPoint
	Previous           *QueryStatHistoryPoint
	BaselineMeanExecMs float64
	BaselineCalls      float64
	BaselineBlocksRead float64
	FirstAbnormalAt    *time.Time
}

const listQueryStatHistorySQL = `
	SELECT collected_at, calls, mean_exec_time_ms, total_exec_time_ms, shared_blocks_read, shared_blocks_hit
	FROM query_stats
	WHERE database_instance_id = $1
	  AND query_id = $2
	  AND collected_at >= $3
	  AND collected_at <= $4
	ORDER BY collected_at DESC
	LIMIT $5
`

const getQueryStatBaselineSQL = `
	SELECT
		COALESCE(avg(mean_exec_time_ms), 0),
		COALESCE(avg(calls::double precision), 0),
		COALESCE(avg(shared_blocks_read::double precision), 0)
	FROM query_stats
	WHERE database_instance_id = $1
	  AND query_id = $2
	  AND collected_at >= $3
	  AND collected_at < $4
`

const getQueryStatFirstAbnormalSQL = `
	SELECT min(collected_at)
	FROM query_stats
	WHERE database_instance_id = $1
	  AND query_id = $2
	  AND collected_at >= $3
	  AND collected_at <= $4
	  AND mean_exec_time_ms >= $5
`

// GetQueryStatHistoricalContext returns current/previous snapshot points and
// a rolling baseline for one query over the given lookback window.
func GetQueryStatHistoricalContext(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, queryID string, latestCollectedAt time.Time, lookback time.Duration, abnormalMeanExecThreshold float64) (QueryStatHistoricalContext, error) {
	since := latestCollectedAt.Add(-lookback)
	rows, err := pool.Query(ctx, listQueryStatHistorySQL, databaseInstanceID, queryID, since, latestCollectedAt, 12)
	if err != nil {
		return QueryStatHistoricalContext{}, fmt.Errorf("failed to query query stat history for %s: %w", queryID, err)
	}
	defer rows.Close()

	points := make([]QueryStatHistoryPoint, 0, 12)
	for rows.Next() {
		var point QueryStatHistoryPoint
		if err := rows.Scan(&point.CollectedAt, &point.Calls, &point.MeanExecTimeMs, &point.TotalExecTimeMs, &point.SharedBlocksRead, &point.SharedBlocksHit); err != nil {
			return QueryStatHistoricalContext{}, fmt.Errorf("failed to scan query stat history for %s: %w", queryID, err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return QueryStatHistoricalContext{}, fmt.Errorf("failed to read query stat history for %s: %w", queryID, err)
	}

	var context QueryStatHistoricalContext
	if len(points) > 0 {
		context.Current = &points[0]
	}
	if len(points) > 1 {
		context.Previous = &points[1]
	}

	baselineEnd := latestCollectedAt
	if context.Current != nil {
		baselineEnd = context.Current.CollectedAt
	}
	if err := pool.QueryRow(ctx, getQueryStatBaselineSQL, databaseInstanceID, queryID, since, baselineEnd).Scan(&context.BaselineMeanExecMs, &context.BaselineCalls, &context.BaselineBlocksRead); err != nil {
		return QueryStatHistoricalContext{}, fmt.Errorf("failed to query query baseline for %s: %w", queryID, err)
	}

	if abnormalMeanExecThreshold > 0 {
		var firstAt *time.Time
		if err := pool.QueryRow(ctx, getQueryStatFirstAbnormalSQL, databaseInstanceID, queryID, since, latestCollectedAt, abnormalMeanExecThreshold).Scan(&firstAt); err != nil {
			return QueryStatHistoricalContext{}, fmt.Errorf("failed to query first abnormal query stat for %s: %w", queryID, err)
		}
		context.FirstAbnormalAt = firstAt
	}

	return context, nil
}

// ListQueryStatHistory returns a bounded, newest-last series for one query
// resource inside the requested window.
func ListQueryStatHistory(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, queryID string, since, until time.Time, limit int) ([]QueryStatHistoryPoint, error) {
	rows, err := pool.Query(ctx, listQueryStatHistorySQL, databaseInstanceID, queryID, since, until, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query query stat history for %s: %w", queryID, err)
	}
	defer rows.Close()

	points := make([]QueryStatHistoryPoint, 0, limit)
	for rows.Next() {
		var point QueryStatHistoryPoint
		if err := rows.Scan(&point.CollectedAt, &point.Calls, &point.MeanExecTimeMs, &point.TotalExecTimeMs, &point.SharedBlocksRead, &point.SharedBlocksHit); err != nil {
			return nil, fmt.Errorf("failed to scan query stat history for %s: %w", queryID, err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read query stat history for %s: %w", queryID, err)
	}

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
