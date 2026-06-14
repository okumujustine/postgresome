package rules

import (
	"fmt"
	"time"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
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
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "vacuum",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       float64(table.DeadRows),
				ThresholdValue:     criticalDeadRows,
			})

		case float64(table.DeadRows) >= warningDeadRows && (neverVacuumed || staleVacuum):
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "vacuum",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
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
