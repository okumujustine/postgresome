package findingcontract

import (
	"testing"

	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
)

func TestBuild_TableBloat(t *testing.T) {
	primaryImpact, secondaryImpacts, primaryAction, secondaryActions := Build(Input{
		RuleKey:         config.RuleKeyHighDeadRows,
		ResourceType:    "table",
		ResourceName:    "public.users",
		ImpactSummary:   "Queries touching this table may become slower.",
		SuggestedAction: "Review autovacuum settings for public.users.",
	})

	if primaryImpact.Code != "query_latency_risk" {
		t.Fatalf("expected primary impact code query_latency_risk, got %q", primaryImpact.Code)
	}
	if primaryAction.Code != "review_autovacuum_settings" {
		t.Fatalf("expected primary action code review_autovacuum_settings, got %q", primaryAction.Code)
	}
	if len(secondaryImpacts) == 0 {
		t.Fatal("expected supporting impacts for high dead rows")
	}
	if len(secondaryActions) == 0 {
		t.Fatal("expected supporting actions for high dead rows")
	}
}

func TestBuild_Fallbacks(t *testing.T) {
	primaryImpact, secondaryImpacts, primaryAction, secondaryActions := Build(Input{
		RuleKey:      "custom.unknown",
		ResourceType: "query",
		ResourceName: "q-123",
	})

	if primaryImpact.Code != "database_health_risk" {
		t.Fatalf("expected fallback primary impact code, got %q", primaryImpact.Code)
	}
	if primaryAction.Code != "inspect_related_workload" {
		t.Fatalf("expected fallback primary action code, got %q", primaryAction.Code)
	}
	if primaryAction.Summary == "" {
		t.Fatal("expected fallback primary action summary")
	}
	if len(secondaryImpacts) != 0 || len(secondaryActions) != 0 {
		t.Fatal("expected no secondary items for fallback contract")
	}
}
