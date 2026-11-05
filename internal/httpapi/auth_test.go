package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBearerProtectsAPIRoutes(t *testing.T) {
	handler := RequireBearer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }), "secret")
	for _, test := range []struct {
		path, header string
		want         int
	}{
		{path: "/api/v1/workspaces", want: http.StatusUnauthorized},
		{path: "/api/v1/workspaces", header: "Bearer wrong", want: http.StatusUnauthorized},
		{path: "/api/v1/workspaces", header: "Bearer secret", want: http.StatusNoContent},
		{path: "/healthz", want: http.StatusNoContent},
	} {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		request.Header.Set("Authorization", test.header)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != test.want {
			t.Fatalf("%s: got %d want %d", test.path, recorder.Code, test.want)
		}
	}
}
