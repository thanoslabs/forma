package projections

import (
	"context"
	"fmt"

	"github.com/thanoslabs/forma/internal/adapters/projects"
	"github.com/thanoslabs/forma/internal/adapters/projects/overwatch"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type Project struct {
	ConnectionID string `json:"connectionId"`
	projects.Project
}

type ProjectReader struct{ connections workspaces.Repository }

func NewProjectReader(connections workspaces.Repository) *ProjectReader {
	return &ProjectReader{connections: connections}
}

func (r *ProjectReader) ListProjects(ctx context.Context, workspaceID string) ([]Project, error) {
	connections, err := r.connections.ListConnections(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := []Project{}
	for _, connection := range connections {
		if connection.ProviderType != "project" {
			continue
		}
		adapter, err := r.resolve(connection)
		if err != nil {
			return nil, err
		}
		found, err := adapter.ListProjects(ctx)
		if err != nil {
			return nil, err
		}
		for _, project := range found {
			result = append(result, Project{ConnectionID: connection.ID, Project: project})
		}
	}
	return result, nil
}

func (r *ProjectReader) ListWorkItems(ctx context.Context, workspaceID, connectionID, projectID string) ([]projects.WorkItem, error) {
	connection, err := r.connections.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return nil, err
	}
	if connection.ProviderType != "project" {
		return nil, workspaces.ErrNotFound
	}
	adapter, err := r.resolve(connection)
	if err != nil {
		return nil, err
	}
	return adapter.ListWorkItems(ctx, projectID)
}

func (*ProjectReader) resolve(connection workspaces.Connection) (projects.Adapter, error) {
	switch connection.AdapterID {
	case "overwatch":
		return overwatch.New(overwatch.Config{URL: connection.EndpointURL, Token: connection.Credential})
	default:
		return nil, fmt.Errorf("project adapter %s is unavailable", connection.AdapterID)
	}
}
