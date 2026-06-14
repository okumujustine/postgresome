package metrics

// DimensionDefinition describes a dimension that diagnostic evidence can be
// grouped or filtered by. It does not store any dimension values.
type DimensionDefinition struct {
	Key         string
	Label       string
	Description string
	Category    string

	Available bool
}

var DefaultDimensionCatalog = []DimensionDefinition{
	{
		Key:         "agent",
		Label:       "Agent",
		Description: "The Postgresome agent that collected the metric.",
		Category:    "infrastructure",
		Available:   true,
	},
	{
		Key:         "database_instance",
		Label:       "Database Instance",
		Description: "The PostgreSQL database instance being monitored.",
		Category:    "database",
		Available:   true,
	},
	{
		Key:         "environment",
		Label:       "Environment",
		Description: "The deployment environment such as development, staging, or production.",
		Category:    "infrastructure",
		Available:   true,
	},
	{
		Key:         "table",
		Label:       "Table",
		Description: "Database table-level metrics.",
		Category:    "database",
		Available:   false,
	},
	{
		Key:         "query",
		Label:       "Query",
		Description: "Query-level performance metrics.",
		Category:    "performance",
		Available:   false,
	},
	{
		Key:         "index",
		Label:       "Index",
		Description: "Database index performance metrics.",
		Category:    "database",
		Available:   false,
	},
}
