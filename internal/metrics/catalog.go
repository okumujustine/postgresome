package metrics

type Visualization string

const (
	VisualizationCard      Visualization = "card"
	VisualizationLineChart Visualization = "line_chart"
	VisualizationBarChart  Visualization = "bar_chart"
	VisualizationPieChart  Visualization = "pie_chart"
	VisualizationTable     Visualization = "table"
)

// MetricDefinition describes a metric Postgresome knows how to store and
// query as diagnostic evidence. It does not store any metric values.
type MetricDefinition struct {
	Key         string
	Label       string
	Description string
	Unit        string
	Category    string

	SupportedVisualizations []Visualization
}

var DefaultMetricCatalog = []MetricDefinition{
	{
		Key:         "active_connections",
		Label:       "Active Connections",
		Description: "Number of active connections to the database.",
		Unit:        "count",
		Category:    "connections",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "transaction_commits",
		Label:       "Transaction Commits",
		Description: "Number of transactions that have been committed.",
		Unit:        "count",
		Category:    "transactions",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "transaction_rollbacks",
		Label:       "Transaction Rollbacks",
		Description: "Number of transactions that have been rolled back.",
		Unit:        "count",
		Category:    "transactions",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "blocks_read_from_disk",
		Label:       "Blocks Read From Disk",
		Description: "Number of disk blocks read for this database.",
		Unit:        "count",
		Category:    "cache",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "blocks_hit_in_cache",
		Label:       "Blocks Hit In Cache",
		Description: "Number of times disk blocks were found in the buffer cache.",
		Unit:        "count",
		Category:    "cache",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "rows_scanned",
		Label:       "Rows Scanned",
		Description: "Number of rows scanned by queries in this database.",
		Unit:        "count",
		Category:    "reads",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "rows_fetched",
		Label:       "Rows Fetched",
		Description: "Number of rows fetched by queries in this database.",
		Unit:        "count",
		Category:    "reads",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "rows_inserted",
		Label:       "Rows Inserted",
		Description: "Number of rows inserted into this database.",
		Unit:        "count",
		Category:    "writes",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "rows_updated",
		Label:       "Rows Updated",
		Description: "Number of rows updated in this database.",
		Unit:        "count",
		Category:    "writes",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
	{
		Key:         "rows_deleted",
		Label:       "Rows Deleted",
		Description: "Number of rows deleted from this database.",
		Unit:        "count",
		Category:    "writes",
		SupportedVisualizations: []Visualization{
			VisualizationCard,
			VisualizationLineChart,
			VisualizationBarChart,
		},
	},
}
