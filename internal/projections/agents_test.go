package projections

import (
	"context"
	"testing"

	"github.com/thanoslabs/forma/internal/adapters/agents/mock"
	"github.com/thanoslabs/forma/internal/turns"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type repository struct{ connections []workspaces.Connection }

func (r repository) ListConnections(_ context.Context, workspaceID string) ([]workspaces.Connection, error) {
	result := make([]workspaces.Connection, 0)
	for _, connection := range r.connections {
		if connection.WorkspaceID == workspaceID {
			result = append(result, connection)
		}
	}
	return result, nil
}
func (repository) ListWorkspaces(context.Context) ([]workspaces.Workspace, error) { return nil, nil }
func (repository) CreateWorkspace(context.Context, workspaces.Workspace) (workspaces.Workspace, error) {
	return workspaces.Workspace{}, nil
}
func (repository) CreateConnection(context.Context, workspaces.Connection) (workspaces.Connection, error) {
	return workspaces.Connection{}, nil
}
func (r repository) GetConnection(_ context.Context, workspaceID, connectionID string) (workspaces.Connection, error) {
	for _, connection := range r.connections {
		if connection.WorkspaceID == workspaceID && connection.ID == connectionID {
			return connection, nil
		}
	}
	return workspaces.Connection{}, workspaces.ErrNotFound
}

func TestAgentReaderScopesAndAnnotatesProviderAgents(t *testing.T) {
	reader := NewAgentReader(repository{connections: []workspaces.Connection{
		{ID: "agents-1", WorkspaceID: "workspace-1", AdapterID: "mock", ProviderType: "agent"},
		{ID: "chat-1", WorkspaceID: "workspace-1", AdapterID: "mock", ProviderType: "chat"},
		{ID: "agents-2", WorkspaceID: "workspace-2", AdapterID: "mock", ProviderType: "agent"},
	}}, turns.NewRegistry(mock.New()))
	agents, err := reader.ListAgents(context.Background(), "workspace-1")
	if err != nil || len(agents) != 3 {
		t.Fatalf("agents=%#v err=%v", agents, err)
	}
	for _, agent := range agents {
		if agent.ConnectionID != "agents-1" || agent.AdapterID != "mock" {
			t.Fatalf("unexpected projection: %#v", agent)
		}
	}
}
