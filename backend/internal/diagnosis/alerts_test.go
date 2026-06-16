package diagnosis

import (
	"testing"
	"time"

	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

func TestBuildAlertPayloadUsesDiagnosisFields(t *testing.T) {
	t.Parallel()

	finding := repository.RecentFinding{
		ID:                 "finding-1",
		Status:             "open",
		VerificationStatus: "improving",
		ProblemSummary:     "Table bloat risk detected",
		ResourceName:       "public.users",
		EvidenceSummary:    "Dead tuples increased from baseline and scans are rising.",
		ImpactSummary:      "Queries on this table may become slower.",
		SuggestedAction:    "Review autovacuum settings.",
	}

	payload := BuildAlertPayload(finding)

	if payload.FindingID != "finding-1" {
		t.Fatalf("expected finding id to be preserved, got %q", payload.FindingID)
	}
	if payload.LifecycleHint != "improving" {
		t.Fatalf("expected improving lifecycle hint, got %q", payload.LifecycleHint)
	}
	if payload.ProblemHeadline != finding.ProblemSummary {
		t.Fatalf("expected problem summary headline, got %q", payload.ProblemHeadline)
	}
	if payload.AffectedResource != "public.users" {
		t.Fatalf("expected affected resource, got %q", payload.AffectedResource)
	}
	if payload.PrimaryImpact.Code != "database_health_risk" && payload.PrimaryImpact.Code != "query_latency_risk" {
		t.Fatalf("expected structured primary impact, got %q", payload.PrimaryImpact.Code)
	}
	if payload.PrimaryAction.Code == "" {
		t.Fatal("expected structured primary action code")
	}
	if payload.SuggestedNextStep != finding.SuggestedAction {
		t.Fatalf("expected suggested action, got %q", payload.SuggestedNextStep)
	}
}

func TestBuildAlertPayloadFallsBackForResolvedFinding(t *testing.T) {
	t.Parallel()

	finding := repository.RecentFinding{
		ID:                 "finding-2",
		Status:             "resolved",
		VerificationStatus: "verified_fixed",
		Title:              "Slow query regression detected",
		Message:            "A query exceeded its latency threshold.",
		RuleKey:            "slow_query",
		Recommendation:     "Inspect the plan.",
	}

	payload := BuildAlertPayload(finding)

	if payload.LifecycleHint != "verified fixed" {
		t.Fatalf("expected verified fixed lifecycle hint, got %q", payload.LifecycleHint)
	}
	if payload.ProblemHeadline != finding.Title {
		t.Fatalf("expected title fallback headline, got %q", payload.ProblemHeadline)
	}
	if payload.AffectedResource != finding.RuleKey {
		t.Fatalf("expected rule key fallback resource, got %q", payload.AffectedResource)
	}
	if payload.EvidenceSummary != finding.Message {
		t.Fatalf("expected message fallback evidence summary, got %q", payload.EvidenceSummary)
	}
	if payload.PrimaryAction.Code == "" {
		t.Fatal("expected fallback structured primary action")
	}
	if payload.SuggestedNextStep != finding.Recommendation {
		t.Fatalf("expected recommendation fallback, got %q", payload.SuggestedNextStep)
	}
}

func TestAlertLifecycleHintPriorities(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name     string
		finding  repository.RecentFinding
		expected string
	}{
		{
			name: "resolved verified fixed",
			finding: repository.RecentFinding{
				Status:             "resolved",
				VerificationStatus: "verified_fixed",
				VerifiedFixedAt:    &now,
			},
			expected: "verified fixed",
		},
		{
			name: "resolved pending",
			finding: repository.RecentFinding{
				Status:             "resolved",
				VerificationStatus: "pending",
			},
			expected: "resolved",
		},
		{
			name: "regressed open",
			finding: repository.RecentFinding{
				Status:             "open",
				VerificationStatus: "regressed",
			},
			expected: "regressed",
		},
		{
			name: "fresh open",
			finding: repository.RecentFinding{
				Status:             "open",
				VerificationStatus: "pending",
			},
			expected: "new",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := alertLifecycleHint(tc.finding); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
