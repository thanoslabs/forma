package bridgeapp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

type ObservationStore interface {
	List(context.Context) ([]orchestration.Execution, error)
	EventsAfter(context.Context, uint64) ([]orchestration.Event, error)
}
type ObservationAPI struct {
	store ObservationStore
	token string
}

func NewObservationAPI(store ObservationStore, token string) *ObservationAPI {
	return &ObservationAPI{store: store, token: token}
}
func (a *ObservationAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/executions", a.auth(a.executions))
	mux.HandleFunc("GET /api/v1/events", a.auth(a.events))
}
func (a *ObservationAPI) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		given := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(given) != len(a.token) || subtle.ConstantTimeCompare([]byte(given), []byte(a.token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
func (a *ObservationAPI) executions(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.List(r.Context())
	if err != nil {
		http.Error(w, "unable to list executions", 500)
		return
	}
	respondObservation(w, map[string]any{"executions": items})
}
func (a *ObservationAPI) events(w http.ResponseWriter, r *http.Request) {
	after, err := strconv.ParseUint(r.URL.Query().Get("after"), 10, 64)
	if r.URL.Query().Get("after") == "" {
		after = 0
		err = nil
	}
	if err != nil {
		http.Error(w, "after must be an unsigned integer", 400)
		return
	}
	items, err := a.store.EventsAfter(r.Context(), after)
	if err != nil {
		http.Error(w, "unable to replay events", 500)
		return
	}
	if len(items) > 500 {
		items = items[:500]
	}
	next := after
	if len(items) > 0 {
		next = items[len(items)-1].Sequence
	}
	respondObservation(w, map[string]any{"events": items, "next": next})
}
func respondObservation(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
