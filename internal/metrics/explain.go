package metrics

import "time"

// PlanNode represents a single node in a PostgreSQL EXPLAIN (FORMAT JSON)
// plan tree, including the fields relevant to plan-based analysis rules.
type PlanNode struct {
	NodeType     string `json:"Node Type"`
	RelationName string `json:"Relation Name,omitempty"`
	Alias        string `json:"Alias,omitempty"`
	IndexName    string `json:"Index Name,omitempty"`
	JoinType     string `json:"Join Type,omitempty"`

	StartupCost float64 `json:"Startup Cost"`
	TotalCost   float64 `json:"Total Cost"`
	PlanRows    float64 `json:"Plan Rows"`
	PlanWidth   float64 `json:"Plan Width"`

	Filter  string   `json:"Filter,omitempty"`
	SortKey []string `json:"Sort Key,omitempty"`

	Plans []PlanNode `json:"Plans,omitempty"`
}

// ExplainPlan is the EXPLAIN (FORMAT JSON, GENERIC_PLAN) plan tree for a
// single query from pg_stat_statements.
type ExplainPlan struct {
	QueryID string
	Query   string
	Root    PlanNode
}

// ExplainSnapshot holds the EXPLAIN plans collected for the top queries in a
// single collection cycle.
type ExplainSnapshot struct {
	CollectedAt time.Time

	DatabaseInstanceID string

	Plans []ExplainPlan
}
