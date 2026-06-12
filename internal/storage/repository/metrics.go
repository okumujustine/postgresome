package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const insertMetricPointSQL = `
	INSERT INTO metric_points (time, metric_key, value, unit, database_instance_id, agent_id, dimensions)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`

// InsertMetricPoints persists metric points into the metric_points hypertable.
func InsertMetricPoints(ctx context.Context, pool *pgxpool.Pool, points []metrics.MetricPoint) error {
	for _, point := range points {
		dimensions, err := json.Marshal(point.Dimensions)
		if err != nil {
			return fmt.Errorf("failed to encode dimensions for metric %q: %w", point.Key, err)
		}

		_, err = pool.Exec(ctx, insertMetricPointSQL,
			point.CollectedAt,
			point.Key,
			point.Value,
			point.Unit,
			point.Dimensions["database_instance_id"],
			point.Dimensions["agent_id"],
			dimensions,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric point %q: %w", point.Key, err)
		}
	}

	return nil
}

const queryMetricPointsSQL = `
	SELECT time_bucket($1::interval, time) AS bucket, avg(value) AS avg_value
	FROM metric_points
	WHERE metric_key = $2
	  AND time >= $3
	  AND ($4 = '' OR database_instance_id = $4)
	  AND ($5 = '' OR agent_id = $5)
	GROUP BY bucket
	ORDER BY bucket DESC
`

// MetricQueryParams describes the filters and aggregation window for
// QueryMetricPoints.
type MetricQueryParams struct {
	MetricKey          string
	DatabaseInstanceID string
	AgentID            string
	Since              time.Time
	BucketInterval     string
	Limit              int
}

// MetricQueryPoint is one aggregated time bucket returned by
// QueryMetricPoints.
type MetricQueryPoint struct {
	Time  time.Time
	Value float64
}

// QueryMetricPoints aggregates metric_points into time buckets using
// TimescaleDB's time_bucket, averaging values within each bucket. Results
// are returned oldest to newest.
func QueryMetricPoints(ctx context.Context, pool *pgxpool.Pool, params MetricQueryParams) ([]MetricQueryPoint, error) {
	sql := queryMetricPointsSQL
	args := []any{params.BucketInterval, params.MetricKey, params.Since, params.DatabaseInstanceID, params.AgentID}

	if params.Limit > 0 {
		sql += " LIMIT $6"
		args = append(args, params.Limit)
	}

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric points: %w", err)
	}
	defer rows.Close()

	points := make([]MetricQueryPoint, 0)
	for rows.Next() {
		var point MetricQueryPoint
		if err := rows.Scan(&point.Time, &point.Value); err != nil {
			return nil, fmt.Errorf("failed to scan metric point: %w", err)
		}
		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read metric points: %w", err)
	}

	// The query orders newest to oldest so LIMIT keeps the most recent
	// buckets; reverse to return oldest to newest.
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
