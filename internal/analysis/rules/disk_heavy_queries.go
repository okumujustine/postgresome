package rules

import (
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/analysis/config"
	"github.com/okumujustine/postgresome/internal/metrics"
)

// DiskHeavyQueryRule detects queries that read a large portion of their
// blocks from disk rather than the buffer cache, which can indicate missing
// indexes, inefficient scans, or insufficient cache memory.
type DiskHeavyQueryRule struct {
	Config config.RuleConfig
}

func (r DiskHeavyQueryRule) Name() string {
	return config.RuleKeyDiskHeavyQuery
}

func (r DiskHeavyQueryRule) Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding {
	if !r.Config.Enabled {
		return nil
	}

	warningThreshold := r.Config.Thresholds["warning_ratio"]
	criticalThreshold := r.Config.Thresholds["critical_ratio"]

	findings := make([]analysis.Finding, 0)

	for _, query := range snapshot.Queries {
		var diskReadRatio float64
		if totalBlocks := query.SharedBlocksRead + query.SharedBlocksHit; totalBlocks > 0 {
			diskReadRatio = float64(query.SharedBlocksRead) / float64(totalBlocks)
		}

		switch {
		case diskReadRatio >= criticalThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       diskReadRatio,
				ThresholdValue:     criticalThreshold,
			})

		case diskReadRatio >= warningThreshold:
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           r.Config.Severity,
				Category:           "queries",
				Title:              r.Config.Title,
				Message:            r.Config.Description,
				Recommendation:     r.Config.Recommendation,
				RuleKey:            r.Name(),
				ResourceType:       "query",
				ResourceName:       query.QueryID,
				CurrentValue:       diskReadRatio,
				ThresholdValue:     warningThreshold,
			})
		}
	}

	return findings
}
