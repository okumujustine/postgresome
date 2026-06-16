package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

// TableStatRow is one table's statistics within a TableStatsSnapshot.
type TableStatRow struct {
	SchemaName string
	TableName  string

	LiveRows int64
	DeadRows int64

	SequentialScans    int64
	SequentialRowsRead int64

	IndexScans       int64
	IndexRowsFetched int64

	RowsInserted int64
	RowsUpdated  int64
	RowsDeleted  int64

	LastVacuumAt      *time.Time
	LastAutoVacuumAt  *time.Time
	LastAnalyzeAt     *time.Time
	LastAutoAnalyzeAt *time.Time

	VacuumCount      int64
	AutoVacuumCount  int64
	AnalyzeCount     int64
	AutoAnalyzeCount int64
}

// TableStatsSnapshot is the latest set of table statistics for a database
// instance.
type TableStatsSnapshot struct {
	CollectedAt time.Time
	Tables      []TableStatRow
}

type LatestTableStat struct {
	CollectedAt time.Time
	Table       TableStatRow
}

const upsertTableStatSQL = `
	INSERT INTO table_stats (
		database_instance_id, agent_id, collected_at, schema_name, table_name,
		live_rows, dead_rows, sequential_scans, sequential_rows_read,
		index_scans, index_rows_fetched, rows_inserted, rows_updated, rows_deleted,
		last_vacuum_at, last_autovacuum_at, last_analyze_at, last_autoanalyze_at,
		vacuum_count, autovacuum_count, analyze_count, autoanalyze_count
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9,
		$10, $11, $12, $13, $14,
		$15, $16, $17, $18,
		$19, $20, $21, $22
	)
	ON CONFLICT (database_instance_id, collected_at, schema_name, table_name) DO UPDATE SET
		agent_id = EXCLUDED.agent_id,
		live_rows = EXCLUDED.live_rows,
		dead_rows = EXCLUDED.dead_rows,
		sequential_scans = EXCLUDED.sequential_scans,
		sequential_rows_read = EXCLUDED.sequential_rows_read,
		index_scans = EXCLUDED.index_scans,
		index_rows_fetched = EXCLUDED.index_rows_fetched,
		rows_inserted = EXCLUDED.rows_inserted,
		rows_updated = EXCLUDED.rows_updated,
		rows_deleted = EXCLUDED.rows_deleted,
		last_vacuum_at = EXCLUDED.last_vacuum_at,
		last_autovacuum_at = EXCLUDED.last_autovacuum_at,
		last_analyze_at = EXCLUDED.last_analyze_at,
		last_autoanalyze_at = EXCLUDED.last_autoanalyze_at,
		vacuum_count = EXCLUDED.vacuum_count,
		autovacuum_count = EXCLUDED.autovacuum_count,
		analyze_count = EXCLUDED.analyze_count,
		autoanalyze_count = EXCLUDED.autoanalyze_count
`

// ReplaceTableStats stores a table statistics snapshot for a database
// instance. Historical snapshots are preserved; only rows for the same
// database instance, timestamp, schema, and table are replaced.
func ReplaceTableStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, tables []TableStatRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, table := range tables {
		_, err := tx.Exec(ctx, upsertTableStatSQL,
			databaseInstanceID, agentID, collectedAt, table.SchemaName, table.TableName,
			table.LiveRows, table.DeadRows, table.SequentialScans, table.SequentialRowsRead,
			table.IndexScans, table.IndexRowsFetched, table.RowsInserted, table.RowsUpdated, table.RowsDeleted,
			table.LastVacuumAt, table.LastAutoVacuumAt, table.LastAnalyzeAt, table.LastAutoAnalyzeAt,
			table.VacuumCount, table.AutoVacuumCount, table.AnalyzeCount, table.AutoAnalyzeCount,
		)
		if err != nil {
			return fmt.Errorf("failed to insert table stat for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit table stats transaction: %w", err)
	}

	return nil
}

