package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/okumujustine/postgresome/internal/analysis"
	"github.com/okumujustine/postgresome/internal/metrics"
)

const requestTimeout = 10 * time.Second

// AgentAPIClient communicates with the Postgresome API on behalf of the agent.
type AgentAPIClient struct {
	APIBaseURL string
	httpClient *http.Client
}

// NewAgentAPIClient creates a client for the Postgresome API at apiBaseURL,
// e.g. "http://localhost:9090".
func NewAgentAPIClient(apiBaseURL string) *AgentAPIClient {
	return &AgentAPIClient{
		APIBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// AgentInfo identifies the agent during registration.
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
}

// DatabaseInfo identifies the monitored database during registration.
type DatabaseInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Version string `json:"version"`
}

// RegisterAgentRequest is the payload sent to POST /api/agents/register.
type RegisterAgentRequest struct {
	Agent    AgentInfo    `json:"agent"`
	Database DatabaseInfo `json:"database"`
}

// RegisterAgentResponse is the payload returned by POST /api/agents/register.
type RegisterAgentResponse struct {
	AgentID            string `json:"agent_id"`
	DatabaseInstanceID string `json:"database_instance_id"`
	Status             string `json:"status"`
}

// RegisterAgent registers the agent and the database it monitors with the
// Postgresome API.
func (c *AgentAPIClient) RegisterAgent(ctx context.Context, req RegisterAgentRequest) (*RegisterAgentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode registration request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBaseURL+"/api/agents/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registration request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result RegisterAgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	return &result, nil
}

// metricPointPayload is the JSON representation of a metrics.MetricPoint
// sent to POST /api/metrics/ingest.
type metricPointPayload struct {
	Key         string            `json:"key"`
	Label       string            `json:"label"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Category    string            `json:"category"`
	CollectedAt time.Time         `json:"collected_at"`
	Dimensions  map[string]string `json:"dimensions"`
}

// SendMetricsRequest is the payload sent to POST /api/metrics/ingest.
type SendMetricsRequest struct {
	AgentID            string               `json:"agent_id"`
	DatabaseInstanceID string               `json:"database_instance_id"`
	Metrics            []metricPointPayload `json:"metrics"`
}

// SendMetricsResponse is the payload returned by POST /api/metrics/ingest.
type SendMetricsResponse struct {
	Status   string `json:"status"`
	Inserted int    `json:"inserted"`
}

// SendMetrics sends collected metric points to the Postgresome API for the
// given agent and database instance.
func (c *AgentAPIClient) SendMetrics(ctx context.Context, agentID string, databaseInstanceID string, points []metrics.MetricPoint) (*SendMetricsResponse, error) {
	payload := make([]metricPointPayload, len(points))
	for i, point := range points {
		payload[i] = metricPointPayload{
			Key:         point.Key,
			Label:       point.Label,
			Value:       point.Value,
			Unit:        point.Unit,
			Category:    point.Category,
			CollectedAt: point.CollectedAt,
			Dimensions:  point.Dimensions,
		}
	}

	req := SendMetricsRequest{
		AgentID:            agentID,
		DatabaseInstanceID: databaseInstanceID,
		Metrics:            payload,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode metrics request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBaseURL+"/api/metrics/ingest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send metrics request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("metrics request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result SendMetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode metrics response: %w", err)
	}

	return &result, nil
}

// findingPayload is the JSON representation of an analysis.Finding sent to
// POST /api/findings/ingest.
type findingPayload struct {
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

// SendFindingsRequest is the payload sent to POST /api/findings/ingest.
type SendFindingsRequest struct {
	AgentID            string           `json:"agent_id"`
	DatabaseInstanceID string           `json:"database_instance_id"`
	Findings           []findingPayload `json:"findings"`
}

// SendFindingsResponse is the payload returned by POST /api/findings/ingest.
type SendFindingsResponse struct {
	Status   string `json:"status"`
	Inserted int    `json:"inserted"`
}

// tableStatPayload is the JSON representation of a metrics.TableStats sent
// to POST /api/tables/ingest.
type tableStatPayload struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`

	LiveRows int64 `json:"live_rows"`
	DeadRows int64 `json:"dead_rows"`

	SequentialScans    int64 `json:"sequential_scans"`
	SequentialRowsRead int64 `json:"sequential_rows_read"`

	IndexScans       int64 `json:"index_scans"`
	IndexRowsFetched int64 `json:"index_rows_fetched"`

	RowsInserted int64 `json:"rows_inserted"`
	RowsUpdated  int64 `json:"rows_updated"`
	RowsDeleted  int64 `json:"rows_deleted"`

	LastVacuumAt      *time.Time `json:"last_vacuum_at"`
	LastAutoVacuumAt  *time.Time `json:"last_autovacuum_at"`
	LastAnalyzeAt     *time.Time `json:"last_analyze_at"`
	LastAutoAnalyzeAt *time.Time `json:"last_autoanalyze_at"`

	VacuumCount      int64 `json:"vacuum_count"`
	AutoVacuumCount  int64 `json:"autovacuum_count"`
	AnalyzeCount     int64 `json:"analyze_count"`
	AutoAnalyzeCount int64 `json:"autoanalyze_count"`
}

