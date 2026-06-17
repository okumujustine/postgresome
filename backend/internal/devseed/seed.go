package devseed

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/secrets"
	"github.com/okumujustine/postgresome/backend/internal/storage"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

const (
	seedSourceProduction = "seed-src-production"
	seedSourceStaging    = "seed-src-staging"

	seedDatabaseProduction = "seed-db-production"
	seedDatabaseStaging    = "seed-db-staging"

	seedQueryTransactions = "qry_transactions_status_feed"
	seedQuerySessionPrune = "qry_session_cleanup"
	seedQueryOrdersFK     = "qry_orders_fk_migration"
	seedQueryReporting    = "qry_reporting_customer_rollup"
)

var seedSourceIDs = []string{
	seedSourceProduction,
	seedSourceStaging,
}

type querySnapshot struct {
	CollectedAt time.Time
	Rows        []repository.QueryStatRow
}

type tableSnapshot struct {
	CollectedAt time.Time
	Rows        []repository.TableStatRow
}

type lifecycleUpdate struct {
	Status              string
	RegressionCount     int
	LastRegressedAt     *time.Time
	ImprovingSince      *time.Time
	VerifiedFixedAt     *time.Time
	VerificationStatus  string
	VerificationSummary string
	ResolvedAt          *time.Time
}

// Run clears and recreates a deterministic development dataset so the
// frontend can exercise diagnosis flows using backend-served data only.
func Run(ctx context.Context, pool *pgxpool.Pool, protector *secrets.ConnectionProtector) error {
	now := time.Now().UTC().Truncate(time.Minute)

	if err := clearSeedData(ctx, pool); err != nil {
		return err
	}

	if err := seedProductionWorkspace(ctx, pool, protector, now); err != nil {
		return err
	}

	if err := seedStagingWorkspace(ctx, pool, protector, now); err != nil {
		return err
	}

	return nil
}

