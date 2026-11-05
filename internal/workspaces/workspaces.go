package workspaces

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("workspace not found")

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Connection struct {
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"workspaceId"`
	AdapterID    string    `json:"adapterId"`
	ProviderType string    `json:"providerType"`
	DisplayName  string    `json:"displayName"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Repository interface {
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
	CreateWorkspace(ctx context.Context, workspace Workspace) (Workspace, error)
	ListConnections(ctx context.Context, workspaceID string) ([]Connection, error)
	CreateConnection(ctx context.Context, connection Connection) (Connection, error)
	GetConnection(ctx context.Context, workspaceID, connectionID string) (Connection, error)
}
