package diagnosis

import (
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/analysis/rules"
)

func NewDefaultEngine() *analysis.Engine {
	analysisConfig := config.DefaultAnalysisConfig()

	engine := analysis.NewEngine()
	engine.RegisterMetricRules(
		rules.HighConnectionRule{Config: analysisConfig.Rule(config.RuleKeyHighConnections)},
		rules.LowCacheHitRatioRule{Config: analysisConfig.Rule(config.RuleKeyLowCacheHitRatio)},
		rules.HighRollbackRateRule{Config: analysisConfig.Rule(config.RuleKeyHighRollbackRate)},
	)
	engine.RegisterActivityRules(
		rules.IdleConnectionRule{Config: analysisConfig.Rule(config.RuleKeyIdleConnections)},
		rules.LongRunningQueryRule{Config: analysisConfig.Rule(config.RuleKeyLongRunningQuery)},
		rules.BlockedQueryRule{Config: analysisConfig.Rule(config.RuleKeyBlockedQuery)},
	)
	engine.RegisterTableRules(
		rules.HighDeadRowsRule{Config: analysisConfig.Rule(config.RuleKeyHighDeadRows)},
		rules.AutovacuumLagRule{Config: analysisConfig.Rule(config.RuleKeyAutovacuumLag)},
		rules.HighSequentialScanRule{Config: analysisConfig.Rule(config.RuleKeyHighSequentialScan)},
	)
	engine.RegisterQueryRules(
		rules.SlowQueryRule{Config: analysisConfig.Rule(config.RuleKeySlowQuery)},
		rules.ExpensiveQueryRule{Config: analysisConfig.Rule(config.RuleKeyExpensiveQuery)},
		rules.DiskHeavyQueryRule{Config: analysisConfig.Rule(config.RuleKeyDiskHeavyQuery)},
	)
	engine.RegisterExplainRules(rules.MissingIndexRule{}, rules.ExpensiveSortRule{})

	return engine
}
