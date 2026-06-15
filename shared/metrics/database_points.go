package metrics

import "time"

// DatabaseStatsToMetricPoints converts a typed DatabaseStats snapshot into a
// generic list of MetricPoint values, using the Metric Catalog to fill in
// each point's label, unit, and category.
func DatabaseStatsToMetricPoints(stats *DatabaseStats, collectedAt time.Time, dimensions map[string]string) []MetricPoint {
	values := map[string]float64{
		"active_connections":    float64(stats.ActiveConnections),
		"transaction_commits":   float64(stats.TransactionCommits),
		"transaction_rollbacks": float64(stats.TransactionRollbacks),
		"blocks_read_from_disk": float64(stats.BlocksReadFromDisk),
		"blocks_hit_in_cache":   float64(stats.BlocksHitInCache),
		"rows_scanned":          float64(stats.RowsScanned),
		"rows_fetched":          float64(stats.RowsFetched),
		"rows_inserted":         float64(stats.RowsInserted),
		"rows_updated":          float64(stats.RowsUpdated),
		"rows_deleted":          float64(stats.RowsDeleted),
	}

	points := make([]MetricPoint, 0, len(DefaultMetricCatalog))
	for _, definition := range DefaultMetricCatalog {
		value, ok := values[definition.Key]
		if !ok {
			continue
		}

		points = append(points, MetricPoint{
			Key:         definition.Key,
			Label:       definition.Label,
			Value:       value,
			Unit:        definition.Unit,
			Category:    definition.Category,
			CollectedAt: collectedAt,
			Dimensions:  dimensions,
		})
	}

	return points
}
