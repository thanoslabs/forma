package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type memoryWorkspaceRepository struct {
	workspaces    []workspaces.Workspace
	connections   []workspaces.Connection
	conversations []conversations.Conversation
	messages      []conversations.Message
}

func (m *memoryWorkspaceRepository) ListConversations(_ context.Context, workspaceID string) ([]conversations.Conversation, error) {
	result := make([]conversations.Conversation, 0)
	for _, item := range m.conversations {
		if item.WorkspaceID == workspaceID {
			result = append(result, item)
		}
	}
	return result, nil
}
func (m *memoryWorkspaceRepository) CreateConversation(_ context.Context, item conversations.Conversation) (conversations.Conversation, error) {
	found := false
	for _, workspace := range m.workspaces {
		found = found || workspace.ID == item.WorkspaceID
	}
	if !found {
		return conversations.Conversation{}, conversations.ErrNotFound
	}
	item.CreatedAt = time.Now().UTC()
	item.UpdatedAt = item.CreatedAt
	m.conversations = append(m.conversations, item)
	return item, nil
}
func (m *memoryWorkspaceRepository) ListMessages(_ context.Context, workspaceID, conversationID string) ([]conversations.Message, error) {
	result := make([]conversations.Message, 0)
	for _, message := range m.messages {
		if message.ConversationID == conversationID {
			result = append(result, message)
		}
	}
	return result, nil
}
func (m *memoryWorkspaceRepository) CreateMessage(_ context.Context, workspaceID string, message conversations.Message) (conversations.Message, error) {
	found := false
	for _, item := range m.conversations {
		found = found || item.ID == message.ConversationID && item.WorkspaceID == workspaceID
	}
	if !found {
		return conversations.Message{}, conversations.ErrNotFound
	}
	m.messages = append(m.messages, message)
	return message, nil
}

func (m *memoryWorkspaceRepository) ListWorkspaces(context.Context) ([]workspaces.Workspace, error) {
	return append([]workspaces.Workspace(nil), m.workspaces...), nil
}
func (m *memoryWorkspaceRepository) CreateWorkspace(_ context.Context, workspace workspaces.Workspace) (workspaces.Workspace, error) {
	workspace.CreatedAt = time.Now().UTC()
	m.workspaces = append(m.workspaces, workspace)
	return workspace, nil
}
func (m *memoryWorkspaceRepository) ListConnections(_ context.Context, workspaceID string) ([]workspaces.Connection, error) {
	result := make([]workspaces.Connection, 0)
	for _, connection := range m.connections {
		if connection.WorkspaceID == workspaceID {
			result = append(result, connection)
		}
	}
	return result, nil
}
func (m *memoryWorkspaceRepository) CreateConnection(_ context.Context, connection workspaces.Connection) (workspaces.Connection, error) {
	found := false
	for _, workspace := range m.workspaces {
		found = found || workspace.ID == connection.WorkspaceID
	}
	if !found {
		return workspaces.Connection{}, workspaces.ErrNotFound
	}
	connection.CreatedAt = time.Now().UTC()
	m.connections = append(m.connections, connection)
	return connection, nil
}

func TestWorkspaceAPI(t *testing.T) {
	repository := &memoryWorkspaceRepository{}
	mux := http.NewServeMux()
	New(repository, repository).Register(mux)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", bytes.NewBufferString(`{"name":"Forma"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create workspace: %d %s", recorder.Code, recorder.Body.String())
	}
	var workspace workspaces.Workspace
	if err := json.Unmarshal(recorder.Body.Bytes(), &workspace); err != nil || workspace.ID == "" || workspace.Name != "Forma" {
		t.Fatalf("workspace: %#v %v", workspace, err)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+workspace.ID+"/connections", bytes.NewBufferString(`{"adapterId":"mock","providerType":"agent","displayName":"Mock agents"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create connection: %d %s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspace.ID+"/connections", nil))
	var connections []workspaces.Connection
	if err := json.Unmarshal(recorder.Body.Bytes(), &connections); err != nil || len(connections) != 1 || connections[0].AdapterID != "mock" {
		t.Fatalf("connections: %#v %v", connections, err)
	}
}

func TestWorkspaceAPIRejectsInvalidConnection(t *testing.T) {
	mux := http.NewServeMux()
	repository := &memoryWorkspaceRepository{}
	New(repository, repository).Register(mux)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/missing/connections", bytes.NewBufferString(`{"adapterId":"mock","providerType":"unknown","displayName":"Mock"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
