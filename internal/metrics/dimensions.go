package metrics

type AgentDimension struct {
	AgentID      string
	AgentName    string
	AgentVersion string
	Environment  string
}

type DatabaseInstanceDimension struct {
	DatabaseInstanceID string
	AgentID            string
	DatabaseName       string
	Host               string
	Port               int
	PostgresVersion    string
}
