package diagnosis

import (
	"fmt"
	"strings"

	"github.com/okumujustine/postgresome/backend/internal/analysis"
	"github.com/okumujustine/postgresome/backend/internal/analysis/config"
	"github.com/okumujustine/postgresome/backend/internal/findingcontract"
)

func normalizeFindings(findings []analysis.Finding) {
	for i := range findings {
		normalizeFinding(&findings[i])
	}
}

func normalizeFinding(f *analysis.Finding) {
	if f == nil {
		return
	}

	normalizeResource(f)
	normalizeConfidence(f)
	normalizeNarrative(f)
	normalizeImpactAndAction(f)
	normalizeVerification(f)
	normalizeRecommendation(f)
}

func normalizeResource(f *analysis.Finding) {
	if strings.TrimSpace(f.ResourceType) == "" {
		f.ResourceType = "database"
	}
	if strings.TrimSpace(f.ResourceName) != "" {
		return
	}

	switch f.ResourceType {
	case "database":
		f.ResourceName = "database"
	case "query":
		f.ResourceName = "query"
	case "table":
		f.ResourceName = "table"
	default:
		f.ResourceName = f.ResourceType
	}
}

func normalizeConfidence(f *analysis.Finding) {
	if f.ConfidenceScore <= 0 {
		switch strings.ToLower(f.Severity) {
		case "critical":
			f.ConfidenceScore = 0.88
		case "warning":
			f.ConfidenceScore = 0.74
		default:
			f.ConfidenceScore = 0.64
		}
	}

	if strings.TrimSpace(f.ConfidenceLabel) == "" {
		f.ConfidenceLabel = confidenceLabel(f.ConfidenceScore)
	}
}

func normalizeNarrative(f *analysis.Finding) {
	if strings.TrimSpace(f.Title) == "" {
		f.Title = genericTitle(f.RuleKey, f.ResourceName)
	}
	if strings.TrimSpace(f.Message) == "" {
		f.Message = genericMessage(f.RuleKey, f.ResourceName)
	}
	if strings.TrimSpace(f.ProblemSummary) == "" {
		f.ProblemSummary = genericProblemSummary(f)
	}
	if strings.TrimSpace(f.EvidenceSummary) == "" {
		f.EvidenceSummary = genericEvidenceSummary(f)
	}
	if strings.TrimSpace(f.BaselineLabel) == "" && f.ThresholdValue > 0 {
		f.BaselineLabel = "Diagnosis threshold"
	}
	if strings.TrimSpace(f.ChangeSummary) == "" {
		f.ChangeSummary = genericChangeSummary(f)
	}
}

func normalizeImpactAndAction(f *analysis.Finding) {
	primaryImpact, _, primaryAction, _ := findingcontract.FromAnalysis(*f)

	if strings.TrimSpace(f.ImpactSummary) == "" {
		f.ImpactSummary = primaryImpact.Summary
	}
	if strings.TrimSpace(f.SuggestedAction) == "" {
		f.SuggestedAction = primaryAction.Summary
	}
}

func normalizeVerification(f *analysis.Finding) {
	if strings.TrimSpace(f.VerificationStatus) == "" {
		f.VerificationStatus = "pending"
	}

	if strings.TrimSpace(f.VerificationHint) == "" {
		f.VerificationHint = genericVerificationHint(f)
	}

	if strings.TrimSpace(f.VerificationSummary) == "" {
		switch f.VerificationStatus {
		case "improving":
			f.VerificationSummary = "The underlying signals are moving in the right direction, but the finding is still above the diagnosis threshold."
		case "regressed":
			f.VerificationSummary = "This finding reappeared after previously improving or resolving, so the underlying issue likely persists."
		case "verified_fixed":
			f.VerificationSummary = "The supporting signals have returned to a healthier range and the issue is no longer active."
		default:
			f.VerificationSummary = "Use the recommended verification checks to confirm the finding no longer appears after remediation."
		}
	}
}

func normalizeRecommendation(f *analysis.Finding) {
	if strings.TrimSpace(f.Recommendation) == "" {
		f.Recommendation = f.SuggestedAction
	}
	if strings.TrimSpace(f.Message) == "" {
		f.Message = f.ProblemSummary
	}
}

func genericTitle(ruleKey, resourceName string) string {
	switch ruleKey {
	case config.RuleKeyHighConnections:
		return "Connection capacity pressure detected"
	case config.RuleKeyLowCacheHitRatio:
		return "Low cache hit ratio detected"
	case config.RuleKeyHighRollbackRate:
		return "High rollback rate detected"
	case config.RuleKeyIdleConnections:
		return "High idle connection usage detected"
	case config.RuleKeyLongRunningQuery:
		return "Long-running query detected"
	case config.RuleKeyBlockedQuery:
		return "Lock contention detected"
	case config.RuleKeyHighSequentialScan:
		return "High sequential scan activity detected"
	case config.RuleKeyExpensiveQuery:
		return "Expensive query workload detected"
	case config.RuleKeyDiskHeavyQuery:
		return "Disk-heavy query detected"
	case "missing_index":
		return "Missing index likely"
	default:
		if resourceName != "" {
			return fmt.Sprintf("Database issue detected on %s", resourceName)
		}
		return "Database issue detected"
	}
}

