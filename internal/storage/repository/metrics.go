package repository

import (
	"context"
	"encoding/json"
	"fmt"

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
