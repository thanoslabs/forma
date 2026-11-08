package bridgeapp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

type OperatorCommander interface {
	ExecuteCommand(context.Context, string, string) (orchestration.Execution, error)
}

type OperatorAPI struct {
	commander OperatorCommander
	token     string
	mu        sync.Mutex
	seen      map[string]orchestration.Execution
	auditor   interface {
		RecordOperatorCommand(context.Context, string, string, string) error
	}
}

func NewOperatorAPI(commander OperatorCommander, token string, auditors ...interface {
	RecordOperatorCommand(context.Context, string, string, string) error
}) *OperatorAPI {
	api := &OperatorAPI{commander: commander, token: token, seen: make(map[string]orchestration.Execution)}
	if len(auditors) > 0 {
		api.auditor = auditors[0]
	}
	return api
}
func (a *OperatorAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/operator/capabilities", a.auth(a.capabilities))
	mux.HandleFunc("POST /api/v1/executions/{executionID}/commands", a.auth(a.command))
}
func (a *OperatorAPI) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		given := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if a.token == "" || len(given) != len(a.token) || subtle.ConstantTimeCompare([]byte(given), []byte(a.token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
func (a *OperatorAPI) capabilities(w http.ResponseWriter, _ *http.Request) {
	respondObservation(w, map[string]any{"actions": []string{"release", "approve"}, "idempotencyRequired": true})
}
func (a *OperatorAPI) command(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" || len(key) > 200 {
		http.Error(w, "valid Idempotency-Key is required", http.StatusBadRequest)
		return
	}
	var request struct {
		Action string `json:"action"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || (request.Action != "release" && request.Action != "approve") {
		http.Error(w, "action must be release or approve", http.StatusBadRequest)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if execution, ok := a.seen[key]; ok {
		respondObservation(w, execution)
		return
	}
	execution, err := a.commander.ExecuteCommand(r.Context(), r.PathValue("executionID"), request.Action)
	if err != nil {
		http.Error(w, "command rejected", http.StatusConflict)
		return
	}
	if a.auditor != nil {
		if err := a.auditor.RecordOperatorCommand(r.Context(), execution.ID, request.Action, key); err != nil {
			http.Error(w, "command applied but audit persistence failed", http.StatusInternalServerError)
			return
		}
	}
	a.seen[key] = execution
	respondObservation(w, execution)
}
