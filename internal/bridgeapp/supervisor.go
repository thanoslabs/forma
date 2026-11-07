package bridgeapp

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

type Coordinator interface {
	StartNext(context.Context) (orchestration.Execution, error)
	Heartbeat(context.Context, string) (orchestration.Execution, error)
	ReconcileRuntime(context.Context, string) (orchestration.Execution, error)
}
type ActiveStore interface {
	ListActive(context.Context) ([]orchestration.Execution, error)
}

type Supervisor struct {
	coordinator     Coordinator
	store           ActiveStore
	poll, heartbeat time.Duration
	max             int
	log             *slog.Logger
}

func NewSupervisor(c Coordinator, s ActiveStore, poll, heartbeat time.Duration, max int, logger *slog.Logger) *Supervisor {
	return &Supervisor{coordinator: c, store: s, poll: poll, heartbeat: heartbeat, max: max, log: logger}
}
func (s *Supervisor) Run(ctx context.Context) error {
	poll := time.NewTicker(s.poll)
	defer poll.Stop()
	heart := time.NewTicker(s.heartbeat)
	defer heart.Stop()
	if err := s.pollOnce(ctx); err != nil {
		s.log.Error("initial claim poll failed", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-poll.C:
			if err := s.pollOnce(ctx); err != nil {
				s.log.Error("claim poll failed", "error", err)
			}
		case <-heart.C:
			if err := s.heartbeatOnce(ctx); err != nil {
				s.log.Error("heartbeat sweep failed", "error", err)
			}
		}
	}
}
func (s *Supervisor) pollOnce(ctx context.Context) error {
	active, err := s.store.ListActive(ctx)
	if err != nil {
		return err
	}
	if len(active) >= s.max {
		return nil
	}
	execution, err := s.coordinator.StartNext(ctx)
	if errors.Is(err, bridge.ErrNoWork) {
		return nil
	}
	if err == nil {
		s.log.Info("execution started", "execution_id", execution.ID, "terminal", execution.Parent.TerminalID)
	}
	return err
}
func (s *Supervisor) heartbeatOnce(ctx context.Context) error {
	active, err := s.store.ListActive(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, execution := range active {
		if _, err := s.coordinator.ReconcileRuntime(ctx, execution.ID); err != nil {
			errs = append(errs, err)
		}
		if _, err := s.coordinator.Heartbeat(ctx, execution.ID); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
