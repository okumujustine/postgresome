package rules

import (
	"fmt"
	"strings"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	sortCostWarningThreshold  = 1000
	sortCostCriticalThreshold = 10000
)

// ExpensiveSortRule detects Sort plan nodes with a high planner cost, which
// can indicate a missing index on the sort key or insufficient work_mem.
type ExpensiveSortRule struct{}

func (r ExpensiveSortRule) Name() string {
	return "expensive_sort"
}

func (r ExpensiveSortRule) Analyze(snapshot metrics.ExplainSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, plan := range snapshot.Plans {
		walkPlan(plan.Root, func(node metrics.PlanNode) {
			if node.NodeType != "Sort" {
				return
			}

			var severity string
			switch {
			case node.TotalCost >= sortCostCriticalThreshold:
				severity = "critical"
			case node.TotalCost >= sortCostWarningThreshold:
				severity = "warning"
			default:
				return
			}

			sortKey := strings.Join(node.SortKey, ", ")

			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           severity,
				Category:           "query_plan",
				Title:              "Expensive sort operation",
				Message:            fmt.Sprintf("Query %q sorts an estimated %.0f rows by (%s) at planner cost %.0f.", previewQuery(plan.Query, queryPreviewMaxLength), node.PlanRows, sortKey, node.TotalCost),
				Recommendation:     fmt.Sprintf("Consider an index on (%s) to avoid sorting at query time, or review work_mem if this sort spills to disk.", sortKey),
			})
		})
	}

	return findings
}
