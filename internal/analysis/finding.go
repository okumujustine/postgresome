package analysis

import "time"

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
}
