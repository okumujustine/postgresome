package rules

import (
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

// MetricRule inspects database stats and deltas and reports zero or more
// findings. Implementations satisfy analysis.MetricRule and can be
// registered with an analysis.Engine via RegisterMetricRules.
//
// Future rules:
//   - High connections rule
//   - Low cache hit ratio rule
//   - High rollback rate rule
//   - High write pressure rule
type MetricRule interface {
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

// QueryRule inspects a snapshot of pg_stat_statements statistics and reports
// zero or more findings. Implementations satisfy analysis.QueryRule and can
// be registered with an analysis.Engine via RegisterQueryRules.
type QueryRule interface {
	Name() string
	Analyze(snapshot metrics.QueryStatsSnapshot) []analysis.Finding
}

// ExplainRule inspects a snapshot of EXPLAIN plans for the top queries and
// reports zero or more findings. Implementations satisfy
// analysis.ExplainRule and can be registered with an analysis.Engine via
// RegisterExplainRules.
type ExplainRule interface {
	Name() string
	Analyze(snapshot metrics.ExplainSnapshot) []analysis.Finding
}
