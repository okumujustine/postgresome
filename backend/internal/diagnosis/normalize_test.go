package diagnosis

import (
	"strings"
	"testing"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
)

func TestNormalizeFinding_PopulatesDiagnosisFieldsForDatabaseRule(t *testing.T) {
	t.Parallel()

	finding := analysis.Finding{
		Severity:       "warning",
		RuleKey:        config.RuleKeyHighConnections,
		ResourceType:   "database",
		CurrentValue:   186,
		ThresholdValue: 100,
		Title:          "High database connections detected",
		Message:        "Database has an unusually high number of active connections, which may indicate connection leaks, missing pooling, or traffic spikes.",
		Recommendation: "Investigate connection spikes, idle connections, and connection pooling configuration.",
	}

	normalizeFinding(&finding)

	if finding.ResourceName != "database" {
		t.Fatalf("expected normalized database resource name, got %q", finding.ResourceName)
	}
	if finding.ProblemSummary == "" {
		t.Fatal("expected problem summary to be populated")
	}
	if !strings.Contains(finding.EvidenceSummary, "186") {
		t.Fatalf("expected evidence summary to include current value, got %q", finding.EvidenceSummary)
	}
	if finding.ImpactSummary == "" {
		t.Fatal("expected impact summary to be populated")
	}
	if finding.SuggestedAction == "" {
		t.Fatal("expected suggested action to be populated")
	}
	if finding.VerificationHint == "" {
		t.Fatal("expected verification hint to be populated")
	}
	if finding.VerificationStatus != "pending" {
		t.Fatalf("expected pending verification status, got %q", finding.VerificationStatus)
	}
	if finding.ConfidenceLabel == "" || finding.ConfidenceScore == 0 {
		t.Fatalf("expected confidence to be populated, got %q / %.2f", finding.ConfidenceLabel, finding.ConfidenceScore)
	}
}

func TestNormalizeFinding_PopulatesDiagnosisFieldsForExplainRule(t *testing.T) {
	t.Parallel()

	finding := analysis.Finding{
		Severity:       "critical",
		RuleKey:        "missing_index",
		ResourceType:   "query_plan",
		ResourceName:   "query-1:public.orders",
		CurrentValue:   90000,
		ThresholdValue: 50000,
		Title:          "Sequential scan on large table",
		Message:        "Query performs a sequential scan on a large relation.",
		Recommendation: "Consider adding an index.",
	}

	normalizeFinding(&finding)

	if finding.ProblemSummary == "" {
		t.Fatal("expected problem summary")
	}
	if !strings.Contains(finding.EvidenceSummary, "90000") {
		t.Fatalf("expected evidence summary to reference current value, got %q", finding.EvidenceSummary)
	}
	if finding.SuggestedAction == "" {
		t.Fatal("expected suggested action")
	}
	if finding.BaselineLabel != "Diagnosis threshold" {
		t.Fatalf("expected baseline label fallback, got %q", finding.BaselineLabel)
	}
}

func TestNormalizeFinding_PreservesEnrichedNarrative(t *testing.T) {
	t.Parallel()

	finding := analysis.Finding{
		Severity:            "warning",
		RuleKey:             config.RuleKeySlowQuery,
		ResourceType:        "query",
		ResourceName:        "query-123",
		CurrentValue:        220,
		ThresholdValue:      100,
		ProblemSummary:      "Query query-123 is materially slower than its recent baseline.",
		EvidenceSummary:     "Mean execution time rose from 90.0 ms to 220.0 ms.",
		ImpactSummary:       "If this query remains on the hot path, it can increase request latency.",
		SuggestedAction:     "Inspect the execution plan.",
		ConfidenceScore:     0.91,
		ConfidenceLabel:     "high",
		VerificationStatus:  "improving",
		VerificationSummary: "Latency is moving down, but the query is still above threshold.",
		VerificationHint:    "Confirm mean execution time trends back toward baseline.",
	}

	normalizeFinding(&finding)

	if finding.ProblemSummary != "Query query-123 is materially slower than its recent baseline." {
		t.Fatalf("expected enriched problem summary to be preserved, got %q", finding.ProblemSummary)
	}
	if finding.VerificationStatus != "improving" {
		t.Fatalf("expected lifecycle state to be preserved, got %q", finding.VerificationStatus)
	}
	if finding.ConfidenceScore != 0.91 {
		t.Fatalf("expected enriched confidence score to be preserved, got %.2f", finding.ConfidenceScore)
	}
}
