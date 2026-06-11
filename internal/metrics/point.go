package metrics

import "time"

// MetricPoint represents one metric value in a frontend-friendly format,
// allowing dashboards to dynamically render cards, charts, and tables
// without hardcoded fields for every metric.
type MetricPoint struct {
	Key   string
	Label string
	Value float64
	Unit  string

	Category string

	CollectedAt time.Time

	Dimensions map[string]string
}
