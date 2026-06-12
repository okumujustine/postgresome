package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/agent/client"
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/rules"
	"github.com/okumujustine/postgresome/internal/collector"
	"github.com/okumujustine/postgresome/internal/collector/activity"
	"github.com/okumujustine/postgresome/internal/collector/extensions"
	"github.com/okumujustine/postgresome/internal/collector/queries"
	"github.com/okumujustine/postgresome/internal/collector/tables"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	maxSampleSessions     = 5
	maxQueryPreviewLength = 80
	maxTableSummaryItems  = 5
	maxQuerySummaryItems  = 5
)

type Config struct {
	DatabaseURL        string
	Interval           time.Duration
	AgentID            string
	AgentName          string
	AgentEnvironment   string
	APIBaseURL         string
	DatabaseInstanceID string
	LogLevel           string
}

type Runner struct {
	cfg      Config
	logLevel logLevel
}

func NewRunner(cfg Config) *Runner {
	return &Runner{
		cfg:      cfg,
		logLevel: parseLogLevel(cfg.LogLevel),
	}
}

func (r *Runner) Start(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, r.cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	postgresCollector := collector.NewPostgresCollector(pool)
	activityCollector := activity.NewActivityCollector(pool)
	tableStatsCollector := tables.NewTableStatsCollector(pool)
	extensionCollector := extensions.NewExtensionCollector(pool)
	queryStatsCollector := queries.NewQueryStatsCollector(pool)

	info, err := postgresCollector.GetDatabaseInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Connected to PostgreSQL successfully (version: %s)\n", info.Version)

	apiClient := client.NewAgentAPIClient(r.cfg.APIBaseURL)

	if err := r.registerWithAPI(ctx, pool, apiClient, info.Version); err != nil {
		return err
	}

	pgStatStatementsEnabled, err := extensionCollector.IsPgStatStatementsEnabled(ctx)
	if err != nil {
		fmt.Printf("failed to check pg_stat_statements availability: %v\n", err)
	} else if pgStatStatementsEnabled {
		fmt.Println("pg_stat_statements enabled - query analysis available")
	} else {
		fmt.Println("pg_stat_statements not enabled - skipping query analysis")
	}

	fmt.Printf("Collection interval: %s\n", r.cfg.Interval)

	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	engine := analysis.NewEngine()
	engine.RegisterMetricRules(rules.HighConnectionRule{}, rules.LowCacheHitRatioRule{}, rules.HighRollbackRateRule{})
	engine.RegisterActivityRules(rules.IdleConnectionRule{}, rules.LongRunningQueryRule{}, rules.BlockedQueryRule{})
	engine.RegisterTableRules(rules.AutovacuumLagRule{}, rules.HighSequentialScanRule{})
	engine.RegisterQueryRules(rules.SlowQueryRule{}, rules.ExpensiveQueryRule{}, rules.DiskHeavyQueryRule{})

	var previousStats *metrics.DatabaseStats

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			collectionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

			stats, err := postgresCollector.GetDatabaseStats(collectionCtx)
			cancel()

			if err != nil {
				fmt.Printf("failed to collect database stats: %v\n", err)
				continue
			}

			if r.logLevel == logLevelDebug {
				printDatabaseStats(stats)
			}

			activityCtx, activityCancel := context.WithTimeout(ctx, 5*time.Second)
			activitySnapshot, err := activityCollector.GetDatabaseActivity(activityCtx)
			activityCancel()

			if err != nil {
				fmt.Printf("failed to collect database activity: %v\n", err)
			} else {
				activitySnapshot.DatabaseInstanceID = stats.DatabaseName
				if r.logLevel == logLevelDebug {
					printDatabaseActivity(activitySnapshot)
				}
			}

			tableStatsCtx, tableStatsCancel := context.WithTimeout(ctx, 5*time.Second)
			tableStatsSnapshot, err := tableStatsCollector.GetTableStats(tableStatsCtx)
			tableStatsCancel()

			if err != nil {
				fmt.Printf("failed to collect table stats: %v\n", err)
			} else {
				tableStatsSnapshot.DatabaseInstanceID = stats.DatabaseName
				if r.logLevel == logLevelDebug {
					printTableHealthSummary(tableStatsSnapshot)
				}
			}

			var queryStatsSnapshot *metrics.QueryStatsSnapshot

			if pgStatStatementsEnabled {
				queryStatsCtx, queryStatsCancel := context.WithTimeout(ctx, 5*time.Second)
				snapshot, err := queryStatsCollector.GetQueryStats(queryStatsCtx)
				queryStatsCancel()

				if err != nil {
					fmt.Printf("failed to collect query stats: %v\n", err)
				} else {
					snapshot.DatabaseInstanceID = stats.DatabaseName
					if r.logLevel == logLevelDebug {
						printQueryPerformanceSummary(snapshot)
					}
					queryStatsSnapshot = snapshot
				}
			}

			collectedAt := time.Now()

			dimensions := map[string]string{
				"agent_id":             r.cfg.AgentID,
				"database_instance_id": r.cfg.DatabaseInstanceID,
				"environment":          r.cfg.AgentEnvironment,
				"database_name":        stats.DatabaseName,
			}

			points := metrics.DatabaseStatsToMetricPoints(stats, collectedAt, dimensions)

			if r.logLevel == logLevelDebug {
				printMetricPoints(points)
			}

			sendMetricsCtx, sendMetricsCancel := context.WithTimeout(ctx, 5*time.Second)
			_, err = apiClient.SendMetrics(sendMetricsCtx, r.cfg.AgentID, r.cfg.DatabaseInstanceID, points)
			sendMetricsCancel()

			metricsSent := 0
			if err != nil {
				fmt.Printf("failed to send metrics to API: %v\n", err)
			} else {
				metricsSent = len(points)
			}

			var delta metrics.DatabaseMetricDelta

			if previousStats == nil {
				delta = metrics.DatabaseMetricDelta{
					CollectedAt:        collectedAt,
					DatabaseInstanceID: stats.DatabaseName,
				}
			} else {
				delta = metrics.CalculateDatabaseDelta(previousStats, stats, collectedAt, stats.DatabaseName)
				if r.logLevel == logLevelDebug {
					printDatabaseDelta(delta)
				}
			}

			findings := engine.AnalyzeDatabaseMetrics(*stats, delta)

			if activitySnapshot != nil {
				findings = append(findings, engine.AnalyzeDatabaseActivity(*activitySnapshot)...)
			}

			if tableStatsSnapshot != nil {
				findings = append(findings, engine.AnalyzeTableStats(*tableStatsSnapshot)...)
			}

			if queryStatsSnapshot != nil {
				findings = append(findings, engine.AnalyzeQueryStats(*queryStatsSnapshot)...)
			}

			for i := range findings {
				findings[i].AgentID = r.cfg.AgentID
				findings[i].DatabaseInstanceID = r.cfg.DatabaseInstanceID
			}

			sendFindingsCtx, sendFindingsCancel := context.WithTimeout(ctx, 5*time.Second)
			_, err = apiClient.SendFindings(sendFindingsCtx, r.cfg.AgentID, r.cfg.DatabaseInstanceID, findings)
			sendFindingsCancel()

			findingsSent := 0
			if err != nil {
				fmt.Printf("failed to send findings to API: %v\n", err)
			} else {
				findingsSent = len(findings)
			}

			activitySessions := 0
			if activitySnapshot != nil {
				activitySessions = len(activitySnapshot.Activities)
			}

			tablesAnalyzed := 0
			if tableStatsSnapshot != nil {
				tablesAnalyzed = len(tableStatsSnapshot.Tables)
			}

			queriesAnalyzed := 0
			if queryStatsSnapshot != nil {
				queriesAnalyzed = len(queryStatsSnapshot.Queries)
			}

			printCollectionSummary(collectionSummary{
				DatabaseName:        stats.DatabaseName,
				MetricsSent:         metricsSent,
				FindingsSent:        findingsSent,
				ActivitySessions:    activitySessions,
				TablesAnalyzed:      tablesAnalyzed,
				QueriesAnalyzed:     queriesAnalyzed,
				QueryAnalysisActive: pgStatStatementsEnabled,
			})

			if len(findings) > 0 {
				printFindingsSummary(findings)

				if r.logLevel == logLevelDebug {
					printFindingsDetail(findings)
				}
			}

			previousStats = stats
		}
	}
}

