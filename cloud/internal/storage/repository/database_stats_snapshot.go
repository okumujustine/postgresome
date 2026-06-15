package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/shared/metrics"
)

const listLatestMetricSnapshotTimesSQL = `
	SELECT DISTINCT time
	FROM metric_points
	WHERE database_instance_id = $1
	  AND ($2 = '' OR agent_id = $2)
	ORDER BY time DESC
	LIMIT 2
`

const listMetricSnapshotValuesSQL = `
	SELECT metric_key, value
	FROM metric_points
	WHERE database_instance_id = $1
	  AND ($2 = '' OR agent_id = $2)
	  AND time = $3
`

type DatabaseStatsPair struct {
	Current     *metrics.DatabaseStats
	Previous    *metrics.DatabaseStats
	CollectedAt time.Time
}

func GetLatestDatabaseStatsPair(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID, databaseName string) (DatabaseStatsPair, error) {
	rows, err := pool.Query(ctx, listLatestMetricSnapshotTimesSQL, databaseInstanceID, agentID)
	if err != nil {
		return DatabaseStatsPair{}, fmt.Errorf("failed to query latest metric snapshot times: %w", err)
	}
	defer rows.Close()

	times := make([]time.Time, 0, 2)
	for rows.Next() {
		var ts time.Time
		if err := rows.Scan(&ts); err != nil {
			return DatabaseStatsPair{}, fmt.Errorf("failed to scan metric snapshot time: %w", err)
		}
		times = append(times, ts)
	}

	if err := rows.Err(); err != nil {
		return DatabaseStatsPair{}, fmt.Errorf("failed to read metric snapshot times: %w", err)
	}

	if len(times) == 0 {
		return DatabaseStatsPair{}, nil
	}

	current, err := loadDatabaseStatsAt(ctx, pool, databaseInstanceID, agentID, databaseName, times[0])
	if err != nil {
		return DatabaseStatsPair{}, err
	}

	var previous *metrics.DatabaseStats
	if len(times) > 1 {
		previous, err = loadDatabaseStatsAt(ctx, pool, databaseInstanceID, agentID, databaseName, times[1])
		if err != nil {
			return DatabaseStatsPair{}, err
		}
	}

	return DatabaseStatsPair{
		Current:     current,
		Previous:    previous,
		CollectedAt: times[0],
	}, nil
}

func loadDatabaseStatsAt(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID, databaseName string, collectedAt time.Time) (*metrics.DatabaseStats, error) {
	rows, err := pool.Query(ctx, listMetricSnapshotValuesSQL, databaseInstanceID, agentID, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric snapshot values: %w", err)
	}
	defer rows.Close()

	stats := &metrics.DatabaseStats{DatabaseName: databaseName}

	for rows.Next() {
		var (
			key   string
			value float64
		)

		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan metric snapshot value: %w", err)
		}

		switch key {
		case "active_connections":
			stats.ActiveConnections = int(math.Round(value))
		case "transaction_commits":
			stats.TransactionCommits = int64(math.Round(value))
		case "transaction_rollbacks":
			stats.TransactionRollbacks = int64(math.Round(value))
		case "blocks_read_from_disk":
			stats.BlocksReadFromDisk = int64(math.Round(value))
		case "blocks_hit_in_cache":
			stats.BlocksHitInCache = int64(math.Round(value))
		case "rows_scanned":
			stats.RowsScanned = int64(math.Round(value))
		case "rows_fetched":
			stats.RowsFetched = int64(math.Round(value))
		case "rows_inserted":
			stats.RowsInserted = int64(math.Round(value))
		case "rows_updated":
			stats.RowsUpdated = int64(math.Round(value))
		case "rows_deleted":
			stats.RowsDeleted = int64(math.Round(value))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read metric snapshot values: %w", err)
	}

	return stats, nil
}
