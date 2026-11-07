package filestore

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

func execution(id string, state orchestration.State) orchestration.Execution {
	now := time.Now().UTC()
	return orchestration.Execution{SchemaVersion: orchestration.SchemaVersion, ID: id, Attempt: 1, State: state, Work: orchestration.WorkRef{ConnectionID: "overwatch", ProjectID: "repo", ItemID: "issue"}, Claim: &orchestration.ClaimRef{ID: "claim", ExpiresAt: now.Add(time.Minute)}, Helpers: []orchestration.RuntimeRef{}, Results: []orchestration.ResultLink{}, CreatedAt: now, UpdatedAt: now}
}

func TestStoreSurvivesRestartAndReplaysEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bridge", "state.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	first := execution("execution-1", orchestration.StateClaimed)
	if err := store.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	first.State = orchestration.StateRunning
	first.Parent = &orchestration.RuntimeRef{ConnectionID: "herdr", TerminalID: "t1"}
	first.UpdatedAt = first.UpdatedAt.Add(time.Second)
	if err := store.Update(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	active, err := reopened.ListActive(context.Background())
	if err != nil || len(active) != 1 || active[0].Parent.TerminalID != "t1" {
		t.Fatalf("active=%#v err=%v", active, err)
	}
	events, err := reopened.EventsAfter(context.Background(), 0)
	if err != nil || len(events) != 2 || events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("events=%#v err=%v", events, err)
	}
	remaining, _ := reopened.EventsAfter(context.Background(), 1)
	if len(remaining) != 1 || remaining[0].Sequence != 2 {
		t.Fatalf("remaining=%#v", remaining)
	}
}

func TestStoreOnlyListsNonTerminalExecutions(t *testing.T) {
	store, _ := Open(filepath.Join(t.TempDir(), "state.json"))
	_ = store.Create(context.Background(), execution("active", orchestration.StateClaimed))
	_ = store.Create(context.Background(), execution("done", orchestration.StateCompleted))
	active, err := store.ListActive(context.Background())
	if err != nil || len(active) != 1 || active[0].ID != "active" {
		t.Fatalf("active=%#v err=%v", active, err)
	}
}