func genericMessage(ruleKey, resourceName string) string {
	switch ruleKey {
	case config.RuleKeyHighConnections:
		return "The database is using more active connections than its current workload normally tolerates."
	case config.RuleKeyLowCacheHitRatio:
		return "PostgreSQL is serving a larger share of reads from disk instead of cache."
	case config.RuleKeyHighRollbackRate:
		return "A larger share of transactions is rolling back instead of committing successfully."
	case config.RuleKeyIdleConnections:
		return "A large share of database sessions is idle or idle in transaction."
	case config.RuleKeyLongRunningQuery:
		return "One or more queries are running longer than expected."
	case config.RuleKeyBlockedQuery:
		return "Sessions are waiting on locks held by other transactions."
	case config.RuleKeyHighSequentialScan:
		return fmt.Sprintf("Reads against %s are relying heavily on sequential scans.", resourceName)
	case config.RuleKeyExpensiveQuery:
		return fmt.Sprintf("Query %s is consuming a large share of total database time.", resourceName)
	case config.RuleKeyDiskHeavyQuery:
		return fmt.Sprintf("Query %s is reading heavily from disk instead of cache.", resourceName)
	case "missing_index":
		return fmt.Sprintf("The planner is using a sequential scan on %s where an index may be missing.", resourceName)
	default:
		return "Postgresome detected a database issue that needs investigation."
	}
}

func genericProblemSummary(f *analysis.Finding) string {
	resource := humanResourceName(f)
	switch f.RuleKey {
	case config.RuleKeyHighConnections:
		return "Active connection usage on this database is above the expected safe range."
	case config.RuleKeyLowCacheHitRatio:
		return "Cache efficiency is below the expected range, so more reads are likely hitting disk."
	case config.RuleKeyHighRollbackRate:
		return "Transactions are failing often enough to suggest application or workload instability."
	case config.RuleKeyIdleConnections:
		return "Too many connections are sitting idle, which reduces headroom for useful work."
	case config.RuleKeyLongRunningQuery:
		return fmt.Sprintf("%s is taking longer than expected to finish.", resource)
	case config.RuleKeyBlockedQuery:
		return "Lock waits are blocking database work and increasing latency risk."
	case config.RuleKeyHighSequentialScan:
		return fmt.Sprintf("%s is reading through sequential scans more often than expected.", resource)
	case config.RuleKeyExpensiveQuery:
		return fmt.Sprintf("%s is consuming a disproportionate amount of database time.", resource)
	case config.RuleKeyDiskHeavyQuery:
		return fmt.Sprintf("%s is relying on disk reads more than expected.", resource)
	case "missing_index":
		return fmt.Sprintf("%s appears to be missing index support for a costly filter path.", resource)
	default:
		return firstNonEmpty(f.Message, fmt.Sprintf("%s needs investigation.", resource))
	}
}

func genericEvidenceSummary(f *analysis.Finding) string {
	current := formatNumericValue(f.CurrentValue, f.RuleKey)
	threshold := formatNumericValue(f.ThresholdValue, f.RuleKey)

	switch {
	case f.ThresholdValue > 0:
		return fmt.Sprintf("Observed value %s crossed the diagnosis threshold of %s for %s.", current, threshold, humanResourceName(f))
	case f.CurrentValue > 0:
		return fmt.Sprintf("Observed value %s triggered this diagnosis for %s.", current, humanResourceName(f))
	default:
		return firstNonEmpty(f.Message, "The supporting PostgreSQL signals crossed the rule conditions for this diagnosis.")
	}
}

func genericChangeSummary(f *analysis.Finding) string {
	current := formatNumericValue(f.CurrentValue, f.RuleKey)
	if f.ThresholdValue > 0 {
		return fmt.Sprintf("The current signal is %s versus a diagnosis threshold of %s.", current, formatNumericValue(f.ThresholdValue, f.RuleKey))
	}
	return fmt.Sprintf("The current signal is %s.", current)
}

func genericVerificationHint(f *analysis.Finding) string {
	if f.ThresholdValue > 0 {
		return fmt.Sprintf("After remediation, confirm the relevant signal for %s falls back below %s and this finding stops recurring.", humanResourceName(f), formatNumericValue(f.ThresholdValue, f.RuleKey))
	}
	return fmt.Sprintf("After remediation, confirm the supporting signals for %s return to a healthier range and the finding does not reappear.", humanResourceName(f))
}

func humanResourceName(f *analysis.Finding) string {
	name := strings.TrimSpace(f.ResourceName)
	if name == "" || name == "database" {
		switch f.ResourceType {
		case "database":
			return "this database"
		case "query":
			return "this query"
		case "table":
			return "this table"
		default:
			return "this resource"
		}
	}
	return name
}

func formatNumericValue(value float64, ruleKey string) string {
	switch ruleKey {
	case config.RuleKeyLowCacheHitRatio, config.RuleKeyHighRollbackRate:
		return fmt.Sprintf("%.2f", value)
	default:
		return fmt.Sprintf("%.0f", value)
	}
}
