package api

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/findingcontract"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
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
	SourceID     string `json:"source_id,omitempty"`
	SourceKind   string `json:"source_kind,omitempty"`
	Provider     string `json:"provider,omitempty"`
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
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
	Status         string `json:"status"`

	ProblemSummary      string                   `json:"problem_summary"`
	EvidenceSummary     string                   `json:"evidence_summary"`
	ImpactSummary       string                   `json:"impact_summary"`
	SuggestedAction     string                   `json:"suggested_action"`
	PrimaryImpact       findingcontract.Impact   `json:"primary_impact"`
	SecondaryImpacts    []findingcontract.Impact `json:"secondary_impacts,omitempty"`
	PrimaryAction       findingcontract.Action   `json:"primary_action"`
	SecondaryActions    []findingcontract.Action `json:"secondary_actions,omitempty"`
	ConfidenceLabel     string                   `json:"confidence_label"`
	ConfidenceScore     float64                  `json:"confidence_score"`
	BaselineValue       float64                  `json:"baseline_value"`
	BaselineLabel       string                   `json:"baseline_label"`
	ChangeSummary       string                   `json:"change_summary"`
	VerificationHint    string                   `json:"verification_hint"`
	VerificationStatus  string                   `json:"verification_status"`
	VerificationSummary string                   `json:"verification_summary"`
	RegressionCount     int                      `json:"regression_count"`
	LastRegressedAt     *time.Time               `json:"last_regressed_at,omitempty"`
	ImprovingSince      *time.Time               `json:"improving_since,omitempty"`
	VerifiedFixedAt     *time.Time               `json:"verified_fixed_at,omitempty"`

	RuleKey      string `json:"rule_key"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`

	CurrentValue   float64 `json:"current_value"`
	ThresholdValue float64 `json:"threshold_value"`

	OccurrenceCount int       `json:"occurrence_count"`
	FirstSeenAt     time.Time `json:"first_seen_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	DetectedAt      time.Time `json:"detected_at"`
}

// toDashboardFindingDTO maps a repository finding row to its API
// representation. DetectedAt mirrors LastSeenAt for backward-compatible
// "time since detected" displays.
func toDashboardFindingDTO(f repository.RecentFinding) dashboardFindingDTO {
	primaryImpact, secondaryImpacts, primaryAction, secondaryActions := findingcontract.FromRepository(f)

	return dashboardFindingDTO{
		ID:                  f.ID,
		Severity:            f.Severity,
		Category:            f.Category,
		Title:               f.Title,
		Message:             f.Message,
		Recommendation:      f.Recommendation,
		Status:              f.Status,
		ProblemSummary:      findingProblemSummary(f),
		EvidenceSummary:     findingEvidenceSummary(f),
		ImpactSummary:       f.ImpactSummary,
		SuggestedAction:     findingSuggestedAction(f),
		PrimaryImpact:       primaryImpact,
		SecondaryImpacts:    secondaryImpacts,
		PrimaryAction:       primaryAction,
		SecondaryActions:    secondaryActions,
		ConfidenceLabel:     findingConfidenceLabel(f),
		ConfidenceScore:     findingConfidenceScore(f),
		BaselineValue:       f.BaselineValue,
		BaselineLabel:       f.BaselineLabel,
		ChangeSummary:       f.ChangeSummary,
		VerificationHint:    f.VerificationHint,
		VerificationStatus:  findingVerificationStatus(f),
		VerificationSummary: f.VerificationSummary,
		RegressionCount:     f.RegressionCount,
		LastRegressedAt:     f.LastRegressedAt,
		ImprovingSince:      f.ImprovingSince,
		VerifiedFixedAt:     f.VerifiedFixedAt,
		RuleKey:             f.RuleKey,
		ResourceType:        f.ResourceType,
		ResourceName:        f.ResourceName,
		CurrentValue:        f.CurrentValue,
		ThresholdValue:      f.ThresholdValue,
		OccurrenceCount:     f.OccurrenceCount,
		FirstSeenAt:         f.FirstSeenAt,
		LastSeenAt:          f.LastSeenAt,
		DetectedAt:          f.LastSeenAt,
	}
}

func findingProblemSummary(f repository.RecentFinding) string {
	if f.ProblemSummary != "" {
		return f.ProblemSummary
	}
	return f.Message
}

func findingEvidenceSummary(f repository.RecentFinding) string {
	if f.EvidenceSummary != "" {
		return f.EvidenceSummary
	}

	return fmt.Sprintf(
		"%s crossed the %s threshold with a current value of %.1f against %.1f.",
		f.ResourceName,
		f.RuleKey,
		f.CurrentValue,
		f.ThresholdValue,
	)
}

func findingSuggestedAction(f repository.RecentFinding) string {
	if f.SuggestedAction != "" {
		return f.SuggestedAction
	}
	return f.Recommendation
}

func findingConfidenceLabel(f repository.RecentFinding) string {
	if f.ConfidenceLabel != "" {
		return f.ConfidenceLabel
	}
	return "medium"
}

func findingConfidenceScore(f repository.RecentFinding) float64 {
	if f.ConfidenceScore > 0 {
		return f.ConfidenceScore
	}
	return 0.5
}

func findingVerificationStatus(f repository.RecentFinding) string {
	if f.VerificationStatus != "" {
		return f.VerificationStatus
	}
	return "pending"
}

// handleDashboardOverview serves GET /api/dashboard/overview, an aggregation
// endpoint that summarizes database health, key diagnosis signals, and
// recent findings for the frontend's health landing view. Detailed
// historical evidence can continue to use GET /api/metrics/query.
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

	now := time.Now()
	currentStart := now.Add(-rangeDuration)
	previousStart := now.Add(-2 * rangeDuration)

	ctx := r.Context()

	activeConnections, err := dashboardGaugeMetric(ctx, s.pool, "active_connections", databaseInstanceID, "", currentStart, previousStart, now, "count")
	if err != nil {
		log.Printf("failed to compute active_connections summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	cacheHitRatio, err := dashboardRatioMetric(ctx, s.pool, "blocks_hit_in_cache", "blocks_read_from_disk", databaseInstanceID, "", currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute cache_hit_ratio summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	rollbackRate, err := dashboardRatioMetric(ctx, s.pool, "transaction_rollbacks", "transaction_commits", databaseInstanceID, "", currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute rollback_rate summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	slowQueries, err := dashboardFindingCountMetric(ctx, s.pool, databaseInstanceID, "", currentStart, previousStart, now)
	if err != nil {
		log.Printf("failed to compute slow_queries summary: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	severityCounts, err := repository.CountFindingsBySeverity(ctx, s.pool, databaseInstanceID, "")
	if err != nil {
		log.Printf("failed to count findings by severity: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	recentFindings, err := repository.GetRecentFindings(ctx, s.pool, databaseInstanceID, "", currentStart, dashboardRecentFindingsLimit)
	if err != nil {
		log.Printf("failed to load recent findings: %v", err)
		http.Error(w, "failed to load dashboard overview", http.StatusInternalServerError)
		return
	}

	recentDTOs := make([]dashboardFindingDTO, len(recentFindings))
	for i, f := range recentFindings {
		recentDTOs[i] = toDashboardFindingDTO(f)
	}

	instanceDTO := dashboardDatabaseInstanceDTO{
		ID:           dbInstance.ID,
		DatabaseName: dbInstance.Name,
		Host:         dbInstance.Host,
		SourceID:     dbInstance.SourceID,
		SourceKind:   dbInstance.SourceKind,
		Provider:     dbInstance.Provider,
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
