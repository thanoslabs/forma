package overwatch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderReadsRepositoriesAndIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing bearer token")
		}
		switch r.URL.Path {
		case "/api/v1/repositories":
			_, _ = w.Write([]byte(`[{"id":"repo-1","full_name":"thanoslabs/forma"}]`))
		case "/api/v1/issues":
			if r.URL.Query().Get("repository_id") != "repo-1" {
				t.Fatal("missing repository filter")
			}
			_, _ = w.Write([]byte(`{"issues":[{"id":"issue-1","title":"Ship it","state":"open","url":"https://example/1","updated_at":"2026-11-06T12:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	provider, err := New(Config{URL: server.URL, Token: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	found, err := provider.ListProjects(context.Background())
	if err != nil || len(found) != 1 || found[0].Name != "thanoslabs/forma" {
		t.Fatalf("projects=%#v err=%v", found, err)
	}
	items, err := provider.ListWorkItems(context.Background(), found[0].ID)
	if err != nil || len(items) != 1 || items[0].ID != "issue-1" {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
