package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/collector"
	"github.com/okumujustine/postgresome/internal/metrics"
)

type Runner struct {
	databaseURL string
	interval    time.Duration
}

func NewRunner(databaseURL string, interval time.Duration) *Runner {
	return &Runner{
		databaseURL: databaseURL,
		interval:    interval,
	}
}

func (r *Runner) Start(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, r.databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	postgresCollector := collector.NewPostgresCollector(pool)

	info, err := postgresCollector.GetDatabaseInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Connected to PostgreSQL successfully")
	fmt.Println("Database version:")
	fmt.Println(info.Version)
	fmt.Println()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

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

			printDatabaseStats(stats)

			collectedAt := time.Now()

			dimensions := map[string]string{
				"database_name": stats.DatabaseName,
				"agent_id":      "local-agent",
				"environment":   "development",
			}

			points := metrics.DatabaseStatsToMetricPoints(stats, collectedAt, dimensions)
			printMetricPoints(points)

			if previousStats == nil {
				fmt.Println("Waiting for next collection cycle to calculate metric changes")
			} else {
				delta := metrics.CalculateDatabaseDelta(previousStats, stats, collectedAt, stats.DatabaseName)
				printDatabaseDelta(delta)
			}

			previousStats = stats
		}
	}
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
	fmt.Printf("+%d commits\n", delta.TransactionCommitsDelta)
	fmt.Printf("+%d rollbacks\n", delta.TransactionRollbacksDelta)
	fmt.Println()

	fmt.Println("Rows changed:")
	fmt.Printf("+%d inserted\n", delta.RowsInsertedDelta)
	fmt.Printf("+%d updated\n", delta.RowsUpdatedDelta)
	fmt.Printf("+%d deleted\n", delta.RowsDeletedDelta)
	fmt.Println()

	fmt.Println("Cache:")
	fmt.Printf("%.1f%% hit ratio\n", delta.CacheHitRatio*100)
	fmt.Println()

	fmt.Println("Transaction Health:")
	fmt.Printf("%.1f%% rollback rate\n", delta.RollbackRate*100)
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
