package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/realtime"
	"github.com/thanoslabs/forma/internal/workspaces"
)

func TestConversationWebSocketReplaysAndStreams(t *testing.T) {
	repository := &memoryWorkspaceRepository{
		workspaces:    []workspaces.Workspace{{ID: "workspace-1", Name: "Forma"}},
		conversations: []conversations.Conversation{{ID: "conversation-1", WorkspaceID: "workspace-1", Title: "Office"}},
		messages:      []conversations.Message{{ID: "message-1", ConversationID: "conversation-1", Body: "history", Author: conversations.Author{Kind: "user", ID: "local-user", Name: "You"}, Mentions: []conversations.Mention{}, OccurredAt: time.Now().UTC()}},
	}
	mux := http.NewServeMux()
	New(repository, repository, realtime.New()).Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, strings.Replace(server.URL, "http://", "ws://", 1)+"/api/v1/workspaces/workspace-1/conversations/conversation-1/messages/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	var replay conversations.Message
	if err := wsjson.Read(ctx, connection, &replay); err != nil || replay.ID != "message-1" {
		t.Fatalf("replay: %#v %v", replay, err)
	}

	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api/v1/workspaces/workspace-1/conversations/conversation-1/messages", bytes.NewBufferString(`{"body":"live","mentions":[]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "live-message")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create message: %d", response.StatusCode)
	}
	var live conversations.Message
	if err := wsjson.Read(ctx, connection, &live); err != nil || live.Body != "live" {
		t.Fatalf("live: %#v %v", live, err)
	}
}
