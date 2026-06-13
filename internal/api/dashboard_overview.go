package api

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/storage/repository"
)

const dashboardRecentFindingsLimit = 5

type dashboardOverviewResponse struct {
	DatabaseInstance dashboardDatabaseInstanceDTO `json:"database_instance"`
	Summary          dashboardSummaryDTO          `json:"summary"`
	Findings         dashboardFindingsDTO         `json:"findings"`
}

type dashboardDatabaseInstanceDTO struct {
	ID           string `json:"id"`
	DatabaseName string `json:"database_name"`
	Host         string `json:"host"`
	Status       string `json:"status"`
}

type dashboardSummaryDTO struct {
	ActiveConnections dashboardMetricDTO `json:"active_connections"`
	CacheHitRatio     dashboardMetricDTO `json:"cache_hit_ratio"`
	RollbackRate      dashboardMetricDTO `json:"rollback_rate"`
	SlowQueries       dashboardMetricDTO `json:"slow_queries"`
}

type dashboardMetricDTO struct {
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	TrendPercent float64 `json:"trend_percent"`
}

type dashboardFindingsDTO struct {
	Critical int                   `json:"critical"`
	Warning  int                   `json:"warning"`
	Info     int                   `json:"info"`
	Recent   []dashboardFindingDTO `json:"recent"`
}

