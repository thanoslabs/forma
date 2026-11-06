package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thanoslabs/forma/internal/realtime"
	"github.com/thanoslabs/forma/internal/workspaces"
	"github.com/thanoslabs/forma/internal/worlds"
)

type memoryWorlds struct{ config worlds.Configuration }

func (m *memoryWorlds) GetWorld(context.Context, string) (worlds.Configuration, error) {
	if m.config.Version == 0 {
		return worlds.Configuration{}, workspaces.ErrNotFound
	}
	return m.config, nil
}
func (m *memoryWorlds) PutWorld(_ context.Context, workspaceID string, config worlds.Configuration) (worlds.Configuration, error) {
	config.WorkspaceID = workspaceID
	m.config = config
	return config, nil
}

func TestWorldConfigurationRoundTripAndValidation(t *testing.T) {
	repository := &memoryWorkspaceRepository{}
	worldRepository := &memoryWorlds{}
	mux := http.NewServeMux()
	New(repository, repository, realtime.New()).WithWorlds(worldRepository).Register(mux)
	valid := `{"version":1,"desks":[{"x":220,"y":205},{"x":430,"y":205},{"x":640,"y":205}]}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/workspace-1/world", bytes.NewBufferString(valid))
	request.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || worldRepository.config.WorkspaceID != "workspace-1" {
		t.Fatalf("put: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/workspace-1/world", nil))
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(`"version":1`)) {
		t.Fatalf("get: %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/workspace-1/world", bytes.NewBufferString(`{"version":1,"desks":[]}`))
	request.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid: %d", recorder.Code)
	}
}
