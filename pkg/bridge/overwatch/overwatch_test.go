package overwatch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/bridge"
)

func TestProviderClaimHeartbeatAndRelease(t *testing.T) {
	expires := time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339Nano)
	requests := make([]string, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing authorization")
		}
		if r.URL.Path == "/api/v1/agent/claim-next" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["kind"] != "issue" || body["lease_seconds"] != float64(600) {
				t.Fatalf("claim body: %#v", body)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "task-1", "item_id": "issue-1", "repository": "thanoslabs/forma", "title": "Implement bridge", "url": "https://example/issues/1", "lease_expires_at": expires})
	}))
	defer server.Close()
	provider, err := New(Config{URL: server.URL, Token: "secret", ConnectionID: "overwatch-1", AgentName: "forma-bridge", Kind: "issue", LeaseSeconds: 600})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := provider.ClaimNext(context.Background(), "execution-1")
	if err != nil || claimed.Work.ItemID != "issue-1" || claimed.Claim.ID != "task-1" {
		t.Fatalf("claimed=%#v err=%v", claimed, err)
	}
	if _, err := provider.Heartbeat(context.Background(), claimed.Work, claimed.Claim); err != nil {
		t.Fatal(err)
	}
	if err := provider.Release(context.Background(), claimed.Work, claimed.Claim, "test"); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 3 || requests[1] != "/api/v1/agent-tasks/task-1/heartbeat" || requests[2] != "/api/v1/agent-tasks/task-1/release" {
		t.Fatalf("requests: %v", requests)
	}
}

func TestProviderMapsNoEligibleWork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	provider, _ := New(Config{URL: server.URL, ConnectionID: "overwatch", AgentName: "bridge"})
	if _, err := provider.ClaimNext(context.Background(), "execution"); err != bridge.ErrNoWork {
		t.Fatalf("err=%v", err)
	}
}
