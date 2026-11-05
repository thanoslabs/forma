package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/workspaces"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPersistedConversationAPI(t *testing.T) {
	repository := &memoryWorkspaceRepository{workspaces: []workspaces.Workspace{{ID: "workspace-1", Name: "Forma"}}}
	mux := http.NewServeMux()
	New(repository, repository, nil).Register(mux)
	request := httptest.NewRequest("POST", "/api/v1/workspaces/workspace-1/conversations", bytes.NewBufferString(`{"title":"Office"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != 201 {
		t.Fatalf("create conversation: %d %s", recorder.Code, recorder.Body.String())
	}
	var conversation conversations.Conversation
	if err := json.Unmarshal(recorder.Body.Bytes(), &conversation); err != nil || conversation.ID == "" {
		t.Fatalf("conversation: %#v %v", conversation, err)
	}
	request = httptest.NewRequest("POST", "/api/v1/workspaces/workspace-1/conversations/"+conversation.ID+"/messages", bytes.NewBufferString(`{"body":"hello","mentions":[]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != 201 {
		t.Fatalf("create message: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/workspaces/workspace-1/conversations/"+conversation.ID+"/messages", nil))
	var messages []conversations.Message
	if err := json.Unmarshal(recorder.Body.Bytes(), &messages); err != nil || len(messages) != 1 || messages[0].Body != "hello" {
		t.Fatalf("messages: %#v %v", messages, err)
	}
}
