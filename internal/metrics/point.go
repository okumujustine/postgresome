package metrics

import "time"

// MetricPoint represents one collected metric value. These points are raw
// evidence for diagnosis and can later be queried for historical analysis,
// anomaly detection, and supporting charts when needed.
type MetricPoint struct {
	Key   string
	Label string
	Value float64
	Unit  string

	Category string

	CollectedAt time.Time

	Dimensions map[string]string
}