// registerWithAPI discovers the monitored database's name and host, then
// registers the agent and database instance with the Postgresome API.
func (r *Runner) registerWithAPI(ctx context.Context, pool *pgxpool.Pool, apiClient *client.AgentAPIClient, postgresVersion string) error {
	var databaseName string

	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		return fmt.Errorf("failed to discover database name: %w", err)
	}

	host := databaseHost(r.cfg.DatabaseURL)

	_, err := apiClient.RegisterAgent(ctx, client.RegisterAgentRequest{
		Agent: client.AgentInfo{
			ID:          r.cfg.AgentID,
			Name:        r.cfg.AgentName,
			Environment: r.cfg.AgentEnvironment,
		},
		Database: client.DatabaseInfo{
			ID:      r.cfg.DatabaseInstanceID,
			Name:    databaseName,
			Host:    host,
			Version: postgresVersion,
		},
	})
	if err != nil {
		return fmt.Errorf("agent registration failed: %w", err)
	}

	fmt.Println("Agent registered successfully")

	return nil
}

// databaseHost extracts the host from a Postgres connection string.
func databaseHost(databaseURL string) string {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return ""
	}

	return config.ConnConfig.Host
}

func printDatabaseStats(stats *metrics.DatabaseStats) {
	fmt.Printf("Database: %s\n", stats.DatabaseName)
	fmt.Printf("Connections: %d\n", stats.ActiveConnections)
	fmt.Printf("Commits: %d\n", stats.TransactionCommits)
	fmt.Printf("Rollbacks: %d\n", stats.TransactionRollbacks)
	fmt.Printf("Blocks Read: %d\n", stats.BlocksReadFromDisk)
	fmt.Printf("Blocks Hit: %d\n", stats.BlocksHitInCache)
	fmt.Printf("Rows Returned: %d\n", stats.RowsScanned)
	fmt.Printf("Rows Fetched: %d\n", stats.RowsFetched)
	fmt.Printf("Rows Inserted: %d\n", stats.RowsInserted)
	fmt.Printf("Rows Updated: %d\n", stats.RowsUpdated)
	fmt.Printf("Rows Deleted: %d\n", stats.RowsDeleted)
	fmt.Println("--------------------------------")
}

