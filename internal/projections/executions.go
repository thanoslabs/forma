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

type ExecutionEvents struct {
	Events []orchestration.Event `json:"events"`
	Next   uint64                `json:"next"`
}

type ExecutionReader struct {
	connections workspaces.Repository
	client      *http.Client
}

func (r *ExecutionReader) ListEvents(ctx context.Context, workspaceID, connectionID, executionID string, after uint64) (ExecutionEvents, error) {
	connection, err := r.connections.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return ExecutionEvents{}, err
	}
	if connection.ProviderType != "orchestration" || connection.AdapterID != bridgeAdapterID {
		return ExecutionEvents{}, workspaces.ErrNotFound
	}
	endpoint, err := bridgeURL(connection.EndpointURL, "/api/v1/events")
	if err != nil {
		return ExecutionEvents{}, err
	}
	query := endpoint.Query()
	query.Set("after", fmt.Sprintf("%d", after))
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ExecutionEvents{}, err
	}
	request.Header.Set("Authorization", "Bearer "+connection.Credential)
	response, err := r.client.Do(request)
	if err != nil {
		return ExecutionEvents{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return ExecutionEvents{}, fmt.Errorf("observer returned %s", response.Status)
	}
	var payload ExecutionEvents
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&payload); err != nil {
		return ExecutionEvents{}, fmt.Errorf("decode observer response: %w", err)
	}
	filtered := make([]orchestration.Event, 0, len(payload.Events))
	for _, event := range payload.Events {
		if err := event.Validate(); err != nil {
			return ExecutionEvents{}, fmt.Errorf("invalid event %s: %w", event.ID, err)
		}
		if event.ExecutionID == executionID {
			filtered = append(filtered, event)
		}
	}
	payload.Events = filtered
	return payload, nil
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
	endpoint, err := bridgeURL(connection.EndpointURL, "/api/v1/executions")
	if err != nil {
		return nil, err
	}
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

func bridgeURL(base, path string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(base))
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return nil, fmt.Errorf("bridge endpoint must be an absolute HTTP URL")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	return endpoint, nil
}
