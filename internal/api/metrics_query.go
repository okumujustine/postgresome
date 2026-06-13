package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/okumujustine/postgresome/internal/storage/repository"
)

const (
	defaultMetricQueryRange    = "1h"
	defaultMetricQueryInterval = "1m"
)

var metricQueryRanges = map[string]time.Duration{
	"15m": 15 * time.Minute,
	"1h":  time.Hour,
	"6h":  6 * time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
}

var metricQueryIntervals = map[string]string{
	"1m":  "1 minute",
	"5m":  "5 minutes",
	"15m": "15 minutes",
	"1h":  "1 hour",
}

type metricQueryResponse struct {
	MetricKey string                `json:"metric_key"`
	Range     string                `json:"range"`
	Interval  string                `json:"interval"`
	Points    []metricQueryPointDTO `json:"points"`
}

type metricQueryPointDTO struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// handleQueryMetrics serves GET /api/metrics/query, returning a time series
// of metric_points aggregated into buckets via TimescaleDB's time_bucket.
func (s *Server) handleQueryMetrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	metricKey := query.Get("metric_key")
	if metricKey == "" {
		http.Error(w, "metric_key is required", http.StatusBadRequest)
		return
	}

	rangeParam := query.Get("range")
	if rangeParam == "" {
		rangeParam = defaultMetricQueryRange
	}

	rangeDuration, ok := metricQueryRanges[rangeParam]
	if !ok {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	intervalParam := query.Get("interval")
	if intervalParam == "" {
		intervalParam = defaultMetricQueryInterval
	}

	bucketInterval, ok := metricQueryIntervals[intervalParam]
	if !ok {
		http.Error(w, "invalid interval", http.StatusBadRequest)
		return
	}

	limit := 0
	if limitParam := query.Get("limit"); limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	databaseInstanceID := query.Get("database_instance_id")
	if databaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	points, err := repository.QueryMetricPoints(r.Context(), s.pool, repository.MetricQueryParams{
		MetricKey:          metricKey,
		DatabaseInstanceID: databaseInstanceID,
		AgentID:            query.Get("agent_id"),
		Since:              time.Now().Add(-rangeDuration),
		BucketInterval:     bucketInterval,
		Limit:              limit,
	})
	if err != nil {
		log.Printf("failed to query metric points: %v", err)
		http.Error(w, "failed to query metrics", http.StatusInternalServerError)
		return
	}

	responsePoints := make([]metricQueryPointDTO, len(points))
	for i, point := range points {
		responsePoints[i] = metricQueryPointDTO{
			Time:  point.Time,
			Value: point.Value,
		}
	}

	writeJSON(w, http.StatusOK, metricQueryResponse{
		MetricKey: metricKey,
		Range:     rangeParam,
		Interval:  intervalParam,
		Points:    responsePoints,
	})
}
