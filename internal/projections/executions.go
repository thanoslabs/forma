package projections

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type OperatorCapabilities struct {
	Actions             []string `json:"actions"`
	IdempotencyRequired bool     `json:"idempotencyRequired"`
}

type bridgeCredentials struct {
	ObserverToken string `json:"observerToken"`
	OperatorToken string `json:"operatorToken"`
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
	request.Header.Set("Authorization", "Bearer "+credentials(connection.Credential).ObserverToken)
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

func (r *ExecutionReader) OperatorCapabilities(ctx context.Context, workspaceID, connectionID string) (OperatorCapabilities, error) {
	connection, err := r.bridgeConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return OperatorCapabilities{}, err
	}
	var result OperatorCapabilities
	err = r.operatorRequest(ctx, connection, http.MethodGet, "/api/v1/operator/capabilities", "", nil, &result)
	return result, err
}

func (r *ExecutionReader) ExecuteCommand(ctx context.Context, workspaceID, connectionID, executionID, action, key string) (Execution, error) {
	connection, err := r.bridgeConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return Execution{}, err
	}
	var result orchestration.Execution
	err = r.operatorRequest(ctx, connection, http.MethodPost, "/api/v1/executions/"+url.PathEscape(executionID)+"/commands", key, map[string]string{"action": action}, &result)
	if err != nil {
		return Execution{}, err
	}
	if err := result.Validate(); err != nil {
		return Execution{}, err
	}
	return Execution{ConnectionID: connectionID, Execution: result}, nil
}

func (r *ExecutionReader) bridgeConnection(ctx context.Context, workspaceID, connectionID string) (workspaces.Connection, error) {
	connection, err := r.connections.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return workspaces.Connection{}, err
	}
	if connection.ProviderType != "orchestration" || connection.AdapterID != bridgeAdapterID {
		return workspaces.Connection{}, workspaces.ErrNotFound
	}
	return connection, nil
}

func (r *ExecutionReader) operatorRequest(ctx context.Context, connection workspaces.Connection, method, path, key string, body any, result any) error {
	creds := credentials(connection.Credential)
	if creds.OperatorToken == "" {
		return errors.New("operator capability is not configured")
	}
	endpoint, err := bridgeURL(connection.EndpointURL, path)
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+creds.OperatorToken)
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response, err := r.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("operator returned %s", response.Status)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(result)
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
	request.Header.Set("Authorization", "Bearer "+credentials(connection.Credential).ObserverToken)
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

func credentials(value string) bridgeCredentials {
	var result bridgeCredentials
	if json.Unmarshal([]byte(value), &result) == nil && result.ObserverToken != "" {
		return result
	}
	return bridgeCredentials{ObserverToken: value}
}

func bridgeURL(base, path string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(base))
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return nil, fmt.Errorf("bridge endpoint must be an absolute HTTP URL")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	return endpoint, nil
}
