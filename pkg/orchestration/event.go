package orchestration

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type EventType string

const (
	EventExecutionChanged EventType = "execution.changed"
	EventRuntimeChanged   EventType = "runtime.changed"
	EventHelperLinked     EventType = "helper.linked"
	EventResultLinked     EventType = "result.linked"
)

type Event struct {
	SchemaVersion int             `json:"schemaVersion"`
	ID            string          `json:"id"`
	ExecutionID   string          `json:"executionId"`
	Type          EventType       `json:"type"`
	Sequence      uint64          `json:"sequence"`
	ObservedAt    time.Time       `json:"observedAt"`
	Payload       json.RawMessage `json:"payload"`
}

func (e Event) Validate() error {
	if e.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version %d", e.SchemaVersion)
	}
	if blank(e.ID) || blank(e.ExecutionID) || e.Sequence == 0 || e.ObservedAt.IsZero() || !json.Valid(e.Payload) {
		return errors.New("event requires IDs, sequence, timestamp and valid JSON payload")
	}
	switch e.Type {
	case EventExecutionChanged, EventRuntimeChanged, EventHelperLinked, EventResultLinked:
		return nil
	default:
		return fmt.Errorf("invalid event type %q", e.Type)
	}
}
