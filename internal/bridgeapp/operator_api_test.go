package bridgeapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

type commandRecorder struct{ calls int }

func (c *commandRecorder) ExecuteCommand(_ context.Context, id, action string) (orchestration.Execution, error) {
	c.calls++
	now := time.Now().UTC()
	return orchestration.Execution{SchemaVersion: 1, ID: id, Attempt: 1, State: orchestration.StateReleased, Work: orchestration.WorkRef{ConnectionID: "overwatch", ProjectID: "project", ItemID: action}, Helpers: []orchestration.RuntimeRef{}, Results: []orchestration.ResultLink{}, CreatedAt: now, UpdatedAt: now}, nil
}

func TestOperatorAPIRequiresSeparateTokenAndIdempotency(t *testing.T) {
	commander := &commandRecorder{}
	mux := http.NewServeMux()
	NewOperatorAPI(commander, "operator-secret").Register(mux)

	unauthorized := httptest.NewRecorder()
	mux.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/operator/capabilities", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", unauthorized.Code)
	}

	for range 2 {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/executions/execution-1/commands", strings.NewReader(`{"action":"release"}`))
		request.Header.Set("Authorization", "Bearer operator-secret")
		request.Header.Set("Idempotency-Key", "command-1")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	}
	if commander.calls != 1 {
		t.Fatalf("calls=%d", commander.calls)
	}
}
