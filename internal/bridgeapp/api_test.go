package bridgeapp

import (
	"context"
	"github.com/thanoslabs/forma/pkg/orchestration"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type observationMemory struct{}

func (observationMemory) List(context.Context) ([]orchestration.Execution, error) {
	return []orchestration.Execution{{ID: "execution-1"}}, nil
}
func (observationMemory) EventsAfter(_ context.Context, after uint64) ([]orchestration.Event, error) {
	return []orchestration.Event{{Sequence: after + 1}}, nil
}
func TestObservationAPIRequiresTokenAndReplaysCursor(t *testing.T) {
	mux := http.NewServeMux()
	NewObservationAPI(observationMemory{}, "secret").Register(mux)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/events?after=4", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != 401 {
		t.Fatalf("unauthorized=%d", recorder.Code)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/events?after=4", nil)
	request.Header.Set("Authorization", "Bearer secret")
	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), `"next":5`) {
		t.Fatalf("response=%d %s", recorder.Code, recorder.Body.String())
	}
}
