package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductionGuardsEnforceOriginsAndHeaders(t *testing.T) {
	handler := ProductionGuards(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }), []string{"https://forma.example"})
	denied := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	denied.Header.Set("Origin", "https://evil.example")
	deniedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden {
		t.Fatalf("denied=%d", deniedRecorder.Code)
	}
	allowed := httptest.NewRequest(http.MethodOptions, "/api/v1/workspaces", nil)
	allowed.Header.Set("Origin", "https://forma.example")
	allowedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(allowedRecorder, allowed)
	if allowedRecorder.Code != http.StatusNoContent || allowedRecorder.Header().Get("Access-Control-Allow-Origin") == "" || allowedRecorder.Header().Get("X-Request-ID") == "" || allowedRecorder.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("headers=%#v", allowedRecorder.Header())
	}
}
