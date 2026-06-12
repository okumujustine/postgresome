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
