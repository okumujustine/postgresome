package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
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

const getMetricRangeStatsSQL = `
	SELECT
		COALESCE(avg(value), 0),
		COALESCE((array_agg(value ORDER BY time ASC))[1], 0),
		COALESCE((array_agg(value ORDER BY time DESC))[1], 0),
		count(*)
	FROM metric_points
	WHERE metric_key = $1
	  AND time >= $2
	  AND time < $3
	  AND ($4 = '' OR database_instance_id = $4)
	  AND ($5 = '' OR agent_id = $5)
`

// MetricRangeStats summarizes a metric over a time range: the average value
// (useful for gauges like active connections) and the first/last values
// (useful for deriving deltas from cumulative counters).
type MetricRangeStats struct {
	Average float64
	First   float64
	Last    float64
	HasData bool
}

// GetMetricRangeStats computes summary statistics for a metric over
// [start, end), optionally filtered by database instance and/or legacy source agent id.
func GetMetricRangeStats(ctx context.Context, pool *pgxpool.Pool, metricKey, databaseInstanceID, agentID string, start, end time.Time) (MetricRangeStats, error) {
	var (
		stats      MetricRangeStats
		pointCount int
	)

	err := pool.QueryRow(ctx, getMetricRangeStatsSQL, metricKey, start, end, databaseInstanceID, agentID).Scan(
		&stats.Average, &stats.First, &stats.Last, &pointCount,
	)
	if err != nil {
		return MetricRangeStats{}, fmt.Errorf("failed to get metric range stats for %q: %w", metricKey, err)
	}

	stats.HasData = pointCount > 0

	return stats, nil
}

type MetricHistoricalContext struct {
	Current         float64
	Previous        float64
	Baseline        float64
	ChangePercent   float64
	HasCurrent      bool
	HasPrevious     bool
	HasBaseline     bool
	FirstAbnormalAt *time.Time
}

const findFirstMetricThresholdCrossingSQL = `
	SELECT min(time)
	FROM metric_points
	WHERE metric_key = $1
	  AND time >= $2
	  AND ($3 = '' OR database_instance_id = $3)
	  AND ($4 = '' OR agent_id = $4)
	  AND value >= $5
`

// GetMetricHistoricalContext summarizes a metric across current, previous,
// and baseline windows so diagnosis code can describe changes relative to
// history instead of only absolute values.
func GetMetricHistoricalContext(ctx context.Context, pool *pgxpool.Pool, metricKey, databaseInstanceID, agentID string, currentStart, currentEnd, previousStart, previousEnd, baselineStart, baselineEnd time.Time, abnormalThreshold float64) (MetricHistoricalContext, error) {
	current, err := GetMetricRangeStats(ctx, pool, metricKey, databaseInstanceID, agentID, currentStart, currentEnd)
	if err != nil {
		return MetricHistoricalContext{}, err
	}
	previous, err := GetMetricRangeStats(ctx, pool, metricKey, databaseInstanceID, agentID, previousStart, previousEnd)
	if err != nil {
		return MetricHistoricalContext{}, err
	}
	baseline, err := GetMetricRangeStats(ctx, pool, metricKey, databaseInstanceID, agentID, baselineStart, baselineEnd)
	if err != nil {
		return MetricHistoricalContext{}, err
	}

	context := MetricHistoricalContext{
		Current:     current.Average,
		Previous:    previous.Average,
		Baseline:    baseline.Average,
		HasCurrent:  current.HasData,
		HasPrevious: previous.HasData,
		HasBaseline: baseline.HasData,
	}

	switch {
	case context.HasBaseline && baseline.Average != 0:
		context.ChangePercent = ((context.Current - context.Baseline) / baseline.Average) * 100
	case context.HasPrevious && previous.Average != 0:
		context.ChangePercent = ((context.Current - context.Previous) / previous.Average) * 100
	}

	if abnormalThreshold > 0 {
		var firstAt *time.Time
		if err := pool.QueryRow(ctx, findFirstMetricThresholdCrossingSQL, metricKey, baselineStart, databaseInstanceID, agentID, abnormalThreshold).Scan(&firstAt); err != nil {
			return MetricHistoricalContext{}, fmt.Errorf("failed to query first threshold crossing for %q: %w", metricKey, err)
		}
		context.FirstAbnormalAt = firstAt
	}

	return context, nil
}
