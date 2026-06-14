package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
)

// HighDeadRowsRule detects tables that have accumulated a large number of
// dead rows, which can indicate autovacuum is not keeping up with UPDATE/
// DELETE activity and may lead to table bloat.
type HighDeadRowsRule struct {
	Config config.RuleConfig
}

func (r HighDeadRowsRule) Name() string {
	return config.RuleKeyHighDeadRows
}

func (r HighDeadRowsRule) Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	threshold := r.Config.Thresholds["dead_rows"]
	findings := make([]analysis.Finding, 0)

	for _, table := range snapshot.Tables {
		deadRows := float64(table.DeadRows)
		if deadRows <= threshold {
			continue
		}

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
			ResourceName:       fmt.Sprintf("%s.%s", table.SchemaName, table.TableName),
			CurrentValue:       deadRows,
			ThresholdValue:     threshold,
		})
	}

	return findings
}
