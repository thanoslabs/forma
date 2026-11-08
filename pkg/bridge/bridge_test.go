package bridge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

type fakeWork struct {
	claimed  ClaimedWork
	claimErr error
	released bool
	hearts   int
}

func (f *fakeWork) ClaimNext(context.Context, string) (ClaimedWork, error) {
	return f.claimed, f.claimErr
}
func (f *fakeWork) Heartbeat(_ context.Context, _ orchestration.WorkRef, claim orchestration.ClaimRef) (orchestration.ClaimRef, error) {
	f.hearts++
	claim.ExpiresAt = claim.ExpiresAt.Add(time.Minute)
	return claim, nil
}
func (f *fakeWork) Release(context.Context, orchestration.WorkRef, orchestration.ClaimRef, string) error {
	f.released = true
	return nil
}

type fakeRuntime struct {
	parent orchestration.RuntimeRef
	err    error
}

func (f fakeRuntime) Start(context.Context, RuntimeStart) (orchestration.RuntimeRef, error) {
	return f.parent, f.err
}

type memoryStore struct {
	executions map[string]orchestration.Execution
}

func (s *memoryStore) Create(_ context.Context, execution orchestration.Execution) error {
	if s.executions == nil {
		s.executions = make(map[string]orchestration.Execution)
	}
	if _, exists := s.executions[execution.ID]; exists {
		return errors.New("duplicate")
	}
	s.executions[execution.ID] = execution
	return nil
}
func (s *memoryStore) Get(_ context.Context, id string) (orchestration.Execution, error) {
	execution, ok := s.executions[id]
	if !ok {
		return orchestration.Execution{}, errors.New("not found")
	}
	return execution, nil
}
func (s *memoryStore) Update(_ context.Context, execution orchestration.Execution) error {
	s.executions[execution.ID] = execution
	return nil
}

func claimedWork() ClaimedWork {
	return ClaimedWork{Work: orchestration.WorkRef{ConnectionID: "overwatch", ProjectID: "repo", ItemID: "issue"}, Claim: orchestration.ClaimRef{ID: "claim", ExpiresAt: time.Now().UTC().Add(time.Minute)}, Prompt: "implement issue"}
}

func TestStartNextPersistsRunningExecution(t *testing.T) {
	work, store := &fakeWork{claimed: claimedWork()}, &memoryStore{}
	coordinator, err := New(work, fakeRuntime{parent: orchestration.RuntimeRef{ConnectionID: "herdr", TerminalID: "terminal-1"}}, store, func() string { return "execution-1" })
	if err != nil {
		t.Fatal(err)
	}
	execution, err := coordinator.StartNext(context.Background())
	if err != nil || execution.State != orchestration.StateRunning || execution.Parent == nil || work.released {
		t.Fatalf("execution=%#v released=%v err=%v", execution, work.released, err)
	}
	updated, err := coordinator.Heartbeat(context.Background(), execution.ID)
	if err != nil || work.hearts != 1 || !updated.Claim.ExpiresAt.After(execution.Claim.ExpiresAt) {
		t.Fatalf("heartbeat=%#v count=%d err=%v", updated, work.hearts, err)
	}
}

func TestStartNextReleasesClaimWhenRuntimeFails(t *testing.T) {
	work, store := &fakeWork{claimed: claimedWork()}, &memoryStore{}
	coordinator, _ := New(work, fakeRuntime{err: errors.New("herdr unavailable")}, store, func() string { return "execution-1" })
	execution, err := coordinator.StartNext(context.Background())
	if err == nil || !work.released || execution.State != orchestration.StateFailed {
		t.Fatalf("execution=%#v released=%v err=%v", execution, work.released, err)
	}
}

func TestStartNextDoesNothingWhenNoWork(t *testing.T) {
	work := &fakeWork{claimErr: ErrNoWork}
	coordinator, _ := New(work, fakeRuntime{}, &memoryStore{}, func() string { return "execution-1" })
	if _, err := coordinator.StartNext(context.Background()); !errors.Is(err, ErrNoWork) {
		t.Fatalf("err=%v", err)
	}
}

func TestOperatorReleaseUpdatesExecutionAndProvider(t *testing.T) {
	work, store := &fakeWork{}, &memoryStore{executions: map[string]orchestration.Execution{}}
	now := time.Now().UTC()
	store.executions["execution-1"] = orchestration.Execution{SchemaVersion: 1, ID: "execution-1", Attempt: 1, State: orchestration.StateRunning, Work: orchestration.WorkRef{ConnectionID: "overwatch", ProjectID: "project", ItemID: "item"}, Claim: &orchestration.ClaimRef{ID: "claim"}, Helpers: []orchestration.RuntimeRef{}, Results: []orchestration.ResultLink{}, CreatedAt: now, UpdatedAt: now}
	coordinator, _ := New(work, fakeRuntime{}, store, func() string { return "unused" })
	execution, err := coordinator.ExecuteCommand(context.Background(), "execution-1", "release")
	if err != nil || execution.State != orchestration.StateReleased || !work.released {
		t.Fatalf("execution=%#v released=%v err=%v", execution, work.released, err)
	}
}
