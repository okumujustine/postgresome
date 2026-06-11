package rules

import (
	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

// Rule inspects database metrics and reports zero or more findings.
// Implementations satisfy analysis.Rule and can be registered with an
// analysis.Engine.
//
// Future rules:
//   - High connections rule
//   - Low cache hit ratio rule
//   - High rollback rate rule
//   - High write pressure rule
type Rule interface {
	Name() string
	Analyze(stats metrics.DatabaseStats, delta metrics.DatabaseMetricDelta) []analysis.Finding
}

// ActivityRule inspects a snapshot of pg_stat_activity sessions and reports
// zero or more findings. Implementations satisfy analysis.ActivityRule and
// can be registered with an analysis.Engine via RegisterActivityRules.
type ActivityRule interface {
	Name() string
	Analyze(snapshot metrics.DatabaseActivitySnapshot) []analysis.Finding
}

// TableRule inspects a snapshot of pg_stat_user_tables statistics and
// reports zero or more findings. Implementations satisfy analysis.TableRule
// and can be registered with an analysis.Engine via RegisterTableRules.
type TableRule interface {
	Name() string
	Analyze(snapshot metrics.TableStatsSnapshot) []analysis.Finding
}
