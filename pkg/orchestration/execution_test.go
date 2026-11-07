package orchestration

import (
	"encoding/json"
	"testing"
	"time"
)

func validExecution() Execution {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	return Execution{SchemaVersion: SchemaVersion, ID: "execution-1", Attempt: 1, State: StateRunning, Work: WorkRef{ConnectionID: "overwatch-1", ProjectID: "repo-1", ItemID: "issue-1"}, Parent: &RuntimeRef{ConnectionID: "herdr-1", MachineID: "local", TerminalID: "terminal-1"}, Helpers: []RuntimeRef{}, Results: []ResultLink{}, CreatedAt: now, UpdatedAt: now}
}

func TestExecutionValidation(t *testing.T) {
	execution := validExecution()
	if err := execution.Validate(); err != nil {
		t.Fatal(err)
	}
	execution.Helpers = []RuntimeRef{{ConnectionID: "herdr-1", TerminalID: "helper-1"}, {ConnectionID: "herdr-1", TerminalID: "helper-1"}}
	if err := execution.Validate(); err == nil {
		t.Fatal("expected duplicate helper to fail")
	}
}

func TestTerminalStates(t *testing.T) {
	execution := validExecution()
	if execution.Terminal() {
		t.Fatal("running execution is not terminal")
	}
	execution.State = StateCompleted
	if !execution.Terminal() {
		t.Fatal("completed execution is terminal")
	}
}

func TestEventValidationAndJSON(t *testing.T) {
	event := Event{SchemaVersion: SchemaVersion, ID: "event-1", ExecutionID: "execution-1", Type: EventExecutionChanged, Sequence: 1, ObservedAt: time.Now().UTC(), Payload: json.RawMessage(`{"state":"running"}`)}
	if err := event.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(event)
	if err != nil || !json.Valid(encoded) {
		t.Fatalf("json=%s err=%v", encoded, err)
	}
}
