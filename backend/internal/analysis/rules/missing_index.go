package rules

import (
	"fmt"
	"strings"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

const (
	seqScanWarningRows  = 1000
	seqScanCriticalRows = 50000
)

// MissingIndexRule detects sequential scans over large estimated row counts,
// which can indicate a missing index on the scanned table.
type MissingIndexRule struct{}

func (r MissingIndexRule) Name() string {
	return "missing_index"
}

func (r MissingIndexRule) Analyze(snapshot metrics.ExplainSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, plan := range snapshot.Plans {
		walkPlan(plan.Root, func(node metrics.PlanNode) {
			if node.NodeType != "Seq Scan" || node.RelationName == "" {
				return
			}

			var severity string
			var thresholdValue float64
			switch {
			case node.PlanRows >= seqScanCriticalRows:
				severity = "critical"
				thresholdValue = seqScanCriticalRows
			case node.PlanRows >= seqScanWarningRows:
				severity = "warning"
				thresholdValue = seqScanWarningRows
			default:
				return
			}

			columns := extractFilterColumns(node.Filter)

			recommendation := fmt.Sprintf("This sequential scan reads an estimated %.0f rows from %q. Review whether an index could avoid the full scan.", node.PlanRows, node.RelationName)
			if len(columns) > 0 {
				recommendation = fmt.Sprintf("Consider adding an index: CREATE INDEX ON %s (%s);", node.RelationName, strings.Join(columns, ", "))
			}

			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           severity,
				Category:           "query_plan",
				Title:              "Sequential scan on large table",
				Message:            fmt.Sprintf("Query %q performs a sequential scan on %q reading an estimated %.0f rows.", previewQuery(plan.Query, queryPreviewMaxLength), node.RelationName, node.PlanRows),
				Recommendation:     recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "query_plan",
				ResourceName:       fmt.Sprintf("%s:%s", plan.QueryID, node.RelationName),
				CurrentValue:       node.PlanRows,
				ThresholdValue:     thresholdValue,
			})
		})
	}

	return findings
}
