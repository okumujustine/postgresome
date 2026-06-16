package findingcontract

import (
	"strings"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type Impact struct {
	Code    string `json:"code"`
	Label   string `json:"label"`
	Summary string `json:"summary"`
}

type Action struct {
	Code    string `json:"code"`
	Label   string `json:"label"`
	Summary string `json:"summary"`
}

type Input struct {
	RuleKey         string
	ResourceType    string
	ResourceName    string
	ImpactSummary   string
	SuggestedAction string
	Recommendation  string
}

func FromRepository(f repository.RecentFinding) (Impact, []Impact, Action, []Action) {
	return Build(Input{
		RuleKey:         f.RuleKey,
		ResourceType:    f.ResourceType,
		ResourceName:    f.ResourceName,
		ImpactSummary:   f.ImpactSummary,
		SuggestedAction: f.SuggestedAction,
		Recommendation:  f.Recommendation,
	})
}

func FromAnalysis(f analysis.Finding) (Impact, []Impact, Action, []Action) {
	return Build(Input{
		RuleKey:         f.RuleKey,
		ResourceType:    f.ResourceType,
		ResourceName:    f.ResourceName,
		ImpactSummary:   f.ImpactSummary,
		SuggestedAction: f.SuggestedAction,
		Recommendation:  f.Recommendation,
	})
}

func Build(input Input) (Impact, []Impact, Action, []Action) {
	primaryImpact, secondaryImpacts := impactContract(input)
	primaryAction, secondaryActions := actionContract(input)
	return primaryImpact, secondaryImpacts, primaryAction, secondaryActions
}

func impactContract(input Input) (Impact, []Impact) {
	primarySummary := firstNonEmpty(input.ImpactSummary, genericImpactSummary(input.ResourceType, input.ResourceName))

	switch input.RuleKey {
	case config.RuleKeyHighDeadRows:
		return Impact{
				Code:    "query_latency_risk",
				Label:   "Query slowdown likely",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "storage_growth",
					Label:   "Storage growth continues",
					Summary: "Dead tuples continue to consume heap space until cleanup catches up.",
				},
				{
					Code:    "maintenance_cost_growth",
					Label:   "Maintenance cost rises",
					Summary: "Future vacuum and analyze work becomes more expensive as cleanup debt accumulates.",
				},
			}
	case config.RuleKeyAutovacuumLag:
		return Impact{
				Code:    "maintenance_debt_growth",
				Label:   "Maintenance debt is growing",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "bloat_compounding",
					Label:   "Bloat can compound",
					Summary: "When cleanup falls behind, dead tuples accumulate faster and table health degrades over time.",
				},
			}
	case config.RuleKeySlowQuery:
		return Impact{
				Code:    "request_latency_risk",
				Label:   "Request latency risk",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "disk_pressure_risk",
					Label:   "Disk pressure may rise",
					Summary: "A slower hot-path query can amplify buffer churn and disk reads across the workload.",
				},
			}
	case config.RuleKeyDiskHeavyQuery:
		return Impact{
				Code:    "disk_io_pressure",
				Label:   "Disk I/O pressure likely",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "cache_efficiency_loss",
					Label:   "Cache efficiency drops",
					Summary: "When a query reads heavily from disk, the working set is less likely to stay warm in memory.",
				},
			}
	case config.RuleKeyExpensiveQuery:
		return Impact{
				Code:    "database_time_consumption",
				Label:   "Database time consumption is high",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "concurrency_headroom_loss",
					Label:   "Concurrency headroom shrinks",
					Summary: "A costly query can crowd out healthier work by monopolizing database time.",
				},
			}
	case config.RuleKeyHighSequentialScan:
		return Impact{
				Code:    "inefficient_read_path",
				Label:   "Reads are taking an inefficient path",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "buffer_churn_risk",
					Label:   "Buffer churn may increase",
					Summary: "Sequential scans across large tables can push useful pages out of cache more quickly.",
				},
			}
	case config.RuleKeyBlockedQuery:
		return Impact{
				Code:    "lock_wait_contention",
				Label:   "Lock contention is blocking work",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "request_backlog_risk",
					Label:   "Request backlog may build",
					Summary: "Blocked sessions often cause latency spikes and cascading waits in the application path.",
				},
			}
	case config.RuleKeyLongRunningQuery:
		return Impact{
				Code:    "execution_backlog_risk",
				Label:   "Execution backlog risk",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "connection_hold_time_growth",
					Label:   "Connections stay busy longer",
					Summary: "Long-running queries keep workers and connections occupied longer than expected.",
				},
			}
	case config.RuleKeyHighRollbackRate:
		return Impact{
				Code:    "transaction_failure_risk",
				Label:   "Transactions are failing",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "application_error_risk",
					Label:   "Application errors may be surfacing",
					Summary: "Rollback spikes often reflect failing writes, timeouts, or constraint violations upstream.",
				},
			}
	case config.RuleKeyLowCacheHitRatio:
		return Impact{
				Code:    "cache_efficiency_loss",
				Label:   "Cache efficiency is below normal",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "disk_read_amplification",
					Label:   "Disk reads are likely increasing",
					Summary: "Lower cache hit ratio usually means more blocks are being served from disk instead of memory.",
				},
			}
	case config.RuleKeyHighConnections, config.RuleKeyIdleConnections:
		return Impact{
				Code:    "connection_capacity_pressure",
				Label:   "Connection capacity is under pressure",
				Summary: primarySummary,
			}, []Impact{
				{
					Code:    "pool_exhaustion_risk",
					Label:   "Pool exhaustion risk",
					Summary: "Connection spikes or idle buildup can reduce headroom for healthy requests.",
				},
			}
	default:
		return Impact{
			Code:    "database_health_risk",
			Label:   "Database health risk",
			Summary: primarySummary,
		}, nil
	}
}

