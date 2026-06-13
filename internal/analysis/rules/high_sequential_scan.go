package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	highSequentialScanWarningRatio          = 0.50
	highSequentialScanWarningRowsThreshold  = 100000
	highSequentialScanCriticalRatio         = 0.80
	highSequentialScanCriticalRowsThreshold = 1000000
)

// HighSequentialScanRule detects tables that are scanned sequentially for a
// large share of reads, which can indicate missing indexes, inefficient
// queries, large table scans, or reporting/analytics workloads.
type HighSequentialScanRule struct{}

func (r HighSequentialScanRule) Name() string {
	return "high_sequential_scan"
}

func (r HighSequentialScanRule) Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, table := range snapshot.Tables {
		var sequentialScanRatio float64
		if totalScans := table.SequentialScans + table.IndexScans; totalScans > 0 {
			sequentialScanRatio = float64(table.SequentialScans) / float64(totalScans)
		}

		resourceName := fmt.Sprintf("%s.%s", table.SchemaName, table.TableName)

		switch {
		case sequentialScanRatio >= highSequentialScanCriticalRatio && table.SequentialRowsRead >= highSequentialScanCriticalRowsThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              "Critical sequential scan pressure detected",
				Message:            fmt.Sprintf("Table %s.%s has heavy sequential scan activity and is reading many rows.", table.SchemaName, table.TableName),
				Recommendation:     "Investigate slow queries and missing indexes for this table.",
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       sequentialScanRatio,
				ThresholdValue:     highSequentialScanCriticalRatio,
			})

		case sequentialScanRatio >= highSequentialScanWarningRatio && table.SequentialRowsRead >= highSequentialScanWarningRowsThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "warning",
				Category:           "queries",
				Title:              "High sequential scan activity detected",
				Message:            fmt.Sprintf("Table %s.%s is using sequential scans for a large share of reads.", table.SchemaName, table.TableName),
				Recommendation:     "Review query patterns and consider whether frequently filtered columns need indexes.",
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       sequentialScanRatio,
				ThresholdValue:     highSequentialScanWarningRatio,
			})
		}
	}

	return findings
}
