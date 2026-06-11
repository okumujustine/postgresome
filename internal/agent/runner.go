package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/collector"
	"github.com/okumujustine/postgresome/internal/model"
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
		}
	}
}

func printDatabaseStats(stats *model.DatabaseStats) {
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
