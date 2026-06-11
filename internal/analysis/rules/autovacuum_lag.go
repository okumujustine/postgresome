package rules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const (
	autovacuumLagWarningDeadRowsThreshold  = 10000
	autovacuumLagCriticalDeadRowsThreshold = 100000
	autovacuumLagCriticalDeadRowRatio      = 0.30
	autovacuumLagStaleThreshold            = 24 * time.Hour
)

// AutovacuumLagRule detects tables where dead rows from UPDATE and DELETE
// activity may be accumulating faster than autovacuum is cleaning them up,
// which can lead to table bloat and degraded query performance.
type AutovacuumLagRule struct{}

func (r AutovacuumLagRule) Name() string {
	return "autovacuum_lag"
}

func (r AutovacuumLagRule) Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding {
	findings := make([]analysis.Finding, 0)

	for _, table := range snapshot.Tables {
		var deadRowRatio float64
		if total := table.LiveRows + table.DeadRows; total > 0 {
			deadRowRatio = float64(table.DeadRows) / float64(total)
		}

		neverVacuumed := table.LastAutoVacuumAt == nil
		staleVacuum := table.LastAutoVacuumAt != nil && snapshot.CollectedAt.Sub(*table.LastAutoVacuumAt) > autovacuumLagStaleThreshold

		switch {
		case (table.DeadRows >= autovacuumLagCriticalDeadRowsThreshold && neverVacuumed) ||
			(deadRowRatio >= autovacuumLagCriticalDeadRowRatio && staleVacuum):
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "critical",
				Category:           "vacuum",
				Title:              "Autovacuum is not keeping up",
				Message:            fmt.Sprintf("Table %s.%s has significant dead row accumulation and no recent autovacuum activity.", table.SchemaName, table.TableName),
				Recommendation:     "Investigate autovacuum configuration and consider manual VACUUM during a safe maintenance window.",
			})

		case table.DeadRows >= autovacuumLagWarningDeadRowsThreshold && (neverVacuumed || staleVacuum):
			findings = append(findings, analysis.Finding{
				DetectedAt:         snapshot.CollectedAt,
				DatabaseInstanceID: snapshot.DatabaseInstanceID,
				Severity:           "warning",
				Category:           "vacuum",
				Title:              "Autovacuum may be lagging",
				Message:            fmt.Sprintf("Table %s.%s has %s dead rows and has not been autovacuumed recently.", table.SchemaName, table.TableName, formatThousands(table.DeadRows)),
				Recommendation:     "Review autovacuum settings, long-running transactions, and table update/delete workload.",
			})
		}
	}

	return findings
}

// formatThousands renders n with comma separators between thousands groups,
// e.g. 25000 -> "25,000".
func formatThousands(n int64) string {
	s := strconv.FormatInt(n, 10)

	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}

	if negative {
		s = "-" + s
	}

	return s
}
