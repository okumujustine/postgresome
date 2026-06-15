package rules

import (
	"fmt"
	"time"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
)

// AutovacuumLagRule detects tables where dead rows from UPDATE and DELETE
// activity may be accumulating faster than autovacuum is cleaning them up,
// which can lead to table bloat and degraded query performance.
type AutovacuumLagRule struct {
	Config config.RuleConfig
}

func (r AutovacuumLagRule) Name() string {
	return config.RuleKeyAutovacuumLag
}

func (r AutovacuumLagRule) Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningDeadRows := r.Config.Thresholds["warning_dead_rows"]
	criticalDeadRows := r.Config.Thresholds["critical_dead_rows"]
	criticalDeadRowRatio := r.Config.Thresholds["critical_dead_row_ratio"]
	staleThreshold := time.Duration(r.Config.Thresholds["stale_hours"] * float64(time.Hour))

	findings := make([]analysis.Finding, 0)

	for _, table := range snapshot.Tables {
		var deadRowRatio float64
		if total := table.LiveRows + table.DeadRows; total > 0 {
			deadRowRatio = float64(table.DeadRows) / float64(total)
		}

		neverVacuumed := table.LastAutoVacuumAt == nil
		staleVacuum := table.LastAutoVacuumAt != nil && snapshot.CollectedAt.Sub(*table.LastAutoVacuumAt) > staleThreshold

		resourceName := fmt.Sprintf("%s.%s", table.SchemaName, table.TableName)

		switch {
		case (float64(table.DeadRows) >= criticalDeadRows && neverVacuumed) ||
			(deadRowRatio >= criticalDeadRowRatio && staleVacuum):
			lastAutovacuum := autovacuumStatus(snapshot.CollectedAt, table.LastAutoVacuumAt)
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "vacuum",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				ProblemSummary:     fmt.Sprintf("Autovacuum appears unable to keep up on %s, and dead tuples are now at a critical level.", resourceName),
				EvidenceSummary:    fmt.Sprintf("%s has %d dead tuples, a dead-row ratio of %.1f%%, and %s.", resourceName, table.DeadRows, deadRowRatio*100, lastAutovacuum),
				ImpactSummary:      "Bloat and stale visibility maps can make reads slower, vacuum work more expensive, and further write amplification more likely.",
				SuggestedAction:    "Review autovacuum thresholds and worker capacity for this table, then plan manual maintenance if urgent cleanup is needed.",
				ConfidenceLabel:    "high",
				ConfidenceScore:    0.91,
				BaselineValue:      deadRowRatio * 100,
				BaselineLabel:      "Dead row ratio (%)",
				ChangeSummary:      "The latest sample shows both tuple accumulation and stale autovacuum activity.",
				VerificationHint:   fmt.Sprintf("After remediation, confirm autovacuum runs again and dead tuples on %s drop below %.0f.", resourceName, warningDeadRows),
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       float64(table.DeadRows),
				ThresholdValue:     criticalDeadRows,
			})

		case float64(table.DeadRows) >= warningDeadRows && (neverVacuumed || staleVacuum):
			lastAutovacuum := autovacuumStatus(snapshot.CollectedAt, table.LastAutoVacuumAt)
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "vacuum",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				ProblemSummary:     fmt.Sprintf("Autovacuum cleanup is lagging on %s, allowing dead tuples to build up.", resourceName),
				EvidenceSummary:    fmt.Sprintf("%s has %d dead tuples and %s.", resourceName, table.DeadRows, lastAutovacuum),
				ImpactSummary:      "If cleanup continues to lag, this table can drift toward bloat and slower scan performance.",
				SuggestedAction:    "Check whether this table needs lower autovacuum thresholds or more frequent maintenance during busy write periods.",
				ConfidenceLabel:    "medium",
				ConfidenceScore:    0.79,
				BaselineValue:      deadRowRatio * 100,
				BaselineLabel:      "Dead row ratio (%)",
				ChangeSummary:      "The latest sample crossed the warning threshold while autovacuum still appears stale.",
				VerificationHint:   fmt.Sprintf("Confirm dead tuples for %s drop back below %.0f after the next vacuum cycle.", resourceName, warningDeadRows),
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       float64(table.DeadRows),
				ThresholdValue:     warningDeadRows,
			})
		}
	}

	return findings
}

func autovacuumStatus(now time.Time, lastAutoVacuumAt *time.Time) string {
	if lastAutoVacuumAt == nil {
		return "no autovacuum run has been recorded"
	}

	age := now.Sub(*lastAutoVacuumAt)
	if age >= 48*time.Hour {
		return fmt.Sprintf("the last autovacuum ran %.1f days ago", age.Hours()/24)
	}

	return fmt.Sprintf("the last autovacuum ran %.1f hours ago", age.Hours())
}
