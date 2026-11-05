// Package turns routes persisted conversation mentions to agent adapters.
package turns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thanoslabs/forma/internal/adapters/agents"
	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type Turn struct {
	ID                 string
	WorkspaceID        string
	ConnectionID       string
	ConversationID     string
	RequestMessageID   string
	AgentParticipantID string
	AgentID            string
	AgentName          string
	IdempotencyKey     string
}

type Store interface {
	StartTurn(context.Context, Turn) (Turn, bool, error)
	CompleteTurn(context.Context, Turn, string, time.Time) (conversations.Message, error)
	FailTurn(context.Context, Turn, error) error
}

type Publisher interface{ Publish(conversations.Message) }

type Registry struct{ adapters map[string]agents.Adapter }

func NewRegistry(adapters ...agents.Adapter) *Registry {
	registry := &Registry{adapters: make(map[string]agents.Adapter)}
	for _, adapter := range adapters {
		registry.adapters[adapter.ID()] = adapter
	}
	return registry
}

type Router struct {
	connections workspaces.Repository
	store       Store
	registry    *Registry
	publisher   Publisher
}

func NewRouter(connections workspaces.Repository, store Store, registry *Registry, publisher Publisher) *Router {
	return &Router{connections: connections, store: store, registry: registry, publisher: publisher}
}

func (r *Router) Route(ctx context.Context, workspaceID, requestKey string, message conversations.Message) error {
	var routeErrors []error
	seen := make(map[string]struct{})
	for _, mention := range message.Mentions {
		if mention.Kind != "agent" || mention.ConnectionID == "" {
			continue
		}
		deduplicationKey := mention.ConnectionID + ":" + mention.ID
		if _, ok := seen[deduplicationKey]; ok {
			continue
		}
		seen[deduplicationKey] = struct{}{}
		if err := r.routeMention(ctx, workspaceID, requestKey, message, mention); err != nil {
			routeErrors = append(routeErrors, err)
		}
	}
	return errors.Join(routeErrors...)
}

func (r *Router) routeMention(ctx context.Context, workspaceID, requestKey string, message conversations.Message, mention conversations.Mention) error {
	connection, err := r.connections.GetConnection(ctx, workspaceID, mention.ConnectionID)
	if err != nil {
		return err
	}
	if connection.ProviderType != "agent" {
		return fmt.Errorf("connection %s is not an agent provider", connection.ID)
	}
	adapter := r.registry.adapters[connection.AdapterID]
	sender, ok := adapter.(agents.InputSender)
	if !ok || !hasCapability(adapter, agents.Input) {
		return fmt.Errorf("adapter %s: %w", connection.AdapterID, agents.ErrUnsupported)
	}
	turn, created, err := r.store.StartTurn(ctx, Turn{WorkspaceID: workspaceID, ConnectionID: connection.ID, ConversationID: message.ConversationID, RequestMessageID: message.ID, AgentID: mention.ID, AgentName: mention.Name, IdempotencyKey: requestKey + ":" + connection.ID + ":" + mention.ID})
	if err != nil || !created {
		return err
	}
	reply, err := sender.SendInput(ctx, mention.ID, message.Body, turn.IdempotencyKey)
	if err != nil {
		if storeErr := r.store.FailTurn(ctx, turn, err); storeErr != nil {
			return errors.Join(err, storeErr)
		}
		return err
	}
	response, err := r.store.CompleteTurn(ctx, turn, reply, time.Now().UTC())
	if err != nil {
		return err
	}
	if r.publisher != nil {
		r.publisher.Publish(response)
	}
	return nil
}

func hasCapability(adapter agents.Adapter, expected agents.Capability) bool {
	for _, capability := range adapter.Capabilities() {
		if capability == expected {
			return true
		}
	}
	return false
}
