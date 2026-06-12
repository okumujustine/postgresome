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
