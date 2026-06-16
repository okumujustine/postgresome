package diagnosis

import (
	"github.com/okumujustine/postgresome/backend/internal/findingcontract"
	"github.com/okumujustine/postgresome/backend/internal/storage/repository"
)

type AlertPayload struct {
	FindingID         string                   `json:"finding_id"`
	LifecycleHint     string                   `json:"lifecycle_hint"`
	ProblemHeadline   string                   `json:"problem_headline"`
	AffectedResource  string                   `json:"affected_resource"`
	EvidenceSummary   string                   `json:"evidence_summary"`
	ImpactSummary     string                   `json:"impact_summary"`
	SuggestedNextStep string                   `json:"suggested_next_step"`
	PrimaryImpact     findingcontract.Impact   `json:"primary_impact"`
	SecondaryImpacts  []findingcontract.Impact `json:"secondary_impacts,omitempty"`
	PrimaryAction     findingcontract.Action   `json:"primary_action"`
	SecondaryActions  []findingcontract.Action `json:"secondary_actions,omitempty"`
}

// BuildAlertPayload turns a finding row into a transport-agnostic diagnosis
// alert payload. Delivery integrations can reuse this shape later without
// rebuilding summarization logic per provider.
func BuildAlertPayload(f repository.RecentFinding) AlertPayload {
	primaryImpact, secondaryImpacts, primaryAction, secondaryActions := findingcontract.FromRepository(f)
	return AlertPayload{
		FindingID:         f.ID,
		LifecycleHint:     alertLifecycleHint(f),
		ProblemHeadline:   firstNonEmptyString(f.ProblemSummary, f.Title, f.Message),
		AffectedResource:  firstNonEmptyString(f.ResourceName, f.RuleKey),
		EvidenceSummary:   firstNonEmptyString(f.EvidenceSummary, f.Message),
		ImpactSummary:     firstNonEmptyString(primaryImpact.Summary, f.ImpactSummary),
		SuggestedNextStep: firstNonEmptyString(primaryAction.Summary, f.SuggestedAction, f.Recommendation, f.VerificationHint),
		PrimaryImpact:     primaryImpact,
		SecondaryImpacts:  secondaryImpacts,
		PrimaryAction:     primaryAction,
		SecondaryActions:  secondaryActions,
	}
}

func alertLifecycleHint(f repository.RecentFinding) string {
	switch {
	case f.Status == "resolved" && f.VerificationStatus == "verified_fixed":
		return "verified fixed"
	case f.Status == "resolved":
		return "resolved"
	case f.VerificationStatus == "regressed":
		return "regressed"
	case f.VerificationStatus == "improving":
		return "improving"
	default:
		return "new"
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