func printDatabaseDelta(delta metrics.DatabaseMetricDelta) {
	fmt.Println("Transactions since last collection:")
	fmt.Printf("- Commits Delta: %d\n", delta.TransactionCommitsDelta)
	fmt.Printf("- Rollbacks Delta: %d\n", delta.TransactionRollbacksDelta)
	fmt.Printf("- Rollback Rate: %.1f%%\n", delta.RollbackRate*100)
	fmt.Println()

	fmt.Println("Cache since last collection:")
	fmt.Printf("- Blocks Read From Disk Delta: %d\n", delta.BlocksReadFromDiskDelta)
	fmt.Printf("- Blocks Hit In Cache Delta: %d\n", delta.BlocksHitInCacheDelta)
	fmt.Printf("- Cache Hit Ratio: %.1f%%\n", delta.CacheHitRatio*100)
	fmt.Println()

	fmt.Println("Rows since last collection:")
	fmt.Printf("- Rows Scanned Delta: %d\n", delta.RowsScannedDelta)
	fmt.Printf("- Rows Fetched Delta: %d\n", delta.RowsFetchedDelta)
	fmt.Printf("- Rows Inserted Delta: %d\n", delta.RowsInsertedDelta)
	fmt.Printf("- Rows Updated Delta: %d\n", delta.RowsUpdatedDelta)
	fmt.Printf("- Rows Deleted Delta: %d\n", delta.RowsDeletedDelta)
	fmt.Println("--------------------------------")
}

