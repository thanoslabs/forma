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

type ConnectionCheck struct {
	ConnectionID string              `json:"connectionId"`
	Status       string              `json:"status"`
	AgentCount   int                 `json:"agentCount"`
	Capabilities []agents.Capability `json:"capabilities"`
	CheckedAt    time.Time           `json:"checkedAt"`
}

type AgentDetail struct {
	Agent        Agent               `json:"agent"`
	Capabilities []agents.Capability `json:"capabilities"`
	Sessions     []agents.Session    `json:"sessions"`
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
		resolvedConnection, err := r.connections.GetConnection(ctx, workspaceID, connection.ID)
		if err != nil {
			readErrors = append(readErrors, fmt.Errorf("connection %s: %w", connection.ID, err))
			continue
		}
		adapter, err := r.registry.Resolve(ctx, resolvedConnection)
		if err != nil {
			readErrors = append(readErrors, fmt.Errorf("connection %s: %w", connection.ID, err))
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

func (r *AgentReader) CheckConnection(ctx context.Context, workspaceID, connectionID string) (ConnectionCheck, error) {
	connection, err := r.connections.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return ConnectionCheck{}, err
	}
	if connection.ProviderType != "agent" {
		return ConnectionCheck{}, fmt.Errorf("connection %s is not an agent provider", connectionID)
	}
	adapter, err := r.registry.Resolve(ctx, connection)
	if err != nil {
		return ConnectionCheck{}, err
	}
	found, err := adapter.ListAgents(ctx)
	if err != nil {
		return ConnectionCheck{}, err
	}
	return ConnectionCheck{ConnectionID: connectionID, Status: "connected", AgentCount: len(found), Capabilities: adapter.Capabilities(), CheckedAt: time.Now().UTC()}, nil
}

func (r *AgentReader) GetAgent(ctx context.Context, workspaceID, connectionID, agentID string) (AgentDetail, error) {
	connection, err := r.connections.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return AgentDetail{}, err
	}
	if connection.ProviderType != "agent" {
		return AgentDetail{}, workspaces.ErrNotFound
	}
	adapter, err := r.registry.Resolve(ctx, connection)
	if err != nil {
		return AgentDetail{}, err
	}
	found, err := adapter.ListAgents(ctx)
	if err != nil {
		return AgentDetail{}, err
	}
	var selected *agents.Agent
	for i := range found {
		if found[i].ID == agentID {
			selected = &found[i]
			break
		}
	}
	if selected == nil {
		return AgentDetail{}, workspaces.ErrNotFound
	}
	sessions := []agents.Session{}
	if hasCapability(adapter.Capabilities(), agents.Sessions) {
		sessions, err = adapter.ListSessions(ctx, agentID)
		if err != nil {
			return AgentDetail{}, err
		}
		if sessions == nil {
			sessions = []agents.Session{}
		}
	}
	return AgentDetail{Agent: Agent{ConnectionID: connection.ID, AdapterID: connection.AdapterID, ID: selected.ID, Name: selected.Name, Status: selected.Status, LastSeen: selected.LastSeen}, Capabilities: adapter.Capabilities(), Sessions: sessions}, nil
}

func hasCapability(capabilities []agents.Capability, expected agents.Capability) bool {
	for _, capability := range capabilities {
		if capability == expected {
			return true
		}
	}
	return false
}
