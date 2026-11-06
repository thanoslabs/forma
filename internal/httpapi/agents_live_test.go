package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/thanoslabs/forma/internal/projections"
	"github.com/thanoslabs/forma/internal/realtime"
)

type staticAgentReader struct{ agents []projections.Agent }

func (r staticAgentReader) ListAgents(context.Context, string) ([]projections.Agent, error) {
	return r.agents, nil
}

func TestAgentWebSocketSendsInitialSnapshot(t *testing.T) {
	mux := http.NewServeMux()
	New(&memoryWorkspaceRepository{}, &memoryWorkspaceRepository{}, realtime.New()).WithAgentProjections(staticAgentReader{agents: []projections.Agent{{ConnectionID: "connection-1", ID: "atlas", Name: "Atlas"}}}).Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, strings.Replace(server.URL, "http://", "ws://", 1)+"/api/v1/workspaces/workspace-1/agents/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	var snapshot agentSnapshot
	if err := wsjson.Read(ctx, connection, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Type != "agents.snapshot" || snapshot.WorkspaceID != "workspace-1" || len(snapshot.Agents) != 1 || snapshot.Agents[0].ID != "atlas" {
		t.Fatalf("snapshot: %#v", snapshot)
	}
}