const listTableStatsSQL = `
	SELECT collected_at, schema_name, table_name, live_rows, dead_rows,
	       sequential_scans, sequential_rows_read, index_scans, index_rows_fetched,
	       rows_inserted, rows_updated, rows_deleted,
	       last_vacuum_at, last_autovacuum_at, last_analyze_at, last_autoanalyze_at,
	       vacuum_count, autovacuum_count, analyze_count, autoanalyze_count
	FROM table_stats
	WHERE database_instance_id = $1
	  AND collected_at = (
		  SELECT max(collected_at)
		  FROM table_stats
		  WHERE database_instance_id = $1
	  )
	ORDER BY dead_rows DESC
`

const listTableStatsAtSQL = `
	SELECT collected_at, schema_name, table_name, live_rows, dead_rows,
	       sequential_scans, sequential_rows_read, index_scans, index_rows_fetched,
	       rows_inserted, rows_updated, rows_deleted,
	       last_vacuum_at, last_autovacuum_at, last_analyze_at, last_autoanalyze_at,
	       vacuum_count, autovacuum_count, analyze_count, autoanalyze_count
	FROM table_stats
	WHERE database_instance_id = $1
	  AND collected_at = $2
	ORDER BY dead_rows DESC
`

// ListTableStats returns the latest table statistics snapshot for a database
// instance. If no snapshot has been stored yet, it returns a zero-value
// snapshot (zero CollectedAt, empty Tables).
func ListTableStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (TableStatsSnapshot, error) {
	return listTableStatsSnapshot(ctx, pool, listTableStatsSQL, databaseInstanceID)
}

func ListTableStatsAt(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string, collectedAt time.Time) (TableStatsSnapshot, error) {
	return listTableStatsSnapshot(ctx, pool, listTableStatsAtSQL, databaseInstanceID, collectedAt)
}

func listTableStatsSnapshot(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (TableStatsSnapshot, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return TableStatsSnapshot{}, fmt.Errorf("failed to query table stats: %w", err)
	}
	defer rows.Close()

	snapshot := TableStatsSnapshot{Tables: make([]TableStatRow, 0)}

	for rows.Next() {
		var (
			table       TableStatRow
			collectedAt time.Time
		)

		if err := rows.Scan(
			&collectedAt, &table.SchemaName, &table.TableName, &table.LiveRows, &table.DeadRows,
			&table.SequentialScans, &table.SequentialRowsRead, &table.IndexScans, &table.IndexRowsFetched,
			&table.RowsInserted, &table.RowsUpdated, &table.RowsDeleted,
			&table.LastVacuumAt, &table.LastAutoVacuumAt, &table.LastAnalyzeAt, &table.LastAutoAnalyzeAt,
			&table.VacuumCount, &table.AutoVacuumCount, &table.AnalyzeCount, &table.AutoAnalyzeCount,
		); err != nil {
			return TableStatsSnapshot{}, fmt.Errorf("failed to scan table stat: %w", err)
		}

		snapshot.CollectedAt = collectedAt
		snapshot.Tables = append(snapshot.Tables, table)
	}

	if err := rows.Err(); err != nil {
		return TableStatsSnapshot{}, fmt.Errorf("failed to read table stats: %w", err)
	}

	return snapshot, nil
}

