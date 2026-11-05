// Package projections reads provider-owned resources into workspace-scoped views.
package projections

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thanoslabs/forma/internal/adapters/agents"
	"github.com/thanoslabs/forma/internal/turns"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type Agent struct {
	ConnectionID string        `json:"connectionId"`
	AdapterID    string        `json:"adapterId"`
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Status       agents.Status `json:"status"`
	LastSeen     time.Time     `json:"lastSeen"`
}

type AgentReader struct {
	connections workspaces.Repository
	registry    *turns.Registry
}

func NewAgentReader(connections workspaces.Repository, registry *turns.Registry) *AgentReader {
	return &AgentReader{connections: connections, registry: registry}
}

func (r *AgentReader) ListAgents(ctx context.Context, workspaceID string) ([]Agent, error) {
	connections, err := r.connections.ListConnections(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]Agent, 0)
	var readErrors []error
	for _, connection := range connections {
		if connection.ProviderType != "agent" {
			continue
		}
		adapter, ok := r.registry.Adapter(connection.AdapterID)
		if !ok {
			readErrors = append(readErrors, fmt.Errorf("connection %s: adapter %s is unavailable", connection.ID, connection.AdapterID))
			continue
		}
		found, err := adapter.ListAgents(ctx)
		if err != nil {
			readErrors = append(readErrors, fmt.Errorf("connection %s: %w", connection.ID, err))
			continue
		}
		for _, agent := range found {
			result = append(result, Agent{ConnectionID: connection.ID, AdapterID: connection.AdapterID, ID: agent.ID, Name: agent.Name, Status: agent.Status, LastSeen: agent.LastSeen})
		}
	}
	return result, errors.Join(readErrors...)
}
