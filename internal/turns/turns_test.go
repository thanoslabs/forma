package turns

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thanoslabs/forma/internal/adapters/agents/mock"
	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type connectionRepository struct{ connection workspaces.Connection }

func (r connectionRepository) GetConnection(_ context.Context, workspaceID, connectionID string) (workspaces.Connection, error) {
	if r.connection.WorkspaceID == workspaceID && r.connection.ID == connectionID {
		return r.connection, nil
	}
	return workspaces.Connection{}, workspaces.ErrNotFound
}
func (connectionRepository) ListWorkspaces(context.Context) ([]workspaces.Workspace, error) {
	return nil, nil
}
func (connectionRepository) CreateWorkspace(context.Context, workspaces.Workspace) (workspaces.Workspace, error) {
	return workspaces.Workspace{}, nil
}
func (connectionRepository) ListConnections(context.Context, string) ([]workspaces.Connection, error) {
	return nil, nil
}
func (connectionRepository) CreateConnection(context.Context, workspaces.Connection) (workspaces.Connection, error) {
	return workspaces.Connection{}, nil
}

type turnStore struct {
	turns     map[string]Turn
	completed []conversations.Message
	failed    error
}

func (s *turnStore) StartTurn(_ context.Context, turn Turn) (Turn, bool, error) {
	if existing, ok := s.turns[turn.IdempotencyKey]; ok {
		return existing, false, nil
	}
	turn.ID = "turn-1"
	turn.AgentParticipantID = "participant-1"
	s.turns[turn.IdempotencyKey] = turn
	return turn, true, nil
}
func (s *turnStore) CompleteTurn(_ context.Context, turn Turn, body string, occurredAt time.Time) (conversations.Message, error) {
	message := conversations.Message{ID: "reply-1", ConversationID: turn.ConversationID, Body: body, Author: conversations.Author{Kind: "agent", ID: turn.AgentID, Name: turn.AgentName}, OccurredAt: occurredAt, ReplyToMessageID: turn.RequestMessageID}
	s.completed = append(s.completed, message)
	return message, nil
}
func (s *turnStore) AcceptTurn(_ context.Context, _ Turn) error          { return nil }
func (s *turnStore) FailTurn(_ context.Context, _ Turn, err error) error { s.failed = err; return nil }

type publisher struct{ messages []conversations.Message }

func (p *publisher) Publish(message conversations.Message) { p.messages = append(p.messages, message) }

func TestRouterDispatchesConnectionScopedMentionOnce(t *testing.T) {
	store := &turnStore{turns: make(map[string]Turn)}
	published := &publisher{}
	router := NewRouter(connectionRepository{connection: workspaces.Connection{ID: "connection-1", WorkspaceID: "workspace-1", AdapterID: "mock", ProviderType: "agent"}}, store, NewRegistry(mock.New()), published)
	message := conversations.Message{ID: "message-1", ConversationID: "conversation-1", Body: "hello @Atlas", Mentions: []conversations.Mention{{Kind: "agent", ID: "atlas", ConnectionID: "connection-1", Name: "Atlas"}}}
	if err := router.Route(context.Background(), "workspace-1", "request-1", message); err != nil {
		t.Fatal(err)
	}
	if err := router.Route(context.Background(), "workspace-1", "request-1", message); err != nil {
		t.Fatal(err)
	}
	if len(store.completed) != 1 || len(published.messages) != 1 || store.completed[0].ReplyToMessageID != message.ID {
		t.Fatalf("completed=%#v published=%#v", store.completed, published.messages)
	}
}

func TestRouterIgnoresUnscopedMention(t *testing.T) {
	store := &turnStore{turns: make(map[string]Turn)}
	router := NewRouter(connectionRepository{}, store, NewRegistry(mock.New()), nil)
	err := router.Route(context.Background(), "workspace-1", "request-1", conversations.Message{Mentions: []conversations.Mention{{Kind: "agent", ID: "atlas"}}})
	if err != nil || len(store.turns) != 0 {
		t.Fatalf("turns=%#v err=%v", store.turns, err)
	}
}

func TestRouterRejectsCrossWorkspaceConnection(t *testing.T) {
	router := NewRouter(connectionRepository{connection: workspaces.Connection{ID: "connection-1", WorkspaceID: "other"}}, &turnStore{turns: make(map[string]Turn)}, NewRegistry(mock.New()), nil)
	err := router.Route(context.Background(), "workspace-1", "request-1", conversations.Message{Mentions: []conversations.Mention{{Kind: "agent", ID: "atlas", ConnectionID: "connection-1"}}})
	if !errors.Is(err, workspaces.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
