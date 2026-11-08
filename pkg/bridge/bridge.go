// Package bridge coordinates work and runtime providers without depending on
// the Forma server or user interface.
package bridge

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

var ErrNoWork = errors.New("no eligible work")
var ErrRuntimeNotFound = errors.New("runtime not found")

type ClaimedWork struct {
	Work   orchestration.WorkRef
	Claim  orchestration.ClaimRef
	Prompt string
}

type WorkProvider interface {
	ClaimNext(context.Context, string) (ClaimedWork, error)
	Heartbeat(context.Context, orchestration.WorkRef, orchestration.ClaimRef) (orchestration.ClaimRef, error)
	Release(context.Context, orchestration.WorkRef, orchestration.ClaimRef, string) error
}

type WorkUpdate struct {
	State     orchestration.State
	Message   string
	Progress  int
	ResultURL string
}
type WorkReporter interface {
	Report(context.Context, orchestration.WorkRef, orchestration.ClaimRef, WorkUpdate) error
}
type WorkStatus struct {
	Active bool
	State  orchestration.State
}
type WorkInspector interface {
	Inspect(context.Context, orchestration.WorkRef, orchestration.ClaimRef) (WorkStatus, error)
}

type RuntimeStart struct {
	ExecutionID string
	Prompt      string
	Work        orchestration.WorkRef
}

type RuntimeProvider interface {
	Start(context.Context, RuntimeStart) (orchestration.RuntimeRef, error)
}

type RuntimeObservation struct {
	State   orchestration.State
	Parent  orchestration.RuntimeRef
	Helpers []orchestration.RuntimeRef
}
type RuntimeObserver interface {
	Observe(context.Context, orchestration.RuntimeRef) (RuntimeObservation, error)
}
type RuntimeRecoverer interface {
	Recover(context.Context, string) (orchestration.RuntimeRef, error)
}

type Store interface {
	Create(context.Context, orchestration.Execution) error
	Get(context.Context, string) (orchestration.Execution, error)
	Update(context.Context, orchestration.Execution) error
}

type IDGenerator func() string

type Coordinator struct {
	work    WorkProvider
	runtime RuntimeProvider
	store   Store
	now     func() time.Time
	newID   IDGenerator
}

func New(work WorkProvider, runtime RuntimeProvider, store Store, newID IDGenerator) (*Coordinator, error) {
	if work == nil || runtime == nil || store == nil || newID == nil {
		return nil, errors.New("work provider, runtime provider, store and ID generator are required")
	}
	return &Coordinator{work: work, runtime: runtime, store: store, newID: newID, now: func() time.Time { return time.Now().UTC() }}, nil
}

// StartNext claims one unit of work and starts its parent runtime. If the
// runtime cannot start, the claim is released before the error is returned.
func (c *Coordinator) StartNext(ctx context.Context) (orchestration.Execution, error) {
	id := c.newID()
	claimed, err := c.work.ClaimNext(ctx, id)
	if err != nil {
		return orchestration.Execution{}, err
	}
	now := c.now()
	execution := orchestration.Execution{SchemaVersion: orchestration.SchemaVersion, ID: id, Attempt: 1, State: orchestration.StateClaimed, Work: claimed.Work, Prompt: claimed.Prompt, Claim: &claimed.Claim, Helpers: []orchestration.RuntimeRef{}, Results: []orchestration.ResultLink{}, CreatedAt: now, UpdatedAt: now}
	if err := execution.Validate(); err != nil {
		_ = c.work.Release(ctx, claimed.Work, claimed.Claim, "invalid execution: "+err.Error())
		return orchestration.Execution{}, fmt.Errorf("validate claimed execution: %w", err)
	}
	if err := c.store.Create(ctx, execution); err != nil {
		_ = c.work.Release(ctx, claimed.Work, claimed.Claim, "execution persistence failed")
		return orchestration.Execution{}, fmt.Errorf("persist claimed execution: %w", err)
	}
	parent, err := c.runtime.Start(ctx, RuntimeStart{ExecutionID: id, Prompt: claimed.Prompt, Work: claimed.Work})
	if err != nil {
		releaseErr := c.work.Release(ctx, claimed.Work, claimed.Claim, "runtime start failed")
		execution.State = orchestration.StateFailed
		execution.UpdatedAt = c.now()
		storeErr := c.store.Update(ctx, execution)
		return execution, errors.Join(fmt.Errorf("start runtime: %w", err), releaseErr, storeErr)
	}
	execution.Parent = &parent
	execution.State = orchestration.StateRunning
	execution.UpdatedAt = c.now()
	if err := execution.Validate(); err != nil {
		releaseErr := c.work.Release(ctx, claimed.Work, claimed.Claim, "runtime returned an invalid identity")
		execution.Parent = nil
		execution.State = orchestration.StateFailed
		execution.UpdatedAt = c.now()
		storeErr := c.store.Update(ctx, execution)
		return execution, errors.Join(fmt.Errorf("validate running execution: %w", err), releaseErr, storeErr)
	}
	if err := c.store.Update(ctx, execution); err != nil {
		return execution, fmt.Errorf("persist running execution: %w", err)
	}
	return execution, nil
}

