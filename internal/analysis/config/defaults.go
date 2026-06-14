package config

// Rule keys, in "category.problem" format. These are the canonical
// identifiers used for Finding.RuleKey, fingerprinting, and configuration
// lookups.
const (
	RuleKeyHighConnections  = "database.high_connections"
	RuleKeyLowCacheHitRatio = "database.low_cache_hit_ratio"
	RuleKeyHighRollbackRate = "database.high_rollback_rate"

	RuleKeyIdleConnections  = "activity.idle_connections"
	RuleKeyLongRunningQuery = "activity.long_running_query"
	RuleKeyBlockedQuery     = "activity.blocked_query"

	RuleKeyHighDeadRows       = "table.high_dead_rows"
	RuleKeyAutovacuumLag      = "table.autovacuum_lag"
	RuleKeyHighSequentialScan = "table.high_sequential_scan"

	RuleKeySlowQuery      = "query.slow_query"
	RuleKeyExpensiveQuery = "query.expensive_query"
	RuleKeyDiskHeavyQuery = "query.disk_heavy_query"
)

// DefaultAnalysisConfig returns the built-in analyzer rule configuration.
// It requires no external configuration file and preserves the thresholds,
// severities, and finding content that the analyzer used before rules
// became configurable.
func DefaultAnalysisConfig() AnalysisConfig {
	return AnalysisConfig{
		Rules: map[string]RuleConfig{
			RuleKeyHighConnections: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning":  100,
					"critical": 500,
				},
				Title:          "High database connections detected",
				Description:    "Database has an unusually high number of active connections, which may indicate connection leaks, missing pooling, or traffic spikes.",
				Recommendation: "Investigate connection spikes, idle connections, and connection pooling configuration.",
			},
			RuleKeyLowCacheHitRatio: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning":  0.95,
					"critical": 0.90,
				},
				Title:          "Low cache hit ratio detected",
				Description:    "PostgreSQL is reading frequently from disk instead of serving data from cache.",
				Recommendation: "Investigate missing indexes, inefficient queries, and shared_buffers configuration.",
			},
			RuleKeyHighRollbackRate: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning":  0.05,
					"critical": 0.15,
				},
				Title:          "High transaction rollback rate detected",
				Description:    "A large percentage of database transactions are being rolled back instead of committed.",
				Recommendation: "Investigate application errors, constraint violations, deadlocks, or timeout issues causing rollbacks.",
			},
			RuleKeyIdleConnections: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_ratio":                0.50,
					"critical_ratio":               0.80,
					"idle_in_transaction_critical": 5,
				},
				Title:          "High idle connection usage detected",
				Description:    "A large share of database sessions are idle or idle in a transaction.",
				Recommendation: "Review application connection pooling and check for connection leaks or long-held idle transactions.",
			},
			RuleKeyLongRunningQuery: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_minutes":  1,
					"critical_minutes": 10,
				},
				Title:          "Long running query detected",
				Description:    "A query has been executing for an unusually long time.",
				Recommendation: "Review query execution plan, indexes, and query complexity; investigate locks if the query appears stuck.",
			},
			RuleKeyBlockedQuery: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning":  1,
					"critical": 5,
				},
				Title:          "Blocked database queries detected",
				Description:    "Database sessions are waiting on locks held by other transactions.",
				Recommendation: "Identify blocking queries and review long-running transactions holding locks.",
			},
			RuleKeyHighDeadRows: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"dead_rows": 100000,
				},
				Title:          "High dead rows detected",
				Description:    "Table contains a large number of dead rows.",
				Recommendation: "Review autovacuum behavior and consider VACUUM ANALYZE.",
			},
			RuleKeyAutovacuumLag: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_dead_rows":       10000,
					"critical_dead_rows":      100000,
					"critical_dead_row_ratio": 0.30,
					"stale_hours":             24,
				},
				Title:          "Autovacuum may be lagging",
				Description:    "Table has accumulated dead rows and has not been autovacuumed recently.",
				Recommendation: "Review autovacuum configuration, long-running transactions, and consider a manual VACUUM during a maintenance window.",
			},
			RuleKeyHighSequentialScan: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_ratio":  0.50,
					"warning_rows":   100000,
					"critical_ratio": 0.80,
					"critical_rows":  1000000,
				},
				Title:          "High sequential scan activity detected",
				Description:    "Table is using sequential scans for a large share of reads.",
				Recommendation: "Review query patterns and consider whether frequently filtered columns need indexes.",
			},
			RuleKeySlowQuery: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_ms":  100,
					"critical_ms": 1000,
				},
				Title:          "Slow query detected",
				Description:    "Query has a high mean execution time.",
				Recommendation: "Review query execution plan, indexes, joins, and filters.",
			},
			RuleKeyExpensiveQuery: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_ms":  10000,
					"critical_ms": 60000,
				},
				Title:          "High total query cost detected",
				Description:    "Query consumes a large amount of total database time across its executions.",
				Recommendation: "Review frequency, indexing, and whether results can be cached or optimized.",
			},
			RuleKeyDiskHeavyQuery: {
				Enabled:  true,
				Severity: "warning",
				Thresholds: map[string]float64{
					"warning_ratio":  0.10,
					"critical_ratio": 0.30,
				},
				Title:          "Disk-heavy query detected",
				Description:    "Query reads a large portion of its blocks from disk rather than cache.",
				Recommendation: "Investigate missing indexes, inefficient scans, and memory/cache behavior.",
			},
		},
	}
}
