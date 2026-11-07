package bridgeapp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/bridge/filestore"
	"github.com/thanoslabs/forma/pkg/bridge/herdr"
	"github.com/thanoslabs/forma/pkg/bridge/overwatch"
)

type offlineHerdr struct {
	mu     sync.Mutex
	starts int
}

func (h *offlineHerdr) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, arg := range args {
		if arg == "run" {
			h.starts++
			return []byte(`{"address":"t1"}`), nil
		}
		if arg == "get" {
			return []byte(`{"address":"t1","state":"live","status":"done"}`), nil
		}
		if arg == "ls" {
			return []byte(`{"terminals":[{"address":"t1","name":"execution-1","state":"live","status":"done"}]}`), nil
		}
	}
	return nil, nil
}

func TestBridgeCompletesRuntimeLoopWithoutFormaAndRecovers(t *testing.T) {
	var mu sync.Mutex
	claims := 0
	reported := ""
	expires := time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.URL.Path == "/api/v1/agent/claim-next":
			claims++
			if claims > 1 {
				w.WriteHeader(204)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-1", "item_id": "issue-1", "repository": "repo", "title": "Work", "url": "https://example/1", "lease_expires_at": expires, "status": "claimed"})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-1", "lease_expires_at": expires, "status": "needs_review"})
		case r.Method == http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			reported, _ = body["status"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-1", "lease_expires_at": expires, "status": reported})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-1", "lease_expires_at": expires, "status": "in_progress"})
		}
	}))
	defer server.Close()
	statePath := filepath.Join(t.TempDir(), "state.json")
	store, _ := filestore.Open(statePath)
	work, _ := overwatch.New(overwatch.Config{URL: server.URL, ConnectionID: "overwatch", AgentName: "bridge", LeaseSeconds: 600})
	runtimeRunner := &offlineHerdr{}
	runtime, _ := herdr.New(herdr.Config{ConnectionID: "herdr", Command: []string{"codex"}, Runner: runtimeRunner})
	coordinator, _ := bridge.New(work, runtime, store, func() string { return "execution-1" })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	supervisor := NewSupervisor(coordinator, store, time.Millisecond, time.Millisecond, 1, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_ = supervisor.Run(ctx)
	mu.Lock()
	gotReported, gotClaims := reported, claims
	mu.Unlock()
	if gotReported != "needs_review" || gotClaims != 1 || runtimeRunner.starts != 1 {
		t.Fatalf("reported=%s claims=%d starts=%d", gotReported, gotClaims, runtimeRunner.starts)
	}
	reopened, _ := filestore.Open(statePath)
	saved, err := reopened.Get(context.Background(), "execution-1")
	if err != nil || saved.Parent == nil || saved.State != "needs_review" {
		t.Fatalf("saved=%#v err=%v", saved, err)
	}
}
