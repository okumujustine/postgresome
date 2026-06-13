package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

const deleteTableStatsSQL = `
	DELETE FROM table_stats WHERE database_instance_id = $1
`

const insertTableStatSQL = `
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
`

// ReplaceTableStats replaces the stored table statistics snapshot for a
// database instance with the given rows, in a single transaction.
func ReplaceTableStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, collectedAt time.Time, tables []TableStatRow) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, deleteTableStatsSQL, databaseInstanceID); err != nil {
		return fmt.Errorf("failed to delete existing table stats: %w", err)
	}

	for _, table := range tables {
		_, err := tx.Exec(ctx, insertTableStatSQL,
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
	ORDER BY dead_rows DESC
`

// ListTableStats returns the latest table statistics snapshot for a database
// instance. If no snapshot has been stored yet, it returns a zero-value
// snapshot (zero CollectedAt, empty Tables).
func ListTableStats(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID string) (TableStatsSnapshot, error) {
	rows, err := pool.Query(ctx, listTableStatsSQL, databaseInstanceID)
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
