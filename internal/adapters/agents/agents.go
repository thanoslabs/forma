// Package agents defines provider-neutral contracts for agent harnesses.
package agents

import (
	"context"
	"errors"
	"time"
)

var ErrUnsupported = errors.New("agent capability unsupported")

type Capability string

const (
	Observe  Capability = "observe"
	Sessions Capability = "sessions"
	Dispatch Capability = "dispatch"
	Input    Capability = "input"
	Cancel   Capability = "cancel"
	Stream   Capability = "stream"
)

type Status string

const (
	Unknown   Status = "unknown"
	Idle      Status = "idle"
	Running   Status = "running"
	Blocked   Status = "blocked"
	Completed Status = "completed"
	Failed    Status = "failed"
)

type Agent struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Status   Status    `json:"status"`
	LastSeen time.Time `json:"lastSeen"`
}

type Session struct {
	ID      string `json:"id"`
	AgentID string `json:"agentId"`
	Status  Status `json:"status"`
}

type Event struct {
	ID         string    `json:"id"`
	AgentID    string    `json:"agentId"`
	SessionID  string    `json:"sessionId,omitempty"`
	Status     Status    `json:"status"`
	OccurredAt time.Time `json:"occurredAt"`
}

type Adapter interface {
	ID() string
	Capabilities() []Capability
	ListAgents(ctx context.Context) ([]Agent, error)
	ListSessions(ctx context.Context, agentID string) ([]Session, error)
}
type Dispatcher interface {
	Dispatch(ctx context.Context, agentID, prompt, idempotencyKey string) (Session, error)
}
type InputSender interface {
	SendInput(ctx context.Context, agentID, body, idempotencyKey string) (string, error)
}
type AsyncInputSender interface {
	SendInputAsync(ctx context.Context, agentID, body, idempotencyKey string) error
}
type Subscriber interface {
	Subscribe(ctx context.Context, cursor string) (<-chan Event, error)
}
