package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thanoslabs/forma/internal/adapters/agents"
)

func TestMockAPI(t *testing.T) {
	mux := http.NewServeMux()
	NewMockAPI().Register(mux)
	for _, path := range []string{"/api/v1/demo/agents", "/api/v1/demo/projects", "/api/v1/demo/projects/demo/items"} {
		r := httptest.NewRecorder()
		mux.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 {
			t.Fatalf("%s: %d", path, r.Code)
		}
	}
	body, _ := json.Marshal(map[string]string{"status": "running"})
	req := httptest.NewRequest("POST", "/api/v1/demo/agents/atlas/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r := httptest.NewRecorder()
	mux.ServeHTTP(r, req)
	if r.Code != 200 {
		t.Fatalf("status update: %d", r.Code)
	}
	r = httptest.NewRecorder()
	mux.ServeHTTP(r, httptest.NewRequest("GET", "/api/v1/demo/agents", nil))
	var list []agents.Agent
	if err := json.Unmarshal(r.Body.Bytes(), &list); err != nil || len(list) != 3 || list[0].Status != agents.Running {
		t.Fatalf("agents: %v %v", list, err)
	}
}
func TestInvalidStatus(t *testing.T) {
	mux := http.NewServeMux()
	NewMockAPI().Register(mux)
	req := httptest.NewRequest("POST", "/api/v1/demo/agents/atlas/status", bytes.NewBufferString(`{"status":"bogus"}`))
	req.Header.Set("Content-Type", "application/json")
	r := httptest.NewRecorder()
	mux.ServeHTTP(r, req)
	if r.Code != 400 {
		t.Fatalf("expected 400, got %d", r.Code)
	}
}

func TestConversationMessages(t *testing.T) {
	mux := http.NewServeMux()
	NewMockAPI().Register(mux)
	body := `{"body":"hello @Atlas","mentions":[{"kind":"agent","id":"atlas","name":"Atlas","text":"@Atlas","start":6,"end":12}]}`
	send := func() sendConversationMessageResponse {
		req := httptest.NewRequest("POST", "/api/v1/demo/conversations/default/messages", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Idempotency-Key", "message-request-1")
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("send: %d %s", recorder.Code, recorder.Body.String())
		}
		var response sendConversationMessageResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response
	}
	first := send()
	if len(first.Messages) != 2 || first.Messages[1].Author.ID != "atlas" || first.Messages[1].ReplyToMessageID != first.Messages[0].ID {
		t.Fatalf("messages: %#v", first.Messages)
	}
	replay := send()
	if len(replay.Messages) != 2 || replay.Messages[0].ID != first.Messages[0].ID {
		t.Fatalf("idempotent replay: %#v", replay.Messages)
	}
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/demo/conversations/default/messages", nil))
	var messages []ConversationMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &messages); err != nil || len(messages) != 2 {
		t.Fatalf("history: %#v %v", messages, err)
	}
}

func TestConversationRequiresIdempotencyKey(t *testing.T) {
	mux := http.NewServeMux()
	NewMockAPI().Register(mux)
	req := httptest.NewRequest("POST", "/api/v1/demo/conversations/default/messages", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
