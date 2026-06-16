package diagnosis

import (
	"strings"
	"testing"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

func TestAssignImprovingLifecycle(t *testing.T) {
	t.Parallel()

	finding := analysis.Finding{
		VerificationStatus:  "regressed",
		VerificationSummary: "old summary",
	}

	assignImprovingLifecycle(&finding, true, "Things are trending down.")
	if finding.VerificationStatus != "improving" {
		t.Fatalf("expected improving status, got %q", finding.VerificationStatus)
	}
	if finding.VerificationSummary != "Things are trending down." {
		t.Fatalf("expected summary to be set, got %q", finding.VerificationSummary)
	}

	assignImprovingLifecycle(&finding, false, "ignored")
	if finding.VerificationStatus != "pending" {
		t.Fatalf("expected pending status, got %q", finding.VerificationStatus)
	}
	if finding.VerificationSummary != "" {
		t.Fatalf("expected summary to be cleared, got %q", finding.VerificationSummary)
	}
}

func TestNewerMaintenanceObserved(t *testing.T) {
	t.Parallel()

	seenAt := time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC)
	after := seenAt.Add(30 * time.Minute)
	before := seenAt.Add(-30 * time.Minute)

	tests := []struct {
		name     string
		row      repository.TableStatRow
		expected bool
	}{
		{
			name: "autovacuum after finding",
			row: repository.TableStatRow{
				LastAutoVacuumAt: &after,
			},
			expected: true,
		},
		{
			name: "manual vacuum after finding",
			row: repository.TableStatRow{
				LastVacuumAt: &after,
			},
			expected: true,
		},
		{
			name: "maintenance before finding",
			row: repository.TableStatRow{
				LastAutoVacuumAt: &before,
				LastVacuumAt:     &before,
			},
			expected: false,
		},
		{
			name:     "no maintenance timestamps",
			row:      repository.TableStatRow{},
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := newerMaintenanceObserved(tc.row, seenAt); got != tc.expected {
				t.Fatalf("expected %t, got %t", tc.expected, got)
			}
		})
	}
}

func TestEnrichTableFindingHighDeadRowsAddsDiagnosisNarrative(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	lastAutoVacuumAt := now.Add(-72 * time.Hour)
	firstAbnormalAt := now.Add(-36 * time.Hour)

	finding := analysis.Finding{
		RuleKey:        config.RuleKeyHighDeadRows,
		ResourceName:   "public.users",
		ThresholdValue: 1000,
	}
	table := repository.TableStatRow{
		SchemaName:        "public",
		TableName:         "users",
		DeadRows:          2400,
		SequentialScans:   18,
		LastAutoVacuumAt:  &lastAutoVacuumAt,
		LastAutoAnalyzeAt: &lastAutoVacuumAt,
	}
	history := repository.TableStatHistoricalContext{
		Current: &repository.TableStatHistoryPoint{
			DeadRows:        2400,
			SequentialScans: 18,
		},
		Previous: &repository.TableStatHistoryPoint{
			DeadRows:        3000,
			SequentialScans: 14,
		},
		BaselineDeadRows: 120,
		BaselineSeqScans: 9,
		FirstAbnormalAt:  &firstAbnormalAt,
	}

	enrichTableFinding(&finding, table, history, now)

	if finding.Title != "Table bloat risk detected" {
		t.Fatalf("expected title to be set, got %q", finding.Title)
	}
	if !strings.Contains(finding.EvidenceSummary, "public.users increased from a 120 baseline to 2400 now") {
		t.Fatalf("expected evidence summary to describe dead-row growth, got %q", finding.EvidenceSummary)
	}
	if finding.BaselineLabel != "Baseline dead tuples" {
		t.Fatalf("expected baseline label, got %q", finding.BaselineLabel)
	}
	if finding.VerificationStatus != "improving" {
		t.Fatalf("expected improving verification state, got %q", finding.VerificationStatus)
	}
	if finding.ConfidenceLabel != "high" {
		t.Fatalf("expected high confidence, got %q (score %.2f)", finding.ConfidenceLabel, finding.ConfidenceScore)
	}
}

func TestEnrichSlowQueryFindingAddsDiagnosisNarrative(t *testing.T) {
	t.Parallel()

	finding := analysis.Finding{
		RuleKey:        config.RuleKeySlowQuery,
		ResourceName:   "query-123",
		ThresholdValue: 150,
	}
	query := repository.QueryStatRow{
		QueryID:           "query-123",
		Calls:             32,
		MeanExecTimeMs:    220,
		SharedBlocksRead:  480,
		SharedBlocksHit:   1200,
		TotalExecTimeMs:   7040,
		RowsReturned:      12,
		DatabaseName:      "app",
		UserName:          "postgres",
		Query:             "select * from users where id = $1",
		TempBlocksWritten: 0,
	}
	firstAbnormalAt := time.Date(2026, 6, 14, 8, 0, 0, 0, time.UTC)
	history := repository.QueryStatHistoricalContext{
		Current: &repository.QueryStatHistoryPoint{
			MeanExecTimeMs:   220,
			Calls:            32,
			SharedBlocksRead: 480,
		},
		Previous: &repository.QueryStatHistoryPoint{
			MeanExecTimeMs:   260,
			Calls:            28,
			SharedBlocksRead: 600,
		},
		BaselineMeanExecMs: 90,
		BaselineCalls:      20,
		BaselineBlocksRead: 150,
		FirstAbnormalAt:    &firstAbnormalAt,
	}

	enrichSlowQueryFinding(&finding, query, history)

	if finding.Title != "Slow query regression detected" {
		t.Fatalf("expected slow-query title, got %q", finding.Title)
	}
	if !strings.Contains(finding.EvidenceSummary, "Mean execution time rose from 90.0 ms to 220.0 ms") {
		t.Fatalf("expected latency narrative, got %q", finding.EvidenceSummary)
	}
	if !strings.Contains(finding.SuggestedAction, "Inspect the execution plan") {
		t.Fatalf("expected action guidance, got %q", finding.SuggestedAction)
	}
	if finding.VerificationStatus != "improving" {
		t.Fatalf("expected improving lifecycle state, got %q", finding.VerificationStatus)
	}
	if finding.ConfidenceLabel != "high" {
		t.Fatalf("expected high confidence, got %q (score %.2f)", finding.ConfidenceLabel, finding.ConfidenceScore)
	}
}

func TestConfidenceLabelThresholds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		score    float64
		expected string
	}{
		{score: 0.90, expected: "high"},
		{score: 0.70, expected: "medium"},
		{score: 0.40, expected: "low"},
	}

	for _, tc := range tests {
		if got := confidenceLabel(tc.score); got != tc.expected {
			t.Fatalf("score %.2f: expected %q, got %q", tc.score, tc.expected, got)
		}
	}
}