func (s TableStatsSnapshot) ToMetrics(databaseInstanceID string) metrics.TableStatsSnapshot {
	tables := make([]metrics.TableStats, len(s.Tables))
	for i, table := range s.Tables {
		tables[i] = metrics.TableStats{
			SchemaName:         table.SchemaName,
			TableName:          table.TableName,
			LiveRows:           table.LiveRows,
			DeadRows:           table.DeadRows,
			SequentialScans:    table.SequentialScans,
			SequentialRowsRead: table.SequentialRowsRead,
			IndexScans:         table.IndexScans,
			IndexRowsFetched:   table.IndexRowsFetched,
			RowsInserted:       table.RowsInserted,
			RowsUpdated:        table.RowsUpdated,
			RowsDeleted:        table.RowsDeleted,
			LastVacuumAt:       table.LastVacuumAt,
			LastAutoVacuumAt:   table.LastAutoVacuumAt,
			LastAnalyzeAt:      table.LastAnalyzeAt,
			LastAutoAnalyzeAt:  table.LastAutoAnalyzeAt,
			VacuumCount:        table.VacuumCount,
			AutoVacuumCount:    table.AutoVacuumCount,
			AnalyzeCount:       table.AnalyzeCount,
			AutoAnalyzeCount:   table.AutoAnalyzeCount,
		}
	}

	return metrics.TableStatsSnapshot{
		CollectedAt:        s.CollectedAt,
		DatabaseInstanceID: databaseInstanceID,
		Tables:             tables,
	}
}

const getTableStatAtOrBeforeSQL = `
	SELECT
		collected_at, schema_name, table_name, live_rows, dead_rows,
		sequential_scans, sequential_rows_read, index_scans, index_rows_fetched,
		rows_inserted, rows_updated, rows_deleted,
		last_vacuum_at, last_autovacuum_at, last_analyze_at, last_autoanalyze_at,
		vacuum_count, autovacuum_count, analyze_count, autoanalyze_count
	FROM table_stats
	WHERE database_instance_id = $1
	  AND schema_name = $2
	  AND table_name = $3
	  AND collected_at <= $4
	ORDER BY collected_at DESC
	LIMIT 1
`

func GetTableStatAtOrBefore(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, schemaName, tableName string, observedAt time.Time) (*LatestTableStat, error) {
	var result LatestTableStat
	err := pool.QueryRow(ctx, getTableStatAtOrBeforeSQL, databaseInstanceID, schemaName, tableName, observedAt).Scan(
		&result.CollectedAt,
		&result.Table.SchemaName,
		&result.Table.TableName,
		&result.Table.LiveRows,
		&result.Table.DeadRows,
		&result.Table.SequentialScans,
		&result.Table.SequentialRowsRead,
		&result.Table.IndexScans,
		&result.Table.IndexRowsFetched,
		&result.Table.RowsInserted,
		&result.Table.RowsUpdated,
		&result.Table.RowsDeleted,
		&result.Table.LastVacuumAt,
		&result.Table.LastAutoVacuumAt,
		&result.Table.LastAnalyzeAt,
		&result.Table.LastAutoAnalyzeAt,
		&result.Table.VacuumCount,
		&result.Table.AutoVacuumCount,
		&result.Table.AnalyzeCount,
		&result.Table.AutoAnalyzeCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get table stat for %s.%s at %s: %w", schemaName, tableName, observedAt.Format(time.RFC3339Nano), err)
	}

	return &result, nil
}

type TableStatHistoryPoint struct {
	CollectedAt      time.Time
	DeadRows         int64
	LiveRows         int64
	SequentialScans  int64
	LastAutoVacuumAt *time.Time
	LastVacuumAt     *time.Time
}

type TableStatHistoricalContext struct {
	Current          *TableStatHistoryPoint
	Previous         *TableStatHistoryPoint
	BaselineDeadRows float64
	BaselineSeqScans float64
	FirstAbnormalAt  *time.Time
}

const listTableStatHistorySQL = `
	SELECT collected_at, dead_rows, live_rows, sequential_scans, last_autovacuum_at, last_vacuum_at
	FROM table_stats
	WHERE database_instance_id = $1
	  AND schema_name = $2
	  AND table_name = $3
	  AND collected_at >= $4
	  AND collected_at <= $5
	ORDER BY collected_at DESC
	LIMIT $6
`