// SendTableStatsRequest is the payload sent to POST /api/tables/ingest.
type SendTableStatsRequest struct {
	AgentID            string             `json:"agent_id"`
	DatabaseInstanceID string             `json:"database_instance_id"`
	CollectedAt        time.Time          `json:"collected_at"`
	Tables             []tableStatPayload `json:"tables"`
}

// SendTableStatsResponse is the payload returned by POST /api/tables/ingest.
type SendTableStatsResponse struct {
	Status string `json:"status"`
	Stored int    `json:"stored"`
}

// SendTableStats sends a table statistics snapshot to the Postgresome API
// for the given agent and database instance.
func (c *AgentAPIClient) SendTableStats(ctx context.Context, agentID string, databaseInstanceID string, snapshot *metrics.TableStatsSnapshot) (*SendTableStatsResponse, error) {
	payload := make([]tableStatPayload, len(snapshot.Tables))
	for i, table := range snapshot.Tables {
		payload[i] = tableStatPayload{
			SchemaName:         table.SchemaName,
			TableName:          table.TableName,
			LiveRows:           table.LiveRows,
			DeadRows:           table.DeadRows,
			SequentialScans:    table.SequentialScans,
			SequentialRowsRead: table.SequentialRowsRead,
			IndexScans:         table.IndexScans,
			IndexRowsFetched:   table.IndexRowsFetched,
			RowsInserted:       table.RowsInserted,
			RowsUpdated:        table.RowsUpdated,
			RowsDeleted:        table.RowsDeleted,
			LastVacuumAt:       table.LastVacuumAt,
			LastAutoVacuumAt:   table.LastAutoVacuumAt,
			LastAnalyzeAt:      table.LastAnalyzeAt,
			LastAutoAnalyzeAt:  table.LastAutoAnalyzeAt,
			VacuumCount:        table.VacuumCount,
			AutoVacuumCount:    table.AutoVacuumCount,
			AnalyzeCount:       table.AnalyzeCount,
			AutoAnalyzeCount:   table.AutoAnalyzeCount,
		}
	}

	req := SendTableStatsRequest{
		AgentID:            agentID,
		DatabaseInstanceID: databaseInstanceID,
		CollectedAt:        snapshot.CollectedAt,
		Tables:             payload,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode table stats request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBaseURL+"/api/tables/ingest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create table stats request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send table stats request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("table stats request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result SendTableStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode table stats response: %w", err)
	}

	return &result, nil
}

