package rules

import (
	"regexp"

	"github.com/okumujustine/postgresome/backend/internal/metrics"
)

// walkPlan recursively visits node and all of its descendants in a
// PostgreSQL EXPLAIN (FORMAT JSON) plan tree.
func walkPlan(node metrics.PlanNode, visit func(metrics.PlanNode)) {
	visit(node)

	for _, child := range node.Plans {
		walkPlan(child, visit)
	}
}

// filterColumnPattern matches identifiers used on the left-hand side of a
// comparison in a plan node's Filter expression, e.g. "user_id" in
// "(user_id = $1)" or "status" in "((status)::text = 'active'::text)".
var filterColumnPattern = regexp.MustCompile(`\(?([a-zA-Z_][a-zA-Z0-9_]*)\)?(?:::[a-zA-Z_][a-zA-Z0-9_]*)?\s*(?:=|<>|!=|<=|>=|<|>|~~)`)

// extractFilterColumns is a best-effort scan of a plan node's Filter
// expression for column names used in comparisons, for use in CREATE INDEX
// suggestions. Duplicate column names are removed.
func extractFilterColumns(filter string) []string {
	if filter == "" {
		return nil
	}

	matches := filterColumnPattern.FindAllStringSubmatch(filter, -1)

	columns := make([]string, 0, len(matches))
	seen := make(map[string]bool, len(matches))

	for _, match := range matches {
		column := match[1]

		if seen[column] {
			continue
		}

		seen[column] = true
		columns = append(columns, column)
	}

	return columns
}