type dashboardFindingDTO struct {
	ID             string    `json:"id"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	Recommendation string    `json:"recommendation"`
	DetectedAt     time.Time `json:"detected_at"`
}

// handleDashboardOverview serves GET /api/dashboard/overview, an aggregation
// endpoint that summarizes a database instance's health, key metrics with
// trends, and recent findings for the dashboard landing page. Detailed
// charts should continue to use GET /api/metrics/query.
func (s *Server) handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	rangeParam := query.Get("range")
	if rangeParam == "" {
		rangeParam = defaultMetricQueryRange
	}

	rangeDuration, ok := metricQueryRanges[rangeParam]
	if !ok {
		http.Error(w, "invalid range", http.StatusBadRequest)
		return
	}

	agentID := query.Get("agent_id")
	databaseInstanceID := query.Get("database_instance_id")
	if databaseInstanceID == "" {
		http.Error(w, "database_instance_id is required", http.StatusBadRequest)
		return
	}

	dbInstance, err := repository.GetDatabaseInstance(r.Context(), s.pool, databaseInstanceID)
	if err != nil {
		log.Printf("failed to load database instance: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}
	if dbInstance == nil {
		http.Error(w, "database instance not found", http.StatusNotFound)
		return
	}

	if agentID == "" {
		agentID = dbInstance.AgentID
	}

	now := time.Now()
	currentStart := now.Add(-rangeDuration)
	previousStart := now.Add(-2 * rangeDuration)

	ctx := r.Context()

	activeConnections, err := dashboardGaugeMetric(ctx, s.pool, "active_connections", databaseInstanceID, agentID, currentStart, previousStart, now, "count")
	if err != nil {
		log.Printf("failed to compute active_connections summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	cacheHitRatio, err := dashboardRatioMetric(ctx, s.pool, "blocks_hit_in_cache", "blocks_read_from_disk", databaseInstanceID, agentID, currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute cache_hit_ratio summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	rollbackRate, err := dashboardRatioMetric(ctx, s.pool, "transaction_rollbacks", "transaction_commits", databaseInstanceID, agentID, currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute rollback_rate summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	slowQueries, err := dashboardFindingCountMetric(ctx, s.pool, databaseInstanceID, agentID, currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute slow_queries summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, agentID, currentStart)
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	recentFindings, err := repository.GetRecentFindings(ctx, s.pool, databaseInstanceID, agentID, currentStart, dashboardRecentFindingsLimit)
	if err != nil {
		log.Printf("failed to load recent findings: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	recentDTOs := make([]dashboardFindingDTO, len(recentFindings))
	for i, f := range recentFindings {
		recentDTOs[i] = dashboardFindingDTO{
			ID:             f.ID,
			Severity:       f.Severity,
			Category:       f.Category,
			Title:          f.Title,
			Message:        f.Message,
			Recommendation: f.Recommendation,
			DetectedAt:     f.DetectedAt,
		}
	}

	instanceDTO := dashboardDatabaseInstanceDTO{
		ID:           dbInstance.ID,
		DatabaseName: dbInstance.Name,
		Host:         dbInstance.Host,
		Status:       dashboardInstanceStatus(severityCounts),
	}

	writeJSON(w, http.StatusOK, dashboardOverviewResponse{
		DatabaseInstance: instanceDTO,
		Summary: dashboardSummaryDTO{
			ActiveConnections: activeConnections,
			CacheHitRatio:     cacheHitRatio,
			RollbackRate:      rollbackRate,
			SlowQueries:       slowQueries,
		},
		Findings: dashboardFindingsDTO{
			Critical: severityCounts.Critical,
			Warning:  severityCounts.Warning,
			Info:     severityCounts.Info,
			Recent:   recentDTOs,
		},
	})
}

// dashboardInstanceStatus derives a coarse health status from finding
// severity counts: any critical findings make the instance "critical", any
// warnings (with no criticals) make it "warning", otherwise "healthy".
func dashboardInstanceStatus(counts repository.FindingSeverityCounts) string {
	switch {
	case counts.Critical > 0:
		return "critical"
	case counts.Warning > 0:
		return "warning"
	default:
		return "healthy"
	}
}

// dashboardGaugeMetric summarizes a gauge-style metric (e.g. active
// connections) as the average value over the current range, with
// trend_percent comparing it to the average over the equally-sized previous
// range.
func dashboardGaugeMetric(ctx context.Context, pool *pgxpool.Pool, metricKey, databaseInstanceID, agentID string, currentStart, previousStart, now time.Time, unit string) (dashboardMetricDTO, error) {
	current, err := repository.GetMetricRangeStats(ctx, pool, metricKey, databaseInstanceID, agentID, currentStart, now)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	previous, err := repository.GetMetricRangeStats(ctx, pool, metricKey, databaseInstanceID, agentID, previousStart, currentStart)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	return dashboardMetricDTO{
		Value:        round1(current.Average),
		Unit:         unit,
		TrendPercent: round1(percentChange(current.Average, previous.Average)),
	}, nil
}

// dashboardRatioMetric summarizes a ratio derived from two cumulative
// counters (e.g. cache hits vs. disk reads) as numerator / (numerator +
// denominator), expressed as a percentage. The ratio is computed from the
// counters' deltas within each range, and trend_percent compares the current
// range's ratio to the previous range's.
func dashboardRatioMetric(ctx context.Context, pool *pgxpool.Pool, numeratorKey, denominatorKey, databaseInstanceID, agentID string, currentStart, previousStart, now time.Time) (dashboardMetricDTO, error) {
	currentRatio, err := dashboardRangeRatio(ctx, pool, numeratorKey, denominatorKey, databaseInstanceID, agentID, currentStart, now)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	previousRatio, err := dashboardRangeRatio(ctx, pool, numeratorKey, denominatorKey, databaseInstanceID, agentID, previousStart, currentStart)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	return dashboardMetricDTO{
		Value:        round1(currentRatio),
		Unit:         "percent",
		TrendPercent: round1(percentChange(currentRatio, previousRatio)),
	}, nil
}

func dashboardRangeRatio(ctx context.Context, pool *pgxpool.Pool, numeratorKey, denominatorKey, databaseInstanceID, agentID string, start, end time.Time) (float64, error) {
	numerator, err := repository.GetMetricRangeStats(ctx, pool, numeratorKey, databaseInstanceID, agentID, start, end)
	if err != nil {
		return 0, err
	}

	denominator, err := repository.GetMetricRangeStats(ctx, pool, denominatorKey, databaseInstanceID, agentID, start, end)
	if err != nil {
		return 0, err
	}

	numeratorDelta := positiveDelta(numerator)
	denominatorDelta := positiveDelta(denominator)

	return ratioPercent(numeratorDelta, numeratorDelta+denominatorDelta), nil
}

// dashboardFindingCountMetric summarizes "slow queries" as the number of
// slow-query findings detected in the current range, with trend_percent
// comparing it to the count from the previous range. This is the closest
// existing equivalent to a "slow queries" metric, since per-query stats are
// not persisted as metric_points.
func dashboardFindingCountMetric(ctx context.Context, pool *pgxpool.Pool, databaseInstanceID, agentID string, currentStart, previousStart, now time.Time) (dashboardMetricDTO, error) {
	current, err := repository.CountFindingsByCategoryTitle(ctx, pool, "queries", "%slow query%", databaseInstanceID, agentID, currentStart, now)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	previous, err := repository.CountFindingsByCategoryTitle(ctx, pool, "queries", "%slow query%", databaseInstanceID, agentID, previousStart, currentStart)
	if err != nil {
		return dashboardMetricDTO{}, err
	}

	return dashboardMetricDTO{
		Value:        float64(current),
		Unit:         "count",
		TrendPercent: round1(percentChange(float64(current), float64(previous))),
	}, nil
}

// positiveDelta returns the change between the first and last values in a
// range, treating decreases (e.g. from a counter reset) as zero.
func positiveDelta(stats repository.MetricRangeStats) float64 {
	delta := stats.Last - stats.First
	if delta < 0 {
		return 0
	}

	return delta
}

// ratioPercent returns numerator/denominator as a percentage, or 0 if the
// denominator is zero.
func ratioPercent(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}

	return numerator / denominator * 100
}

// percentChange returns the percent change of current relative to previous,
// or 0 if previous is zero.
func percentChange(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}

	return (current - previous) / previous * 100
}

// round1 rounds a value to one decimal place for cleaner dashboard display.
func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
