package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thanoslabs/forma/internal/workspaces"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

const bridgeAdapterID = "forma-bridge"

type Execution struct {
	ConnectionID string `json:"connectionId"`
	orchestration.Execution
}

type ExecutionReader struct {
	connections workspaces.Repository
	client      *http.Client
}

func NewExecutionReader(connections workspaces.Repository, client *http.Client) *ExecutionReader {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ExecutionReader{connections: connections, client: client}
}

func (r *ExecutionReader) ListExecutions(ctx context.Context, workspaceID string) ([]Execution, error) {
	connections, err := r.connections.ListConnections(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := []Execution{}
	for _, summary := range connections {
		if summary.ProviderType != "orchestration" {
			continue
		}
		connection, err := r.connections.GetConnection(ctx, workspaceID, summary.ID)
		if err != nil {
			return nil, err
		}
		found, err := r.listBridgeExecutions(ctx, connection)
		if err != nil {
			return nil, fmt.Errorf("bridge connection %s: %w", connection.ID, err)
		}
		for _, execution := range found {
			result = append(result, Execution{ConnectionID: connection.ID, Execution: execution})
		}
	}
	return result, nil
}

func (r *ExecutionReader) listBridgeExecutions(ctx context.Context, connection workspaces.Connection) ([]orchestration.Execution, error) {
	if connection.AdapterID != bridgeAdapterID {
		return nil, fmt.Errorf("orchestration adapter %s is unavailable", connection.AdapterID)
	}
	endpoint, err := url.Parse(strings.TrimSpace(connection.EndpointURL))
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return nil, fmt.Errorf("bridge endpoint must be an absolute HTTP URL")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/executions"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+connection.Credential)
	response, err := r.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("observer returned %s", response.Status)
	}
	var payload struct {
		Executions []orchestration.Execution `json:"executions"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode observer response: %w", err)
	}
	for _, execution := range payload.Executions {
		if err := execution.Validate(); err != nil {
			return nil, fmt.Errorf("invalid execution %s: %w", execution.ID, err)
		}
	}
	return payload.Executions, nil
}
