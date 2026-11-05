package httpapi

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/thanoslabs/forma/internal/adapters/agents"
	"github.com/thanoslabs/forma/internal/adapters/agents/mock"
	projectmock "github.com/thanoslabs/forma/internal/adapters/projects/mock"
)

type MockAPI struct {
	mu          sync.Mutex
	agents      *mock.Provider
	projects    *projectmock.Provider
	messages    []ConversationMessage
	idempotency map[string][]ConversationMessage
	nextMessage int
}

func NewMockAPI() *MockAPI {
	return &MockAPI{agents: mock.New(), projects: projectmock.New(), idempotency: make(map[string][]ConversationMessage), nextMessage: 1}
}

func (a *MockAPI) Register(mux *http.ServeMux) {
	a.registerConversations(mux)
	mux.HandleFunc("GET /api/v1/demo/agents", func(w http.ResponseWriter, r *http.Request) {
		data, err := a.agents.ListAgents(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		respond(w, data)
	})
	mux.HandleFunc("GET /api/v1/demo/projects", func(w http.ResponseWriter, r *http.Request) {
		data, err := a.projects.ListProjects(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		respond(w, data)
	})
	mux.HandleFunc("GET /api/v1/demo/projects/{id}/items", func(w http.ResponseWriter, r *http.Request) {
		data, err := a.projects.ListWorkItems(r.Context(), r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		respond(w, data)
	})
	mux.HandleFunc("POST /api/v1/demo/agents/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		// This endpoint is deliberately opt-in and for local demonstrations only.
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "expected application/json", 415)
			return
		}
		var body struct {
			Status agents.Status `json:"status"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		switch body.Status {
		case agents.Idle, agents.Running, agents.Blocked, agents.Completed, agents.Failed:
		default:
			http.Error(w, "invalid status", 400)
			return
		}
		a.mu.Lock()
		ok := a.agents.SetStatus(r.PathValue("id"), body.Status)
		a.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		respond(w, map[string]string{"status": string(body.Status)})
	})
}
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
