package projections

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thanoslabs/forma/internal/workspaces"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

func TestExecutionReaderScopesAuthenticatesAndAnnotatesBridgeExecutions(t *testing.T) {
	now := time.Now().UTC()
	execution := orchestration.Execution{SchemaVersion: 1, ID: "execution-1", Attempt: 1, State: orchestration.StateRunning, Work: orchestration.WorkRef{ConnectionID: "overwatch-1", ProjectID: "project-1", ItemID: "item-1"}, Helpers: []orchestration.RuntimeRef{}, Results: []orchestration.ResultLink{}, CreatedAt: now, UpdatedAt: now}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/executions" || r.Header.Get("Authorization") != "Bearer observer-secret" {
			http.Error(w, "unexpected request", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"executions": []orchestration.Execution{execution}})
	}))
	defer server.Close()

	reader := NewExecutionReader(repository{connections: []workspaces.Connection{
		{ID: "bridge-1", WorkspaceID: "workspace-1", AdapterID: "forma-bridge", ProviderType: "orchestration", EndpointURL: server.URL, Credential: "observer-secret"},
		{ID: "bridge-2", WorkspaceID: "workspace-2", AdapterID: "forma-bridge", ProviderType: "orchestration", EndpointURL: server.URL, Credential: "observer-secret"},
		{ID: "agents-1", WorkspaceID: "workspace-1", AdapterID: "mock", ProviderType: "agent"},
	}}, server.Client())

	items, err := reader.ListExecutions(context.Background(), "workspace-1")
	if err != nil || len(items) != 1 || items[0].ConnectionID != "bridge-1" || items[0].ID != execution.ID {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

func TestExecutionReaderRejectsInvalidBridgePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"executions": []map[string]any{{"id": "invalid"}}})
	}))
	defer server.Close()
	reader := NewExecutionReader(repository{connections: []workspaces.Connection{{ID: "bridge-1", WorkspaceID: "workspace-1", AdapterID: "forma-bridge", ProviderType: "orchestration", EndpointURL: server.URL}}}, server.Client())
	if _, err := reader.ListExecutions(context.Background(), "workspace-1"); err == nil {
		t.Fatal("expected invalid bridge payload to fail")
	}
}
