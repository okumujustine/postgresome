package analysis

import (
	"fmt"
	"time"
)

type FindingEvidence struct {
	ObservedAt time.Time

	EvidenceType string
	Role         string
	Label        string
	Summary      string
	MetricKey    string
	ReferenceID  string

	CurrentValue    float64
	BaselineValue   float64
	ChangePercent   float64
	ConfidenceScore float64

	Metadata map[string]any
}

type Finding struct {
	ID string

	DetectedAt time.Time

	DatabaseInstanceID string
	AgentID            string

	Severity string
	Category string

	Title          string
	Message        string
	Recommendation string

	// Diagnosis-grade fields let the backend store structured issue
	// explanations without forcing every rule to speak in charts and raw
	// metrics.
	ProblemSummary      string
	EvidenceSummary     string
	ImpactSummary       string
	SuggestedAction     string
	ConfidenceLabel     string
	ConfidenceScore     float64
	BaselineValue       float64
	BaselineLabel       string
	ChangeSummary       string
	VerificationHint    string
	VerificationStatus  string
	VerificationSummary string

	// RuleKey, ResourceType, and ResourceName identify the underlying issue
	// this finding represents, independent of severity or message wording.
	RuleKey      string
	ResourceType string
	ResourceName string

	// CurrentValue and ThresholdValue record the measurement and the
	// threshold it crossed at the time this finding was detected.
	CurrentValue   float64
	ThresholdValue float64

	// Evidence links the diagnosis to the concrete PostgreSQL signals that
	// triggered or support it. The UI should treat these as the proof trail
	// for the finding rather than trying to infer relationships later.
	Evidence []FindingEvidence
}

// Fingerprint returns a stable identifier for the underlying issue this
// finding represents. Findings with the same fingerprint for the same
// database instance are treated as the same issue and deduplicated rather
// than recorded as separate rows.
func (f Finding) Fingerprint() string {
	return fmt.Sprintf("%s:%s:%s", f.RuleKey, f.ResourceType, f.ResourceName)
}
