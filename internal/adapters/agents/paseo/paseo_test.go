package paseo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/thanoslabs/forma/internal/adapters/agents"
)

func TestProviderUsesPaseoWebSocketRPC(t *testing.T) {
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		connection, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer connection.CloseNow()
		ctx := r.Context()
		_, hello, err := connection.Read(ctx)
		if err != nil {
			return
		}
		var helloMessage map[string]any
		_ = json.Unmarshal(hello, &helloMessage)
		if helloMessage["type"] != "hello" {
			return
		}
		_, payload, err := connection.Read(ctx)
		if err != nil {
			return
		}
		var envelope struct {
			Message map[string]any `json:"message"`
		}
		_ = json.Unmarshal(payload, &envelope)
		requests <- envelope.Message
		requestID, _ := envelope.Message["requestId"].(string)
		var response any
		switch envelope.Message["type"] {
		case "fetch_agents_request":
			response = map[string]any{"type": "fetch_agents_response", "payload": map[string]any{"requestId": requestID, "entries": []any{map[string]any{"agent": map[string]any{"id": "agent-1", "title": "Paseo Agent", "status": "running", "updatedAt": "2026-11-05T12:00:00Z"}, "project": nil}}, "pageInfo": map[string]any{"nextCursor": nil, "prevCursor": nil, "hasMore": false}}}
		case "send_agent_message_request":
			response = map[string]any{"type": "send_agent_message_response", "payload": map[string]any{"requestId": requestID, "agentId": "agent-1", "accepted": true, "error": nil}}
		}
		_ = writeJSON(ctx, connection, map[string]any{"type": "session", "message": response})
	}))
	defer server.Close()
	provider, err := New(Config{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Password: "secret", Timeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	found, err := provider.ListAgents(context.Background())
	if err != nil || len(found) != 1 || found[0].Name != "Paseo Agent" || found[0].Status != agents.Running {
		t.Fatalf("agents=%#v err=%v", found, err)
	}
	if err := provider.SendInputAsync(context.Background(), "agent-1", "hello", "message-key"); err != nil {
		t.Fatal(err)
	}
	<-requests
	sent := <-requests
	if sent["messageId"] != "message-key" || sent["text"] != "hello" {
		t.Fatalf("request=%#v", sent)
	}
}

func TestProviderValidatesConfiguration(t *testing.T) {
	if _, err := New(Config{URL: "https://example.com"}); err == nil {
		t.Fatal("expected invalid WebSocket URL")
	}
}