func clearSeedData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DELETE FROM sources WHERE id = ANY($1)`, seedSourceIDs)
	if err != nil {
		return fmt.Errorf("failed to clear seeded sources: %w", err)
	}

	return nil
}

func seedProductionWorkspace(ctx context.Context, pool *pgxpool.Pool, protector *secrets.ConnectionProtector, now time.Time) error {
	if err := seedSourceBase(
		ctx,
		pool,
		protector,
		seedSourceProduction,
		"postgres",
		"Primary production",
		seedDatabaseProduction,
		"app_production",
		"prod-db.internal",
		"postgresql://app:app@prod-db.internal:5432/app_production",
		now.Add(-26*time.Minute),
		now.Add(-18*time.Minute),
	); err != nil {
		return err
	}

	querySnapshots := productionQuerySnapshots(now)
	for _, snapshot := range querySnapshots {
		if err := repository.ReplaceQueryStats(ctx, pool, seedDatabaseProduction, "", snapshot.CollectedAt, snapshot.Rows); err != nil {
			return fmt.Errorf("failed to seed production query stats: %w", err)
		}
	}

	tableSnapshots := productionTableSnapshots(now)
	for _, snapshot := range tableSnapshots {
		if err := repository.ReplaceTableStats(ctx, pool, seedDatabaseProduction, "", snapshot.CollectedAt, snapshot.Rows); err != nil {
			return fmt.Errorf("failed to seed production table stats: %w", err)
		}
	}

	if err := repository.ReplaceActivityStats(
		ctx,
		pool,
		seedDatabaseProduction,
		"",
		now.Add(-12*time.Minute),
		[]repository.ActivityStatRow{
			{
				DatabaseName:     "app_production",
				ProcessID:        8821,
				UserName:         "migration",
				ApplicationName:  "schema-migrator",
				ClientAddress:    "10.12.0.42",
				State:            "active",
				Query:            "ALTER TABLE public.orders ADD CONSTRAINT orders_customer_fk FOREIGN KEY (customer_id) REFERENCES public.customers(id);",
				WaitEventType:    "Lock",
				WaitEvent:        "relation",
				BackendStartedAt: now.Add(-34 * time.Minute),
			},
			{
				DatabaseName:     "app_production",
				ProcessID:        9034,
				UserName:         "api",
				ApplicationName:  "payments-api",
				ClientAddress:    "10.12.0.99",
				State:            "active",
				Query:            "UPDATE public.orders SET updated_at = now() WHERE id = $1;",
				WaitEventType:    "Lock",
				WaitEvent:        "transactionid",
				BackendStartedAt: now.Add(-19 * time.Minute),
				QueryStartedAt:   timePtr(now.Add(-14 * time.Minute)),
				StateChangedAt:   timePtr(now.Add(-14 * time.Minute)),
			},
		},
	); err != nil {
		return fmt.Errorf("failed to seed production activity stats: %w", err)
	}

	findings := productionFindings(now)
	if err := repository.UpsertFindings(ctx, pool, findings); err != nil {
		return fmt.Errorf("failed to seed production findings: %w", err)
	}

	findingIDs := make(map[string]string, len(findings))
	for _, finding := range findings {
		findingIDs[finding.Fingerprint()] = finding.ID
		if err := repository.ReplaceFindingEvidence(ctx, pool, finding.ID, finding.Evidence); err != nil {
			return fmt.Errorf("failed to seed finding evidence for %s: %w", finding.Title, err)
		}
	}

	if err := applyLifecycleUpdate(
		ctx,
		pool,
		findingIDs[fingerprint(config.RuleKeySlowQuery, "query", seedQueryTransactions)],
		lifecycleUpdate{
			Status:              "open",
			RegressionCount:     2,
			LastRegressedAt:     timePtr(now.Add(-96 * time.Minute)),
			VerificationStatus:  "pending",
			VerificationSummary: "The query remains open until mean latency returns close to baseline and disk reads settle.",
		},
	); err != nil {
		return err
	}

	if err := applyLifecycleUpdate(
		ctx,
		pool,
		findingIDs[fingerprint(config.RuleKeyHighDeadRows, "table", "public.user_sessions")],
		lifecycleUpdate{
			Status:              "open",
			RegressionCount:     1,
			ImprovingSince:      timePtr(now.Add(-44 * time.Minute)),
			VerificationStatus:  "improving",
			VerificationSummary: "Dead row growth slowed after the last autovacuum, but the table is still above the healthy range.",
		},
	); err != nil {
		return err
	}

	resolvedAt := now.Add(-6 * time.Hour)
	verifiedAt := now.Add(-5 * time.Hour)
	if err := applyLifecycleUpdate(
		ctx,
		pool,
		findingIDs[fingerprint(config.RuleKeyLowCacheHitRatio, "database", "app_production")],
		lifecycleUpdate{
			Status:              "resolved",
			VerificationStatus:  "verified_fixed",
			VerificationSummary: "Cache hit ratio recovered after the read-heavy reporting job was moved off peak hours.",
			ResolvedAt:          &resolvedAt,
			VerifiedFixedAt:     &verifiedAt,
		},
	); err != nil {
		return err
	}

	return nil
}

func seedStagingWorkspace(ctx context.Context, pool *pgxpool.Pool, protector *secrets.ConnectionProtector, now time.Time) error {
	if err := seedSourceBase(
		ctx,
		pool,
		protector,
		seedSourceStaging,
		"supabase",
		"Staging mirror",
		seedDatabaseStaging,
		"app_staging",
		"staging-db.internal",
		"postgresql://app:app@staging-db.internal:5432/app_staging",
		now.Add(-2*time.Hour),
		now.Add(-110*time.Minute),
	); err != nil {
		return err
	}

	for _, snapshot := range []querySnapshot{
		{
			CollectedAt: now.Add(-35 * time.Minute),
			Rows: []repository.QueryStatRow{
				{
					QueryID:          "qry_staging_dashboard",
					DatabaseName:     "app_staging",
					UserName:         "web",
					Query:            "SELECT id, status, created_at FROM public.orders WHERE status = $1 ORDER BY created_at DESC LIMIT 50;",
					Calls:            820,
					TotalExecTimeMs:  32210,
					MeanExecTimeMs:   39.3,
					MinExecTimeMs:    8.4,
					MaxExecTimeMs:    86.2,
					RowsReturned:     41000,
					SharedBlocksHit:  18420,
					SharedBlocksRead: 920,
				},
			},
		},
	} {
		if err := repository.ReplaceQueryStats(ctx, pool, seedDatabaseStaging, "", snapshot.CollectedAt, snapshot.Rows); err != nil {
			return fmt.Errorf("failed to seed staging query stats: %w", err)
		}
	}

	for _, snapshot := range []tableSnapshot{
		{
			CollectedAt: now.Add(-35 * time.Minute),
			Rows: []repository.TableStatRow{
				{
					SchemaName:         "public",
					TableName:          "orders",
					LiveRows:           142000,
					DeadRows:           1400,
					SequentialScans:    22,
					SequentialRowsRead: 4020,
					IndexScans:         18820,
					IndexRowsFetched:   92010,
					RowsInserted:       4200,
					RowsUpdated:        6100,
					RowsDeleted:        120,
					LastVacuumAt:       timePtr(now.Add(-3 * time.Hour)),
					LastAutoVacuumAt:   timePtr(now.Add(-70 * time.Minute)),
					LastAnalyzeAt:      timePtr(now.Add(-6 * time.Hour)),
					LastAutoAnalyzeAt:  timePtr(now.Add(-80 * time.Minute)),
					VacuumCount:        2,
					AutoVacuumCount:    14,
					AnalyzeCount:       1,
					AutoAnalyzeCount:   12,
				},
			},
		},
	} {
		if err := repository.ReplaceTableStats(ctx, pool, seedDatabaseStaging, "", snapshot.CollectedAt, snapshot.Rows); err != nil {
			return fmt.Errorf("failed to seed staging table stats: %w", err)
		}
	}

	return nil
}

func seedSourceBase(
	ctx context.Context,
	pool *pgxpool.Pool,
	protector *secrets.ConnectionProtector,
	sourceID string,
	provider string,
	sourceName string,
	databaseID string,
	databaseName string,
	host string,
	connectionURI string,
	startedAt time.Time,
	completedAt time.Time,
) error {
	if err := repository.UpsertSource(ctx, pool, sourceID, "direct", provider, sourceName, ""); err != nil {
		return fmt.Errorf("failed to upsert source %s: %w", sourceID, err)
	}

	encryptedURI, err := protector.Encrypt(connectionURI)
	if err != nil {
		return fmt.Errorf("failed to encrypt connection URI for %s: %w", sourceID, err)
	}

	if err := repository.UpsertSourceConnectionProfile(ctx, pool, sourceID, encryptedURI); err != nil {
		return fmt.Errorf("failed to upsert connection profile for %s: %w", sourceID, err)
	}

	if err := storage.UpsertDatabaseInstance(ctx, pool, databaseID, sourceID, "", databaseName, host); err != nil {
		return fmt.Errorf("failed to upsert database instance %s: %w", databaseID, err)
	}

	if err := repository.MarkSourceCheckRunning(ctx, pool, sourceID, startedAt); err != nil {
		return fmt.Errorf("failed to mark source %s running: %w", sourceID, err)
	}

	if err := repository.MarkSourceCheckFinished(ctx, pool, sourceID, completedAt, "succeeded", ""); err != nil {
		return fmt.Errorf("failed to mark source %s finished: %w", sourceID, err)
	}

	return nil
}

func productionQuerySnapshots(now time.Time) []querySnapshot {
	return []querySnapshot{
		{
			CollectedAt: now.Add(-72 * time.Hour),
			Rows: []repository.QueryStatRow{
				buildQuery(seedQueryTransactions, "api", 156200, 640000, 4.1, 1.2, 34.4, 0, 240000, 9200, transactionsQuerySQL),
				buildQuery(seedQuerySessionPrune, "worker", 10400, 182000, 17.5, 9.2, 42.8, 12000, 18200, 2200, sessionPruneQuerySQL),
				buildQuery(seedQueryReporting, "reporting", 320, 90400, 282.5, 110.4, 540.0, 912000, 12040, 9400, reportingQuerySQL),
			},
		},
		{
			CollectedAt: now.Add(-24 * time.Hour),
			Rows: []repository.QueryStatRow{
				buildQuery(seedQueryTransactions, "api", 171800, 862000, 5.0, 1.4, 56.8, 0, 302000, 14220, transactionsQuerySQL),
				buildQuery(seedQuerySessionPrune, "worker", 12100, 224000, 18.5, 8.8, 48.2, 14000, 21400, 2800, sessionPruneQuerySQL),
				buildQuery(seedQueryReporting, "reporting", 350, 98800, 282.3, 108.6, 552.2, 985000, 13020, 10110, reportingQuerySQL),
			},
		},
		{
			CollectedAt: now.Add(-6 * time.Hour),
			Rows: []repository.QueryStatRow{
				buildQuery(seedQueryTransactions, "api", 178900, 1138000, 6.4, 1.3, 92.0, 0, 404000, 46220, transactionsQuerySQL),
				buildQuery(seedQuerySessionPrune, "worker", 12640, 241200, 19.1, 9.0, 52.3, 15010, 23800, 3400, sessionPruneQuerySQL),
				buildQuery(seedQueryReporting, "reporting", 364, 100400, 275.8, 104.2, 540.8, 992000, 13880, 9640, reportingQuerySQL),
				buildQuery(seedQueryOrdersFK, "migration", 1, 40200, 40200, 40200, 40200, 0, 1800, 220, ordersForeignKeyQuerySQL),
			},
		},
		{
			CollectedAt: now.Add(-12 * time.Minute),
			Rows: []repository.QueryStatRow{
				buildQuery(seedQueryTransactions, "api", 184002, 1834200, 9.97, 1.12, 245.8, 0, 1240020, 182340, transactionsQuerySQL),
				buildQuery(seedQuerySessionPrune, "worker", 12904, 272400, 21.1, 9.4, 58.0, 16620, 26110, 4100, sessionPruneQuerySQL),
				buildQuery(seedQueryReporting, "reporting", 372, 104200, 280.1, 103.8, 562.0, 1004000, 14200, 9820, reportingQuerySQL),
				buildQuery(seedQueryOrdersFK, "migration", 12, 338000, 28166.7, 16004.1, 92120.4, 0, 8400, 1900, ordersForeignKeyQuerySQL),
			},
		},
	}
}

func productionTableSnapshots(now time.Time) []tableSnapshot {
	return []tableSnapshot{
		{
			CollectedAt: now.Add(-72 * time.Hour),
			Rows: []repository.TableStatRow{
				buildUserSessionsTable(now.Add(-72*time.Hour), 812000, 52000, 440, 120200, 4200, 1882000, 482000, 220400, 18040),
				buildOrdersTable(now.Add(-72*time.Hour), 2420000, 3800, 22, 4020, 18120, 92040, 8420, 14400, 1200),
			},
		},
		{
			CollectedAt: now.Add(-24 * time.Hour),
			Rows: []repository.TableStatRow{
				buildUserSessionsTable(now.Add(-24*time.Hour), 842000, 184000, 820, 244200, 4300, 1918000, 524000, 244000, 19400),
				buildOrdersTable(now.Add(-24*time.Hour), 2442000, 4200, 24, 4320, 18220, 92820, 9200, 15600, 1300),
			},
		},
		{
			CollectedAt: now.Add(-6 * time.Hour),
			Rows: []repository.TableStatRow{
				buildUserSessionsTable(now.Add(-6*time.Hour), 861000, 620000, 1380, 482000, 4380, 1930000, 620100, 290200, 21080),
				buildOrdersTable(now.Add(-6*time.Hour), 2461000, 5100, 28, 4680, 18340, 93420, 9800, 16220, 1400),
			},
		},
		{
			CollectedAt: now.Add(-12 * time.Minute),
			Rows: []repository.TableStatRow{
				buildUserSessionsTable(now.Add(-12*time.Minute), 882000, 2500000, 1980, 822000, 4410, 1944200, 702000, 344200, 22840),
				buildOrdersTable(now.Add(-12*time.Minute), 2473000, 5900, 31, 5040, 18410, 94010, 10240, 17080, 1520),
			},
		},
	}
}

func productionFindings(now time.Time) []analysis.Finding {
	latestObservedAt := now.Add(-12 * time.Minute)

	return []analysis.Finding{
		{
			DetectedAt:         latestObservedAt,
			DatabaseInstanceID: seedDatabaseProduction,
			Severity:           "warning",
			Category:           "query_performance",
			Title:              "Slow transaction feed query detected",
			Message:            "A frequently executed transactions query is taking materially longer than its normal execution path.",
			Recommendation:     "Review the filter path and execution plan for the transactions feed query.",
			ProblemSummary:     "The transactions feed query is now averaging close to 10 ms where the historical baseline stayed near 4 ms. The planner is reading far more pages from disk before producing the result set.",
			EvidenceSummary:    "Mean execution time more than doubled while shared block reads climbed sharply and the same query remained hot in the workload.",
			ImpactSummary:      "Customer timeline views and worker polling on recent transactions will feel slower under load.",
			SuggestedAction:    "Open the query detail, compare the latest plan with the healthier baseline, and validate whether an index or selectivity change caused the regression.",
			ConfidenceLabel:    "high",
			ConfidenceScore:    0.93,
			BaselineValue:      4.1,
			BaselineLabel:      "baseline mean latency (ms)",
			ChangeSummary:      "Mean execution time increased from 4.1 ms to 9.97 ms while disk reads for the query expanded significantly.",
			VerificationHint:   "This finding is healthy again once mean latency returns near baseline and disk reads normalize.",
			VerificationStatus: "pending",
			VerificationSummary: "The regression is still active.",
			RuleKey:             config.RuleKeySlowQuery,
			ResourceType:        "query",
			ResourceName:        seedQueryTransactions,
			CurrentValue:        9.97,
			ThresholdValue:      8,
			Evidence: []analysis.FindingEvidence{
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "query_stat",
					Role:             "trigger",
					Label:            "Mean execution time increased",
					Summary:          "The query now averages 9.97 ms instead of the baseline 4.1 ms.",
					MetricKey:        "mean_exec_time_ms",
					ReferenceID:      seedQueryTransactions,
					CurrentValue:     9.97,
					BaselineValue:    4.1,
					ChangePercent:    143.17,
					ConfidenceScore:  0.93,
					Metadata:         map[string]any{"unit": "ms"},
				},
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "query_stat",
					Role:             "supporting",
					Label:            "Disk reads climbed sharply",
					Summary:          "Shared blocks read increased from 9.2k to 182.3k, pointing to a colder read path.",
					MetricKey:        "shared_blocks_read",
					ReferenceID:      seedQueryTransactions,
					CurrentValue:     182340,
					BaselineValue:    9200,
					ChangePercent:    1882.0,
					ConfidenceScore:  0.88,
					Metadata:         map[string]any{"unit": "blocks"},
				},
			},
		},
		{
			DetectedAt:         latestObservedAt,
			DatabaseInstanceID: seedDatabaseProduction,
			Severity:           "warning",
			Category:           "table_bloat",
			Title:              "Table bloat risk detected on user_sessions",
			Message:            "Dead rows have accumulated much faster than autovacuum is reclaiming them.",
			Recommendation:     "Review autovacuum cadence and consider a manual cleanup if growth continues.",
			ProblemSummary:     "The user_sessions table built up roughly 2.5 million dead rows and scan pressure is rising with it, which usually means cleanup is lagging behind write churn.",
			EvidenceSummary:    "Dead rows grew from tens of thousands into the millions while sequential scans increased and the latest autovacuum is no longer keeping up.",
			ImpactSummary:      "Queries touching session cleanup and lookup paths may read more heap pages than necessary and degrade over time.",
			SuggestedAction:    "Inspect autovacuum settings for user_sessions and confirm there is no long-running transaction preventing cleanup.",
			ConfidenceLabel:    "high",
			ConfidenceScore:    0.9,
			BaselineValue:      52000,
			BaselineLabel:      "baseline dead rows",
			ChangeSummary:      "Dead rows increased from about 52k to 2.5M and sequential scans rose at the same time.",
			VerificationHint:   "This finding improves when dead rows flatten out and future scans stop drifting upward.",
			VerificationStatus: "improving",
			VerificationSummary: "Cleanup signals have started improving, but the table is still materially above baseline.",
			RuleKey:             config.RuleKeyHighDeadRows,
			ResourceType:        "table",
			ResourceName:        "public.user_sessions",
			CurrentValue:        2500000,
			ThresholdValue:      100000,
			Evidence: []analysis.FindingEvidence{
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "table_stat",
					Role:             "trigger",
					Label:            "Dead rows crossed the warning threshold",
					Summary:          "Dead rows are now 2.5M versus a healthy baseline near 52k.",
					MetricKey:        "dead_rows",
					ReferenceID:      "public.user_sessions",
					CurrentValue:     2500000,
					BaselineValue:    52000,
					ChangePercent:    4707.69,
					ConfidenceScore:  0.9,
					Metadata:         map[string]any{"unit": "rows"},
				},
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "table_stat",
					Role:             "supporting",
					Label:            "Sequential scans continued rising",
					Summary:          "Sequential scan volume rose from 440 to 1,980 across the observation window.",
					MetricKey:        "sequential_scans",
					ReferenceID:      "public.user_sessions",
					CurrentValue:     1980,
					BaselineValue:    440,
					ChangePercent:    350,
					ConfidenceScore:  0.82,
					Metadata:         map[string]any{"unit": "scans"},
				},
			},
		},
		{
			DetectedAt:         latestObservedAt,
			DatabaseInstanceID: seedDatabaseProduction,
			Severity:           "critical",
			Category:           "lock_contention",
			Title:              "Blocking lock detected around orders migration",
			Message:            "A schema migration is holding a lock that is blocking application work on orders.",
			Recommendation:     "Inspect the blocking session and decide whether the migration should be cancelled or rescheduled.",
			ProblemSummary:     "The foreign key migration on orders is holding a stronger lock than expected, and blocked sessions started queueing behind it within minutes.",
			EvidenceSummary:    "Blocked session count rose from 0 to 12 and lock wait time spiked while the migration statement remained active.",
			ImpactSummary:      "Write-heavy order flows can stall behind the lock and surface request latency or timeouts upstream.",
			SuggestedAction:    "Identify the migration backend first, then review whether a lower-locking rollout path is required.",
			ConfidenceLabel:    "high",
			ConfidenceScore:    0.97,
			BaselineValue:      0,
			BaselineLabel:      "baseline blocked sessions",
			ChangeSummary:      "Blocked sessions rose from 0 to 12 shortly after the migration began.",
			VerificationHint:   "The issue is clear once the blocking session disappears and waiting sessions return to zero.",
			VerificationStatus: "pending",
			VerificationSummary: "The blocking lock is still active.",
			RuleKey:             config.RuleKeyBlockedQuery,
			ResourceType:        "lock",
			ResourceName:        "public.orders",
			CurrentValue:        12,
			ThresholdValue:      5,
			Evidence: []analysis.FindingEvidence{
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "activity_stat",
					Role:             "trigger",
					Label:            "Blocked sessions increased",
					Summary:          "Twelve sessions are currently waiting on the migration lock.",
					MetricKey:        "blocked_sessions",
					ReferenceID:      "pid:8821",
					CurrentValue:     12,
					BaselineValue:    0,
					ChangePercent:    100,
					ConfidenceScore:  0.97,
					Metadata:         map[string]any{"blocking_pid": 8821},
				},
				{
					ObservedAt:       latestObservedAt,
					EvidenceType:     "activity_stat",
					Role:             "supporting",
					Label:            "Migration statement remains active",
					Summary:          "The foreign key migration is still holding a relation lock on orders.",
					MetricKey:        "lock_wait",
					ReferenceID:      "pid:8821",
					CurrentValue:     14,
					BaselineValue:    0,
					ChangePercent:    100,
					ConfidenceScore:  0.9,
					Metadata:         map[string]any{"unit": "minutes"},
				},
			},
		},
		{
			DetectedAt:         now.Add(-7 * time.Hour),
			DatabaseInstanceID: seedDatabaseProduction,
			Severity:           "warning",
			Category:           "cache_efficiency",
			Title:              "Low cache hit ratio detected",
			Message:            "The database served a much larger share of reads from disk than normal.",
			Recommendation:     "Review the read-heavy workload and memory pressure that caused the cache miss spike.",
			ProblemSummary:     "A reporting burst pushed the cache hit ratio down below the healthy baseline for a short period.",
			EvidenceSummary:    "The ratio fell from the usual high-90s into the low-80s before recovering after workload scheduling changed.",
			ImpactSummary:      "The read path became slower during the incident window because more blocks came from disk.",
			SuggestedAction:    "Confirm the workload shift remains outside peak hours so the cache can stay warm for application traffic.",
			ConfidenceLabel:    "medium",
			ConfidenceScore:    0.84,
			BaselineValue:      0.98,
			BaselineLabel:      "baseline cache hit ratio",
			ChangeSummary:      "Cache hit ratio dropped from 0.98 to 0.82 during the reporting window.",
			VerificationHint:   "Healthy once the cache hit ratio remains back above 0.95 through peak traffic.",
			VerificationStatus: "pending",
			VerificationSummary: "The incident has already stabilized.",
			RuleKey:             config.RuleKeyLowCacheHitRatio,
			ResourceType:        "database",
			ResourceName:        "app_production",
			CurrentValue:        0.82,
			ThresholdValue:      0.95,
			Evidence: []analysis.FindingEvidence{
				{
					ObservedAt:       now.Add(-7 * time.Hour),
					EvidenceType:     "metric_point",
					Role:             "trigger",
					Label:            "Cache hit ratio dropped",
					Summary:          "The observed cache hit ratio fell below the warning threshold during the reporting burst.",
					MetricKey:        "cache_hit_ratio",
					ReferenceID:      "app_production",
					CurrentValue:     0.82,
					BaselineValue:    0.98,
					ChangePercent:    -16.33,
					ConfidenceScore:  0.84,
					Metadata:         map[string]any{"unit": "ratio"},
				},
			},
		},
	}
}

func buildQuery(queryID, userName string, calls int64, total float64, mean float64, min float64, max float64, rows int64, hit int64, read int64, sql string) repository.QueryStatRow {
	return repository.QueryStatRow{
		QueryID:          queryID,
		DatabaseName:     "app_production",
		UserName:         userName,
		Query:            sql,
		Calls:            calls,
		TotalExecTimeMs:  total,
		MeanExecTimeMs:   mean,
		MinExecTimeMs:    min,
		MaxExecTimeMs:    max,
		RowsReturned:     rows,
		SharedBlocksHit:  hit,
		SharedBlocksRead: read,
	}
}

func buildUserSessionsTable(collectedAt time.Time, liveRows int64, deadRows int64, sequentialScans int64, sequentialRowsRead int64, indexScans int64, indexRowsFetched int64, rowsInserted int64, rowsUpdated int64, rowsDeleted int64) repository.TableStatRow {
	return repository.TableStatRow{
		SchemaName:         "public",
		TableName:          "user_sessions",
		LiveRows:           liveRows,
		DeadRows:           deadRows,
		SequentialScans:    sequentialScans,
		SequentialRowsRead: sequentialRowsRead,
		IndexScans:         indexScans,
		IndexRowsFetched:   indexRowsFetched,
		RowsInserted:       rowsInserted,
		RowsUpdated:        rowsUpdated,
		RowsDeleted:        rowsDeleted,
		LastVacuumAt:       timePtr(collectedAt.Add(-36 * time.Hour)),
		LastAutoVacuumAt:   timePtr(collectedAt.Add(-30 * time.Hour)),
		LastAnalyzeAt:      timePtr(collectedAt.Add(-16 * time.Hour)),
		LastAutoAnalyzeAt:  timePtr(collectedAt.Add(-12 * time.Hour)),
		VacuumCount:        1,
		AutoVacuumCount:    8,
		AnalyzeCount:       1,
		AutoAnalyzeCount:   6,
	}
}

func buildOrdersTable(collectedAt time.Time, liveRows int64, deadRows int64, sequentialScans int64, sequentialRowsRead int64, indexScans int64, indexRowsFetched int64, rowsInserted int64, rowsUpdated int64, rowsDeleted int64) repository.TableStatRow {
	return repository.TableStatRow{
		SchemaName:         "public",
		TableName:          "orders",
		LiveRows:           liveRows,
		DeadRows:           deadRows,
		SequentialScans:    sequentialScans,
		SequentialRowsRead: sequentialRowsRead,
		IndexScans:         indexScans,
		IndexRowsFetched:   indexRowsFetched,
		RowsInserted:       rowsInserted,
		RowsUpdated:        rowsUpdated,
		RowsDeleted:        rowsDeleted,
		LastVacuumAt:       timePtr(collectedAt.Add(-8 * time.Hour)),
		LastAutoVacuumAt:   timePtr(collectedAt.Add(-90 * time.Minute)),
		LastAnalyzeAt:      timePtr(collectedAt.Add(-5 * time.Hour)),
		LastAutoAnalyzeAt:  timePtr(collectedAt.Add(-2 * time.Hour)),
		VacuumCount:        2,
		AutoVacuumCount:    18,
		AnalyzeCount:       2,
		AutoAnalyzeCount:   15,
	}
}

func applyLifecycleUpdate(ctx context.Context, pool *pgxpool.Pool, findingID string, update lifecycleUpdate) error {
	if findingID == "" {
		return fmt.Errorf("missing seeded finding id for lifecycle update")
	}

	_, err := pool.Exec(ctx, `
		UPDATE findings
		SET
			status = $2,
			regression_count = $3,
			last_regressed_at = $4,
			improving_since = $5,
			verified_fixed_at = $6,
			verification_status = $7,
			verification_summary = $8,
			resolved_at = $9
		WHERE id = $1::uuid
	`, findingID, update.Status, update.RegressionCount, update.LastRegressedAt, update.ImprovingSince, update.VerifiedFixedAt, update.VerificationStatus, update.VerificationSummary, update.ResolvedAt)
	if err != nil {
		return fmt.Errorf("failed to update seeded finding lifecycle for %s: %w", findingID, err)
	}

	return nil
}

func fingerprint(ruleKey, resourceType, resourceName string) string {
	return fmt.Sprintf("%s:%s:%s", ruleKey, resourceType, resourceName)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

const transactionsQuerySQL = `SELECT id, customer_id, status, created_at
FROM public.transactions
WHERE status = $1
ORDER BY created_at DESC
LIMIT 100;`

const sessionPruneQuerySQL = `DELETE FROM public.user_sessions
WHERE expires_at < now()
RETURNING id;`

const reportingQuerySQL = `SELECT customer_id, count(*), max(created_at)
FROM public.orders
WHERE created_at >= now() - interval '30 days'
GROUP BY customer_id
ORDER BY max(created_at) DESC
LIMIT 500;`

const ordersForeignKeyQuerySQL = `ALTER TABLE public.orders
ADD CONSTRAINT orders_customer_fk
FOREIGN KEY (customer_id) REFERENCES public.customers(id);`
