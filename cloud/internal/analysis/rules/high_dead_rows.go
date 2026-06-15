package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
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
			ProblemSummary:     fmt.Sprintf("Table %s.%s has accumulated %.0f dead tuples, which puts it at risk of table bloat.", table.SchemaName, table.TableName, deadRows),
			EvidenceSummary:    fmt.Sprintf("Dead tuples for %s.%s reached %.0f, above the configured %.0f threshold. The table has seen %d updates and %d deletes since stats collection began.", table.SchemaName, table.TableName, deadRows, threshold, table.RowsUpdated, table.RowsDeleted),
			ImpactSummary:      "As dead tuples accumulate, scans can touch more heap pages and queries against this table may become slower and less predictable.",
			SuggestedAction:    "Review whether autovacuum is keeping up for this table, and schedule a manual vacuum if cleanup is clearly lagging behind churn.",
			ConfidenceLabel:    "medium",
			ConfidenceScore:    0.74,
			ChangeSummary:      "The latest table stats show dead tuples above the bloat-risk threshold.",
			VerificationHint:   fmt.Sprintf("After cleanup, confirm dead tuples for %s.%s fall back below %.0f.", table.SchemaName, table.TableName, threshold),
			RuleKey:            r.Name(),
			ResourceType:       "table",
			ResourceName:       fmt.Sprintf("%s.%s", table.SchemaName, table.TableName),
			CurrentValue:       deadRows,
			ThresholdValue:     threshold,
		})
	}

	return findings
}
