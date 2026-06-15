package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/cloud/internal/analysis"
	"github.com/okumujustine/postgresome/cloud/internal/analysis/config"
	"github.com/okumujustine/postgresome/shared/metrics"
)

// HighSequentialScanRule detects tables that are scanned sequentially for a
// large share of reads, which can indicate missing indexes, inefficient
// queries, large table scans, or reporting/analytics workloads.
type HighSequentialScanRule struct {
	Config config.RuleConfig
}

func (r HighSequentialScanRule) Name() string {
	return config.RuleKeyHighSequentialScan
}

func (r HighSequentialScanRule) Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningRatio := r.Config.Thresholds["warning_ratio"]
	warningRows := r.Config.Thresholds["warning_rows"]
	criticalRatio := r.Config.Thresholds["critical_ratio"]
	criticalRows := r.Config.Thresholds["critical_rows"]

	findings := make([]analysis.Finding, 0)

	for _, table := range snapshot.Tables {
		var sequentialScanRatio float64
		if totalScans := table.SequentialScans + table.IndexScans; totalScans > 0 {
			sequentialScanRatio = float64(table.SequentialScans) / float64(totalScans)
		}

		resourceName := fmt.Sprintf("%s.%s", table.SchemaName, table.TableName)

		switch {
		case sequentialScanRatio >= criticalRatio && float64(table.SequentialRowsRead) >= criticalRows:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       sequentialScanRatio,
				ThresholdValue:     criticalRatio,
			})

		case sequentialScanRatio >= warningRatio && float64(table.SequentialRowsRead) >= warningRows:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "table",
				ResourceName:       resourceName,
				CurrentValue:       sequentialScanRatio,
				ThresholdValue:     warningRatio,
			})
		}
	}

	return findings
}
