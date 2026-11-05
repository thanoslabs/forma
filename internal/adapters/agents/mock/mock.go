package mock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/thanoslabs/forma/internal/adapters/agents"
)

type Provider struct {
	mu       sync.RWMutex
	agents   []agents.Agent
	sessions map[string][]agents.Session
}

func New() *Provider {
	return &Provider{
		agents: []agents.Agent{
			{ID: "atlas", Name: "Atlas", Status: agents.Idle, LastSeen: time.Now().UTC()},
			{ID: "nova", Name: "Nova", Status: agents.Idle, LastSeen: time.Now().UTC()},
			{ID: "pixel", Name: "Pixel", Status: agents.Idle, LastSeen: time.Now().UTC()},
		},
		sessions: make(map[string][]agents.Session),
	}
}
func (*Provider) ID() string { return "mock" }
func (*Provider) Capabilities() []agents.Capability {
	return []agents.Capability{agents.Observe, agents.Sessions, agents.Input}
}
func (p *Provider) ListAgents(ctx context.Context) ([]agents.Agent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]agents.Agent(nil), p.agents...), nil
}
func (p *Provider) ListSessions(ctx context.Context, agentID string) ([]agents.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]agents.Session(nil), p.sessions[agentID]...), nil
}
func (p *Provider) SetStatus(id string, status agents.Status) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.agents {
		if p.agents[i].ID == id {
			p.agents[i].Status = status
			p.agents[i].LastSeen = time.Now().UTC()
			return true
		}
	}
	return false
}
func (p *Provider) SendInput(ctx context.Context, agentID, body, idempotencyKey string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if body == "" || idempotencyKey == "" {
		return "", errors.New("body and idempotency key are required")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, agent := range p.agents {
		if agent.ID == agentID {
			return agent.Name + " received your message: " + body, nil
		}
	}
	return "", errors.New("agent not found")
}
