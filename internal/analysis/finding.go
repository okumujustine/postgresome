package analysis

import (
	"fmt"
	"time"
)

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

	// RuleKey, ResourceType, and ResourceName identify the underlying issue
	// this finding represents, independent of severity or message wording.
	RuleKey      string
	ResourceType string
	ResourceName string

	// CurrentValue and ThresholdValue record the measurement and the
	// threshold it crossed at the time this finding was detected.
	CurrentValue   float64
	ThresholdValue float64
}

// Fingerprint returns a stable identifier for the underlying issue this
// finding represents. Findings with the same fingerprint for the same
// database instance are treated as the same issue and deduplicated rather
// than recorded as separate rows.
func (f Finding) Fingerprint() string {
	return fmt.Sprintf("%s:%s:%s", f.RuleKey, f.ResourceType, f.ResourceName)
}
