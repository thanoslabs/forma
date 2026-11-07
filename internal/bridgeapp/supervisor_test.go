package bridgeapp

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

type fakeCoordinator struct{ starts, hearts int }

func (f *fakeCoordinator) StartNext(context.Context) (orchestration.Execution, error) {
	f.starts++
	return orchestration.Execution{ID: "e", Parent: &orchestration.RuntimeRef{TerminalID: "t"}}, nil
}
func (f *fakeCoordinator) Heartbeat(context.Context, string) (orchestration.Execution, error) {
	f.hearts++
	return orchestration.Execution{}, nil
}

type fakeActiveStore struct{ items []orchestration.Execution }

func (f fakeActiveStore) ListActive(context.Context) ([]orchestration.Execution, error) {
	return f.items, nil
}
func TestSupervisorPollsAndHeartbeats(t *testing.T) {
	c := &fakeCoordinator{}
	s := NewSupervisor(c, fakeActiveStore{items: []orchestration.Execution{{ID: "active"}}}, time.Millisecond, time.Millisecond, 2, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if c.starts == 0 || c.hearts == 0 {
		t.Fatalf("starts=%d hearts=%d", c.starts, c.hearts)
	}
}