func printDatabaseActivity(snapshot *metrics.DatabaseActivitySnapshot) {
	fmt.Println("Database Activity:")
	fmt.Println()
	fmt.Printf("Total Sessions: %d\n", len(snapshot.Activities))
	fmt.Println()

	stateCounts := make(map[string]int)
	for _, a := range snapshot.Activities {
		state := a.State
		if state == "" {
			state = "background"
		}
		stateCounts[state]++
	}

	states := make([]string, 0, len(stateCounts))
	for state := range stateCounts {
		states = append(states, state)
	}
	sort.Strings(states)

	fmt.Println("States:")
	for _, state := range states {
		fmt.Printf("%s: %d\n", state, stateCounts[state])
	}
	fmt.Println("--------------------------------")

	for i, a := range snapshot.Activities {
		if i >= maxSampleSessions {
			break
		}

		fmt.Println("PID:")
		fmt.Println(a.ProcessID)
		fmt.Println()
		fmt.Println("User:")
		fmt.Println(a.UserName)
		fmt.Println()
		fmt.Println("Application:")
		fmt.Println(a.ApplicationName)
		fmt.Println()
		fmt.Println("State:")
		fmt.Println(a.State)
		fmt.Println()
		fmt.Println("Wait Event Type:")
		fmt.Println(a.WaitEventType)
		fmt.Println()
		fmt.Println("Wait Event:")
		fmt.Println(a.WaitEvent)
		fmt.Println()
		fmt.Println("Query Duration:")
		fmt.Println(formatQueryDuration(a.QueryStartedAt, snapshot.CollectedAt))
		fmt.Println()
		fmt.Println("Query:")
		fmt.Println(previewQuery(a.Query, maxQueryPreviewLength))
		fmt.Println("--------------------------------")
	}
}

func formatQueryDuration(queryStartedAt *time.Time, collectedAt time.Time) string {
	if queryStartedAt == nil {
		return "n/a"
	}

	duration := max(collectedAt.Sub(*queryStartedAt), 0)

	return fmt.Sprintf("%.0f seconds", duration.Seconds())
}

func previewQuery(query string, maxLength int) string {
	query = strings.Join(strings.Fields(query), " ")

	if len(query) <= maxLength {
		return query
	}

	return query[:maxLength] + "..."
}

// collectionSummary holds the counts reported at the end of a collection
// cycle.
type collectionSummary struct {
	DatabaseName        string
	MetricsSent         int
	FindingsSent        int
	ActivitySessions    int
	TablesAnalyzed      int
	QueriesAnalyzed     int
	QueryAnalysisActive bool
}

func printCollectionSummary(s collectionSummary) {
	fmt.Println("Collection completed:")
	fmt.Printf("- Database: %s\n", s.DatabaseName)
	fmt.Printf("- Metrics sent: %d\n", s.MetricsSent)
	fmt.Printf("- Findings sent: %d\n", s.FindingsSent)
	fmt.Printf("- Activity sessions: %d\n", s.ActivitySessions)
	fmt.Printf("- Tables analyzed: %d\n", s.TablesAnalyzed)
	if s.QueryAnalysisActive {
		fmt.Printf("- Queries analyzed: %d\n", s.QueriesAnalyzed)
	}
}

// printFindingsSummary prints a one-line summary per finding: severity,
// category, and title.
func printFindingsSummary(findings []analysis.Finding) {
	for _, finding := range findings {
		fmt.Printf("- [%s] %s: %s\n", finding.Severity, finding.Category, finding.Title)
	}
}

