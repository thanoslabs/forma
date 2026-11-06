package mock

import (
	"context"
	"testing"
)

func TestMockProjects(t *testing.T) {
	p := New()
	projects, err := p.ListProjects(context.Background())
	if err != nil || len(projects) != 1 {
		t.Fatalf("projects: %v %v", projects, err)
	}
	items, err := p.ListWorkItems(context.Background(), projects[0].ID)
	if err != nil || len(items) != 1 || items[0].ProjectID != projects[0].ID {
		t.Fatalf("items: %v %v", items, err)
	}
}