const getTableStatBaselineSQL = `
	SELECT
		COALESCE(avg(dead_rows::double precision), 0),
		COALESCE(avg(sequential_scans::double precision), 0)
	FROM table_stats
	WHERE database_instance_id = $1
	  AND schema_name = $2
	  AND table_name = $3
	  AND collected_at >= $4
	  AND collected_at < $5
`

const getTableStatFirstAbnormalSQL = `
	SELECT min(collected_at)
	FROM table_stats
	WHERE database_instance_id = $1
	  AND schema_name = $2
	  AND table_name = $3
	  AND collected_at >= $4
	  AND collected_at <= $5
	  AND dead_rows >= $6
`

// GetTableStatHistoricalContext returns current/previous snapshot points and
// a rolling baseline for a single table over the given lookback window.
func GetTableStatHistoricalContext(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, schemaName, tableName string, latestCollectedAt time.Time, lookback time.Duration, abnormalDeadRowsThreshold float64) (TableStatHistoricalContext, error) {
	since := latestCollectedAt.Add(-lookback)
	rows, err := pool.Query(ctx, listTableStatHistorySQL, databaseInstanceID, schemaName, tableName, since, latestCollectedAt, 12)
	if err != nil {
		return TableStatHistoricalContext{}, fmt.Errorf("failed to query table stat history for %s.%s: %w", schemaName, tableName, err)
	}
	defer rows.Close()

	points := make([]TableStatHistoryPoint, 0, 12)
	for rows.Next() {
		var point TableStatHistoryPoint
		if err := rows.Scan(&point.CollectedAt, &point.DeadRows, &point.LiveRows, &point.SequentialScans, &point.LastAutoVacuumAt, &point.LastVacuumAt); err != nil {
			return TableStatHistoricalContext{}, fmt.Errorf("failed to scan table stat history for %s.%s: %w", schemaName, tableName, err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return TableStatHistoricalContext{}, fmt.Errorf("failed to read table stat history for %s.%s: %w", schemaName, tableName, err)
	}

	var context TableStatHistoricalContext
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
	if err := pool.QueryRow(ctx, getTableStatBaselineSQL, databaseInstanceID, schemaName, tableName, since, baselineEnd).Scan(&context.BaselineDeadRows, &context.BaselineSeqScans); err != nil {
		return TableStatHistoricalContext{}, fmt.Errorf("failed to query table baseline for %s.%s: %w", schemaName, tableName, err)
	}

	if abnormalDeadRowsThreshold > 0 {
		var firstAt *time.Time
		if err := pool.QueryRow(ctx, getTableStatFirstAbnormalSQL, databaseInstanceID, schemaName, tableName, since, latestCollectedAt, abnormalDeadRowsThreshold).Scan(&firstAt); err != nil {
			return TableStatHistoricalContext{}, fmt.Errorf("failed to query first abnormal table stat for %s.%s: %w", schemaName, tableName, err)
		}
		context.FirstAbnormalAt = firstAt
	}

	return context, nil
}

// ListTableStatHistory returns a bounded, newest-last series for one table
// resource inside the requested window.
func ListTableStatHistory(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, schemaName, tableName string, since, until time.Time, limit int) ([]TableStatHistoryPoint, error) {
	rows, err := pool.Query(ctx, listTableStatHistorySQL, databaseInstanceID, schemaName, tableName, since, until, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query table stat history for %s.%s: %w", schemaName, tableName, err)
	}
	defer rows.Close()

	points := make([]TableStatHistoryPoint, 0, limit)
	for rows.Next() {
		var point TableStatHistoryPoint
		if err := rows.Scan(&point.CollectedAt, &point.DeadRows, &point.LiveRows, &point.SequentialScans, &point.LastAutoVacuumAt, &point.LastVacuumAt); err != nil {
			return nil, fmt.Errorf("failed to scan table stat history for %s.%s: %w", schemaName, tableName, err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read table stat history for %s.%s: %w", schemaName, tableName, err)
	}

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