func actionContract(input Input) (Action, []Action) {
	primarySummary := firstNonEmpty(input.SuggestedAction, input.Recommendation, genericActionSummary(input.ResourceType, input.ResourceName))

	switch input.RuleKey {
	case config.RuleKeyHighDeadRows:
		return Action{
				Code:    "review_autovacuum_settings",
				Label:   "Review autovacuum settings",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "inspect_table_write_pattern",
					Label:   "Inspect table write pattern",
					Summary: "Check whether high update/delete churn on this table is accelerating dead tuple growth.",
				},
			}
	case config.RuleKeyAutovacuumLag:
		return Action{
				Code:    "tune_autovacuum_capacity",
				Label:   "Tune autovacuum capacity",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "schedule_manual_vacuum",
					Label:   "Schedule manual vacuum",
					Summary: "If the table needs immediate relief, run targeted maintenance during a safe window.",
				},
			}
	case config.RuleKeySlowQuery:
		return Action{
				Code:    "compare_execution_plan",
				Label:   "Compare execution plan",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "review_index_usage",
					Label:   "Review index usage",
					Summary: "Confirm the query is still using the expected index path after the regression started.",
				},
			}
	case config.RuleKeyDiskHeavyQuery:
		return Action{
				Code:    "inspect_query_access_path",
				Label:   "Inspect query access path",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "review_cache_behavior",
					Label:   "Review cache behavior",
					Summary: "Check whether memory pressure or a colder working set explains the jump in disk reads.",
				},
			}
	case config.RuleKeyExpensiveQuery:
		return Action{
				Code:    "optimize_query_or_cache_results",
				Label:   "Optimize query or cache results",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "measure_query_frequency",
					Label:   "Measure query frequency",
					Summary: "Confirm whether the cost comes from one expensive execution pattern or from repeated calls at scale.",
				},
			}
	case config.RuleKeyHighSequentialScan:
		return Action{
				Code:    "review_index_coverage",
				Label:   "Review index coverage",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "inspect_filter_patterns",
					Label:   "Inspect filter patterns",
					Summary: "Look at the predicates most often used against this table before adding or changing indexes.",
				},
			}
	case config.RuleKeyBlockedQuery:
		return Action{
				Code:    "identify_blocking_transaction",
				Label:   "Identify the blocking transaction",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "review_lock_holding_workload",
					Label:   "Review lock-holding workload",
					Summary: "Find the query or transaction holding the lock longest and inspect why it stays open.",
				},
			}
	case config.RuleKeyLongRunningQuery:
		return Action{
				Code:    "inspect_long_running_plan",
				Label:   "Inspect the long-running plan",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "check_lock_waits",
					Label:   "Check lock waits",
					Summary: "Confirm the query is truly slow rather than blocked behind another transaction.",
				},
			}
	case config.RuleKeyHighRollbackRate:
		return Action{
				Code:    "inspect_failed_transactions",
				Label:   "Inspect failed transactions",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "review_constraints_and_timeouts",
					Label:   "Review constraints and timeouts",
					Summary: "Rollback spikes often line up with constraint violations, statement timeouts, or deadlocks.",
				},
			}
	case config.RuleKeyLowCacheHitRatio:
		return Action{
				Code:    "inspect_working_set_and_indexes",
				Label:   "Inspect working set and indexes",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "review_shared_buffers",
					Label:   "Review shared_buffers",
					Summary: "If query patterns look healthy, confirm memory configuration still fits the workload.",
				},
			}
	case config.RuleKeyHighConnections, config.RuleKeyIdleConnections:
		return Action{
				Code:    "review_connection_pooling",
				Label:   "Review connection pooling",
				Summary: primarySummary,
			}, []Action{
				{
					Code:    "check_for_connection_leaks",
					Label:   "Check for connection leaks",
					Summary: "Look for clients that open connections faster than they release them back to the pool.",
				},
			}
	default:
		return Action{
			Code:    "inspect_related_workload",
			Label:   "Inspect the related workload",
			Summary: primarySummary,
		}, nil
	}
}

func genericImpactSummary(resourceType, resourceName string) string {
	switch strings.ToLower(resourceType) {
	case "query":
		return "This query is behaving abnormally enough to affect request latency or database efficiency."
	case "table", "index":
		return "This database object is showing unhealthy behavior that can degrade reads, writes, or maintenance over time."
	default:
		if resourceName != "" {
			return "This database signal moved far enough from normal behavior to justify investigation."
		}
		return "This finding points to a database health risk that deserves investigation."
	}
}

func genericActionSummary(resourceType, resourceName string) string {
	switch strings.ToLower(resourceType) {
	case "query":
		return "Inspect the related query path and compare it with a healthier execution pattern."
	case "table", "index":
		return "Inspect this database object and verify whether maintenance, growth, or access-path behavior changed."
	default:
		if resourceName != "" {
			return "Inspect the related workload and confirm what changed immediately before the finding appeared."
		}
		return "Inspect the related workload and confirm which signal moved first."
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
