package checkup

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

const (
	defaultQueryStatsLimit  = 50
	defaultExplainPlanLimit = 5
)

var placeholderPattern = regexp.MustCompile(`\$\d+`)

type Collector struct{}

type CollectedEvidence struct {
	CollectedAt  time.Time
	DatabaseName string
	Database     metrics.DatabaseStats
	MetricPoints []metrics.MetricPoint
	Activities   []repository.ActivityStatRow
	Tables       []repository.TableStatRow
	Queries      []repository.QueryStatRow
	Plans        []repository.ExplainPlanRow
	Warnings     []string
}

type explainEnvelope struct {
	Plan metrics.PlanNode `json:"Plan"`
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Collect(ctx context.Context, sourceID, databaseInstanceID, connectionURI string) (*CollectedEvidence, error) {
	targetPool, err := connectTargetDatabase(ctx, connectionURI)
	if err != nil {
		return nil, err
	}
	defer targetPool.Close()

	collectedAt := time.Now().UTC()

	databaseStats, err := collectDatabaseStats(ctx, targetPool)
	if err != nil {
		return nil, err
	}

	result := &CollectedEvidence{
		CollectedAt:  collectedAt,
		DatabaseName: databaseStats.DatabaseName,
		Database:     databaseStats,
		MetricPoints: databaseMetricPoints(sourceID, databaseInstanceID, collectedAt, databaseStats),
		Activities:   make([]repository.ActivityStatRow, 0),
		Tables:       make([]repository.TableStatRow, 0),
		Queries:      make([]repository.QueryStatRow, 0),
		Plans:        make([]repository.ExplainPlanRow, 0),
		Warnings:     make([]string, 0),
	}

	activities, err := collectActivityStats(ctx, targetPool)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("activity snapshot skipped: %v", err))
	} else {
		result.Activities = activities
	}

	tables, err := collectTableStats(ctx, targetPool)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("table statistics skipped: %v", err))
	} else {
		result.Tables = tables
	}

	queries, err := collectQueryStats(ctx, targetPool, defaultQueryStatsLimit)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("query statistics skipped: %v", err))
	} else {
		result.Queries = queries

		plans, planWarnings := collectExplainPlans(ctx, targetPool, queries, defaultExplainPlanLimit)
		result.Plans = plans
		result.Warnings = append(result.Warnings, planWarnings...)
	}

	return result, nil
}

func connectTargetDatabase(ctx context.Context, connectionURI string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	config.MaxConns = 4
	config.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to source database: %w", err)
	}

	return pool, nil
}

const collectDatabaseStatsSQL = `
	SELECT
		current_database(),
		(
			SELECT count(*)
			FROM pg_stat_activity
			WHERE datname = current_database()
			  AND state = 'active'
		),
		xact_commit,
		xact_rollback,
		blks_read,
		blks_hit,
		tup_returned,
		tup_fetched,
		tup_inserted,
		tup_updated,
		tup_deleted
	FROM pg_stat_database
	WHERE datname = current_database()
`

func collectDatabaseStats(ctx context.Context, pool *pgxpool.Pool) (metrics.DatabaseStats, error) {
	var stats metrics.DatabaseStats

	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := pool.QueryRow(queryCtx, collectDatabaseStatsSQL).Scan(
		&stats.DatabaseName,
		&stats.ActiveConnections,
		&stats.TransactionCommits,
		&stats.TransactionRollbacks,
		&stats.BlocksReadFromDisk,
		&stats.BlocksHitInCache,
		&stats.RowsScanned,
		&stats.RowsFetched,
		&stats.RowsInserted,
		&stats.RowsUpdated,
		&stats.RowsDeleted,
	)
	if err != nil {
		return metrics.DatabaseStats{}, fmt.Errorf("failed to collect database stats from pg_stat_database: %w", err)
	}

	return stats, nil
}

const collectActivityStatsSQL = `
	SELECT
		COALESCE(datname, current_database()),
		pid,
		COALESCE(usename, ''),
		COALESCE(application_name, ''),
		COALESCE(client_addr::text, ''),
		COALESCE(state, ''),
		COALESCE(query, ''),
		COALESCE(wait_event_type, ''),
		COALESCE(wait_event, ''),
		backend_start,
		query_start,
		state_change
	FROM pg_stat_activity
	WHERE datname = current_database()
	  AND pid <> pg_backend_pid()
	ORDER BY
		CASE WHEN state = 'active' THEN 0 ELSE 1 END,
		query_start DESC NULLS LAST,
		pid ASC
`