func (c *Coordinator) Recover(ctx context.Context, executionID string) (orchestration.Execution, error) {
	execution, err := c.store.Get(ctx, executionID)
	if err != nil {
		return orchestration.Execution{}, err
	}
	if execution.Terminal() {
		return execution, nil
	}
	if execution.Claim == nil {
		return execution, errors.New("execution has no claim")
	}
	if inspector, ok := c.work.(WorkInspector); ok {
		status, err := inspector.Inspect(ctx, execution.Work, *execution.Claim)
		if err != nil {
			return execution, err
		}
		if !status.Active {
			execution.State = status.State
			execution.UpdatedAt = c.now()
			return execution, c.store.Update(ctx, execution)
		}
	}
	if execution.Parent != nil {
		return c.ReconcileRuntime(ctx, executionID)
	}
	recoverer, ok := c.runtime.(RuntimeRecoverer)
	if !ok {
		return execution, errors.New("runtime recovery is unavailable")
	}
	parent, err := recoverer.Recover(ctx, execution.ID)
	if errors.Is(err, ErrRuntimeNotFound) {
		parent, err = c.runtime.Start(ctx, RuntimeStart{ExecutionID: execution.ID, Prompt: execution.Prompt, Work: execution.Work})
	}
	if err != nil {
		return execution, err
	}
	execution.Parent = &parent
	execution.State = orchestration.StateRunning
	execution.UpdatedAt = c.now()
	if err := execution.Validate(); err != nil {
		return execution, err
	}
	return execution, c.store.Update(ctx, execution)
}

func (c *Coordinator) Heartbeat(ctx context.Context, executionID string) (orchestration.Execution, error) {
	execution, err := c.store.Get(ctx, executionID)
	if err != nil {
		return orchestration.Execution{}, err
	}
	if execution.Terminal() || execution.Claim == nil {
		return execution, errors.New("execution has no active claim")
	}
	claim, err := c.work.Heartbeat(ctx, execution.Work, *execution.Claim)
	if err != nil {
		return execution, err
	}
	execution.Claim = &claim
	execution.UpdatedAt = c.now()
	if err := c.store.Update(ctx, execution); err != nil {
		return execution, err
	}
	return execution, nil
}

func (c *Coordinator) ReconcileRuntime(ctx context.Context, executionID string) (orchestration.Execution, error) {
	execution, err := c.store.Get(ctx, executionID)
	if err != nil {
		return orchestration.Execution{}, err
	}
	observer, ok := c.runtime.(RuntimeObserver)
	if !ok || execution.Parent == nil {
		return execution, errors.New("runtime observation is unavailable")
	}
	observation, err := observer.Observe(ctx, *execution.Parent)
	if err != nil {
		return execution, err
	}
	execution.Parent, execution.Helpers, execution.State, execution.UpdatedAt = &observation.Parent, observation.Helpers, observation.State, c.now()
	if execution.Claim != nil && execution.ReportedState != execution.State {
		if reporter, ok := c.work.(WorkReporter); ok {
			if err := reporter.Report(ctx, execution.Work, *execution.Claim, WorkUpdate{State: execution.State, Message: "Herdr runtime is " + string(execution.State)}); err != nil {
				return execution, err
			}
			execution.ReportedState = execution.State
		}
	}
	if err := execution.Validate(); err != nil {
		return execution, err
	}
	if err := c.store.Update(ctx, execution); err != nil {
		return execution, err
	}
	return execution, nil
}

func (c *Coordinator) ExecuteCommand(ctx context.Context, executionID, action string) (orchestration.Execution, error) {
	execution, err := c.store.Get(ctx, executionID)
	if err != nil {
		return orchestration.Execution{}, err
	}
	if execution.Terminal() || execution.Claim == nil {
		return execution, errors.New("execution has no active claim")
	}
	switch action {
	case "release":
		if err := c.work.Release(ctx, execution.Work, *execution.Claim, "released by operator"); err != nil {
			return execution, err
		}
		execution.State = orchestration.StateReleased
	case "approve":
		reporter, ok := c.work.(WorkReporter)
		if !ok {
			return execution, errors.New("outcome reporting is unavailable")
		}
		if execution.State != orchestration.StateReview {
			return execution, errors.New("only review-ready executions can be approved")
		}
		if err := reporter.Report(ctx, execution.Work, *execution.Claim, WorkUpdate{State: orchestration.StateCompleted, Message: "approved by operator", Progress: 100}); err != nil {
			return execution, err
		}
		execution.State, execution.ReportedState = orchestration.StateCompleted, orchestration.StateCompleted
	default:
		return execution, fmt.Errorf("unsupported operator action %q", action)
	}
	execution.UpdatedAt = c.now()
	return execution, c.store.Update(ctx, execution)
}