// printFindingsDetail prints the full message and recommendation for each
// finding. Only called in debug mode.
func printFindingsDetail(findings []analysis.Finding) {
	for _, finding := range findings {
		fmt.Println("Finding detected:")
		fmt.Println()
		fmt.Println("Severity:")
		fmt.Println(finding.Severity)
		fmt.Println()
		fmt.Println("Category:")
		fmt.Println(finding.Category)
		fmt.Println()
		fmt.Println("Title:")
		fmt.Println(finding.Title)
		fmt.Println()
		fmt.Println("Message:")
		fmt.Println(finding.Message)
		fmt.Println()
		fmt.Println("Recommendation:")
		fmt.Println(finding.Recommendation)
		fmt.Println("--------------------------------")
	}
}

func printTableHealthSummary(snapshot *metrics.TableStatsSnapshot) {
	fmt.Println("Table Health Summary:")
	fmt.Println()
	fmt.Printf("Total Tables: %d\n", len(snapshot.Tables))
	fmt.Println()

	byDeadRows := make([]metrics.TableStats, len(snapshot.Tables))
	copy(byDeadRows, snapshot.Tables)
	sort.Slice(byDeadRows, func(i, j int) bool {
		return byDeadRows[i].DeadRows > byDeadRows[j].DeadRows
	})

	fmt.Println("Top tables by dead rows:")
	fmt.Println()
	for i, table := range byDeadRows {
		if i >= maxTableSummaryItems {
			break
		}

		fmt.Printf("%d. %s.%s\n", i+1, table.SchemaName, table.TableName)
		fmt.Printf("   Live Rows: %d\n", table.LiveRows)
		fmt.Printf("   Dead Rows: %d\n", table.DeadRows)
		fmt.Println()
	}

	bySequentialScans := make([]metrics.TableStats, len(snapshot.Tables))
	copy(bySequentialScans, snapshot.Tables)
	sort.Slice(bySequentialScans, func(i, j int) bool {
		return bySequentialScans[i].SequentialScans > bySequentialScans[j].SequentialScans
	})

	fmt.Println("Top tables by sequential scans:")
	fmt.Println()
	for i, table := range bySequentialScans {
		if i >= maxTableSummaryItems {
			break
		}

		fmt.Printf("%d. %s.%s\n", i+1, table.SchemaName, table.TableName)
		fmt.Printf("   Sequential Scans: %d\n", table.SequentialScans)
		fmt.Printf("   Sequential Rows Read: %d\n", table.SequentialRowsRead)
		fmt.Println()
	}

	fmt.Println("--------------------------------")
}

func printQueryPerformanceSummary(snapshot *metrics.QueryStatsSnapshot) {
	fmt.Println("Query Performance Summary:")
	fmt.Println()
	fmt.Println("Top queries by total execution time:")
	fmt.Println()

	for i, q := range snapshot.Queries {
		if i >= maxQuerySummaryItems {
			break
		}

		fmt.Printf("%d. Calls: %d\n", i+1, q.Calls)
		fmt.Printf("   Mean Time: %.1f ms\n", q.MeanExecutionTimeMs)
		fmt.Printf("   Total Time: %.0f ms\n", q.TotalExecutionTimeMs)
		fmt.Printf("   Rows Returned: %d\n", q.RowsReturned)
		fmt.Printf("   Query Preview: %s\n", previewQuery(q.Query, maxQueryPreviewLength))
		fmt.Println()
	}

	fmt.Println("--------------------------------")
}

func printMetricPoints(points []metrics.MetricPoint) {
	fmt.Println("MetricPoints:")
	for _, point := range points {
		fmt.Printf("Metric: %s\n", point.Label)
		fmt.Printf("Key: %s\n", point.Key)
		fmt.Printf("Value: %v\n", point.Value)
		fmt.Printf("Unit: %s\n", point.Unit)
		fmt.Printf("Category: %s\n", point.Category)
		fmt.Printf("Dimensions: %v\n", point.Dimensions)
		fmt.Println("--------------------------------")
	}
}