func collectActivityStats(ctx context.Context, pool *pgxpool.Pool) ([]repository.ActivityStatRow, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := pool.Query(queryCtx, collectActivityStatsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to collect pg_stat_activity rows: %w", err)
	}
	defer rows.Close()

	activities := make([]repository.ActivityStatRow, 0)
	for rows.Next() {
		var row repository.ActivityStatRow
		if err := rows.Scan(
			&row.DatabaseName,
			&row.ProcessID,
			&row.UserName,
			&row.ApplicationName,
			&row.ClientAddress,
			&row.State,
			&row.Query,
			&row.WaitEventType,
			&row.WaitEvent,
			&row.BackendStartedAt,
			&row.QueryStartedAt,
			&row.StateChangedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pg_stat_activity row: %w", err)
		}
		activities = append(activities, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pg_stat_activity rows: %w", err)
	}

	return activities, nil
}

const collectTableStatsSQL = `
	SELECT
		schemaname,
		relname,
		n_live_tup,
		n_dead_tup,
		seq_scan,
		seq_tup_read,
		idx_scan,
		idx_tup_fetch,
		n_tup_ins,
		n_tup_upd,
		n_tup_del,
		last_vacuum,
		last_autovacuum,
		last_analyze,
		last_autoanalyze,
		vacuum_count,
		autovacuum_count,
		analyze_count,
		autoanalyze_count
	FROM pg_stat_user_tables
	ORDER BY n_dead_tup DESC, seq_scan DESC, relname ASC
`

func collectTableStats(ctx context.Context, pool *pgxpool.Pool) ([]repository.TableStatRow, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := pool.Query(queryCtx, collectTableStatsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to collect pg_stat_user_tables rows: %w", err)
	}
	defer rows.Close()

	tables := make([]repository.TableStatRow, 0)
	for rows.Next() {
		var row repository.TableStatRow
		if err := rows.Scan(
			&row.SchemaName,
			&row.TableName,
			&row.LiveRows,
			&row.DeadRows,
			&row.SequentialScans,
			&row.SequentialRowsRead,
			&row.IndexScans,
			&row.IndexRowsFetched,
			&row.RowsInserted,
			&row.RowsUpdated,
			&row.RowsDeleted,
			&row.LastVacuumAt,
			&row.LastAutoVacuumAt,
			&row.LastAnalyzeAt,
			&row.LastAutoAnalyzeAt,
			&row.VacuumCount,
			&row.AutoVacuumCount,
			&row.AnalyzeCount,
			&row.AutoAnalyzeCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pg_stat_user_tables row: %w", err)
		}
		tables = append(tables, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pg_stat_user_tables rows: %w", err)
	}

	return tables, nil
}

const pgStatStatementsAvailableSQL = `
	SELECT EXISTS (
		SELECT 1
		FROM pg_extension
		WHERE extname = 'pg_stat_statements'
	)
`

const collectQueryStatsSQL = `
	SELECT
		COALESCE(queryid::text, md5(query)),
		current_database(),
		COALESCE(usename, ''),
		query,
		calls,
		total_exec_time,
		mean_exec_time,
		min_exec_time,
		max_exec_time,
		rows,
		shared_blks_hit,
		shared_blks_read,
		shared_blks_dirtied,
		shared_blks_written,
		temp_blks_read,
		temp_blks_written
	FROM pg_stat_statements
	WHERE dbid = (
		SELECT oid
		FROM pg_database
		WHERE datname = current_database()
	)
	ORDER BY total_exec_time DESC, calls DESC
	LIMIT $1
`

func collectQueryStats(ctx context.Context, pool *pgxpool.Pool, limit int) ([]repository.QueryStatRow, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var available bool
	if err := pool.QueryRow(queryCtx, pgStatStatementsAvailableSQL).Scan(&available); err != nil {
		return nil, fmt.Errorf("failed to check pg_stat_statements availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("pg_stat_statements is not installed on the source database")
	}

	rows, err := pool.Query(queryCtx, collectQueryStatsSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to collect pg_stat_statements rows: %w", err)
	}
	defer rows.Close()

	queries := make([]repository.QueryStatRow, 0)
	for rows.Next() {
		var row repository.QueryStatRow
		if err := rows.Scan(
			&row.QueryID,
			&row.DatabaseName,
			&row.UserName,
			&row.Query,
			&row.Calls,
			&row.TotalExecTimeMs,
			&row.MeanExecTimeMs,
			&row.MinExecTimeMs,
			&row.MaxExecTimeMs,
			&row.RowsReturned,
			&row.SharedBlocksHit,
			&row.SharedBlocksRead,
			&row.SharedBlocksDirtied,
			&row.SharedBlocksWritten,
			&row.TempBlocksRead,
			&row.TempBlocksWritten,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pg_stat_statements row: %w", err)
		}
		queries = append(queries, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pg_stat_statements rows: %w", err)
	}

	return queries, nil
}

func collectExplainPlans(ctx context.Context, pool *pgxpool.Pool, queries []repository.QueryStatRow, limit int) ([]repository.ExplainPlanRow, []string) {
	plans := make([]repository.ExplainPlanRow, 0, limit)
	warnings := make([]string, 0)

	for _, query := range queries {
		if len(plans) >= limit {
			break
		}
		if !isExplainableQuery(query.Query) {
			continue
		}

		queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		var rawPlan []byte
		err := pool.QueryRow(queryCtx, "EXPLAIN (FORMAT JSON) "+query.Query).Scan(&rawPlan)
		cancel()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("explain skipped for query %s: %v", query.QueryID, err))
			continue
		}

		var envelope []explainEnvelope
		if err := json.Unmarshal(rawPlan, &envelope); err != nil || len(envelope) == 0 {
			warnings = append(warnings, fmt.Sprintf("explain decode skipped for query %s", query.QueryID))
			continue
		}

		plans = append(plans, repository.ExplainPlanRow{
			QueryID: query.QueryID,
			Query:   query.Query,
			Root:    envelope[0].Plan,
		})
	}

	return plans, warnings
}

func isExplainableQuery(query string) bool {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, ";") {
		return false
	}
	if placeholderPattern.MatchString(trimmed) {
		return false
	}

	upper := strings.ToUpper(trimmed)
	return strings.HasPrefix(upper, "SELECT ") || strings.HasPrefix(upper, "WITH ")
}

func databaseMetricPoints(sourceID, databaseInstanceID string, collectedAt time.Time, stats metrics.DatabaseStats) []metrics.MetricPoint {
	dimensions := map[string]string{
		"source_id":            sourceID,
		"database_instance_id": databaseInstanceID,
		"database_name":        stats.DatabaseName,
	}

	return []metrics.MetricPoint{
		newMetricPoint("active_connections", "Active Connections", "count", "connections", collectedAt, float64(stats.ActiveConnections), dimensions),
		newMetricPoint("transaction_commits", "Transaction Commits", "count", "transactions", collectedAt, float64(stats.TransactionCommits), dimensions),
		newMetricPoint("transaction_rollbacks", "Transaction Rollbacks", "count", "transactions", collectedAt, float64(stats.TransactionRollbacks), dimensions),
		newMetricPoint("blocks_read_from_disk", "Blocks Read From Disk", "count", "cache", collectedAt, float64(stats.BlocksReadFromDisk), dimensions),
		newMetricPoint("blocks_hit_in_cache", "Blocks Hit In Cache", "count", "cache", collectedAt, float64(stats.BlocksHitInCache), dimensions),
		newMetricPoint("rows_scanned", "Rows Scanned", "count", "reads", collectedAt, float64(stats.RowsScanned), dimensions),
		newMetricPoint("rows_fetched", "Rows Fetched", "count", "reads", collectedAt, float64(stats.RowsFetched), dimensions),
		newMetricPoint("rows_inserted", "Rows Inserted", "count", "writes", collectedAt, float64(stats.RowsInserted), dimensions),
		newMetricPoint("rows_updated", "Rows Updated", "count", "writes", collectedAt, float64(stats.RowsUpdated), dimensions),
		newMetricPoint("rows_deleted", "Rows Deleted", "count", "writes", collectedAt, float64(stats.RowsDeleted), dimensions),
	}
}

func newMetricPoint(key, label, unit, category string, collectedAt time.Time, value float64, dimensions map[string]string) metrics.MetricPoint {
	pointDimensions := make(map[string]string, len(dimensions))
	for k, v := range dimensions {
		pointDimensions[k] = v
	}

	return metrics.MetricPoint{
		Key:         key,
		Label:       label,
		Value:       value,
		Unit:        unit,
		Category:    category,
		CollectedAt: collectedAt,
		Dimensions:  pointDimensions,
	}
}