// queryStatPayload is the JSON representation of a metrics.QueryStats sent
// to POST /api/queries/ingest.
type queryStatPayload struct {
	QueryID      string `json:"query_id"`
	DatabaseName string `json:"database_name"`
	UserName     string `json:"user_name"`
	Query        string `json:"query"`

	Calls int64 `json:"calls"`

	TotalExecTimeMs float64 `json:"total_exec_time_ms"`
	MeanExecTimeMs  float64 `json:"mean_exec_time_ms"`
	MinExecTimeMs   float64 `json:"min_exec_time_ms"`
	MaxExecTimeMs   float64 `json:"max_exec_time_ms"`

	RowsReturned int64 `json:"rows_returned"`

	SharedBlocksHit     int64 `json:"shared_blocks_hit"`
	SharedBlocksRead    int64 `json:"shared_blocks_read"`
	SharedBlocksDirtied int64 `json:"shared_blocks_dirtied"`
	SharedBlocksWritten int64 `json:"shared_blocks_written"`

	TempBlocksRead    int64 `json:"temp_blocks_read"`
	TempBlocksWritten int64 `json:"temp_blocks_written"`
}

// SendQueryStatsRequest is the payload sent to POST /api/queries/ingest.
type SendQueryStatsRequest struct {
	AgentID            string             `json:"agent_id"`
	DatabaseInstanceID string             `json:"database_instance_id"`
	CollectedAt        time.Time          `json:"collected_at"`
	Queries            []queryStatPayload `json:"queries"`
}

// SendQueryStatsResponse is the payload returned by POST /api/queries/ingest.
type SendQueryStatsResponse struct {
	Status string `json:"status"`
	Stored int    `json:"stored"`
}

// SendQueryStats sends a query statistics snapshot to the Postgresome API
// for the given agent and database instance.
func (c *AgentAPIClient) SendQueryStats(ctx context.Context, agentID string, databaseInstanceID string, snapshot *metrics.QueryStatsSnapshot) (*SendQueryStatsResponse, error) {
	payload := make([]queryStatPayload, len(snapshot.Queries))
	for i, q := range snapshot.Queries {
		payload[i] = queryStatPayload{
			QueryID:             q.QueryID,
			DatabaseName:        q.DatabaseName,
			UserName:            q.UserName,
			Query:               q.Query,
			Calls:               q.Calls,
			TotalExecTimeMs:     q.TotalExecutionTimeMs,
			MeanExecTimeMs:      q.MeanExecutionTimeMs,
			MinExecTimeMs:       q.MinExecutionTimeMs,
			MaxExecTimeMs:       q.MaxExecutionTimeMs,
			RowsReturned:        q.RowsReturned,
			SharedBlocksHit:     q.SharedBlocksHit,
			SharedBlocksRead:    q.SharedBlocksRead,
			SharedBlocksDirtied: q.SharedBlocksDirtied,
			SharedBlocksWritten: q.SharedBlocksWritten,
			TempBlocksRead:      q.TempBlocksRead,
			TempBlocksWritten:   q.TempBlocksWritten,
		}
	}

	req := SendQueryStatsRequest{
		AgentID:            agentID,
		DatabaseInstanceID: databaseInstanceID,
		CollectedAt:        snapshot.CollectedAt,
		Queries:            payload,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query stats request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBaseURL+"/api/queries/ingest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create query stats request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send query stats request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query stats request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result SendQueryStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query stats response: %w", err)
	}

	return &result, nil
}

// SendFindings sends analyzer findings to the Postgresome API for the given
// agent and database instance. An empty findings slice is valid.
func (c *AgentAPIClient) SendFindings(ctx context.Context, agentID string, databaseInstanceID string, findings []analysis.Finding) (*SendFindingsResponse, error) {
	payload := make([]findingPayload, len(findings))
	for i, finding := range findings {
		payload[i] = findingPayload{
			Severity:       finding.Severity,
			Category:       finding.Category,
			Title:          finding.Title,
			Message:        finding.Message,
			Recommendation: finding.Recommendation,
		}
	}

	req := SendFindingsRequest{
		AgentID:            agentID,
		DatabaseInstanceID: databaseInstanceID,
		Findings:           payload,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode findings request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIBaseURL+"/api/findings/ingest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create findings request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send findings request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("findings request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result SendFindingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode findings response: %w", err)
	}

	return &result, nil
}
