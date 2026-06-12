package rules

import (
	"fmt"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	diskHeavyQueryWarningReadRatio  = 0.10
	diskHeavyQueryCriticalReadRatio = 0.30
)

// DiskHeavyQueryRule detects queries that read a large portion of their
// blocks from disk rather than the buffer cache, which can indicate missing
// indexes, inefficient scans, or insufficient cache memory.
type DiskHeavyQueryRule struct{}

func (r DiskHeavyQueryRule) Name() string {
	return "disk_heavy_query"
}

func (r DiskHeavyQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		var diskReadRatio float64
		if totalBlocks := query.SharedBlocksRead + query.SharedBlocksHit; totalBlocks > 0 {
			diskReadRatio = float64(query.SharedBlocksRead) / float64(totalBlocks)
		}

		switch {
		case diskReadRatio >= diskHeavyQueryCriticalReadRatio:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              "Critical disk-heavy query detected",
				Message:            fmt.Sprintf("Query %q reads %.1f%% of its blocks from disk.", previewQuery(query.Query, queryPreviewMaxLength), diskReadRatio*100),
				Recommendation:     "This query reads a large portion of blocks from disk. Investigate missing indexes, inefficient scans, and memory/cache behavior.",
			})

		case diskReadRatio >= diskHeavyQueryWarningReadRatio:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "warning",
				Category:           "queries",
				Title:              "Disk-heavy query detected",
				Message:            fmt.Sprintf("Query %q reads %.1f%% of its blocks from disk.", previewQuery(query.Query, queryPreviewMaxLength), diskReadRatio*100),
				Recommendation:     "This query reads a large portion of blocks from disk. Investigate missing indexes, inefficient scans, and memory/cache behavior.",
			})
		}
	}

	return findings
}
