// Package filestore provides a small atomic durable store for a standalone
// bridge deployment. It has no dependency on the Forma service or database.
package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/thanoslabs/forma/pkg/orchestration"
)

var ErrNotFound = errors.New("execution not found")

type state struct {
	Executions map[string]orchestration.Execution `json:"executions"`
	Events     []orchestration.Event              `json:"events"`
	Sequence   uint64                             `json:"sequence"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	state state
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("store path is required")
	}
	s := &Store{path: path, state: state{Executions: make(map[string]orchestration.Execution), Events: []orchestration.Event{}}}
	encoded, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(encoded, &s.state); err != nil {
		return nil, fmt.Errorf("decode bridge state: %w", err)
	}
	if s.state.Executions == nil {
		s.state.Executions = make(map[string]orchestration.Execution)
	}
	return s, nil
}

func (s *Store) Create(ctx context.Context, execution orchestration.Execution) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := execution.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.state.Executions[execution.ID]; exists {
		return errors.New("execution already exists")
	}
	return s.change(execution)
}

func (s *Store) Get(ctx context.Context, id string) (orchestration.Execution, error) {
	if err := ctx.Err(); err != nil {
		return orchestration.Execution{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	execution, exists := s.state.Executions[id]
	if !exists {
		return orchestration.Execution{}, ErrNotFound
	}
	return execution, nil
}

func (s *Store) Update(ctx context.Context, execution orchestration.Execution) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := execution.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.state.Executions[execution.ID]; !exists {
		return ErrNotFound
	}
	return s.change(execution)
}

func (s *Store) ListActive(ctx context.Context) ([]orchestration.Execution, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]orchestration.Execution, 0)
	for _, execution := range s.state.Executions {
		if !execution.Terminal() {
			result = append(result, execution)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result, nil
}

func (s *Store) EventsAfter(ctx context.Context, sequence uint64) ([]orchestration.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]orchestration.Event, 0)
	for _, event := range s.state.Events {
		if event.Sequence > sequence {
			result = append(result, event)
		}
	}
	return result, nil
}

func (s *Store) change(execution orchestration.Execution) error {
	payload, err := json.Marshal(execution)
	if err != nil {
		return err
	}
	next := cloneState(s.state)
	next.Executions[execution.ID] = execution
	next.Sequence++
	event := orchestration.Event{SchemaVersion: orchestration.SchemaVersion, ID: fmt.Sprintf("%s:%d", execution.ID, next.Sequence), ExecutionID: execution.ID, Type: orchestration.EventExecutionChanged, Sequence: next.Sequence, ObservedAt: time.Now().UTC(), Payload: payload}
	if err := event.Validate(); err != nil {
		return err
	}
	next.Events = append(next.Events, event)
	if err := writeAtomic(s.path, next); err != nil {
		return err
	}
	s.state = next
	return nil
}

func cloneState(source state) state {
	executions := make(map[string]orchestration.Execution, len(source.Executions))
	for id, execution := range source.Executions {
		executions[id] = execution
	}
	return state{Executions: executions, Events: append([]orchestration.Event(nil), source.Events...), Sequence: source.Sequence}
}

func writeAtomic(path string, value state) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".bridge-state-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := json.NewEncoder(temporary).Encode(value); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
