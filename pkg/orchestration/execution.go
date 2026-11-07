// Package orchestration defines the provider-neutral contract shared by the
// standalone bridge, project/runtime adapters, and observing clients.
package orchestration

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const SchemaVersion = 1

type State string

const (
	StatePending   State = "pending"
	StateClaimed   State = "claimed"
	StateRunning   State = "running"
	StateBlocked   State = "blocked"
	StateReview    State = "needs_review"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateReleased  State = "released"
)

type WorkRef struct {
	ConnectionID string `json:"connectionId"`
	ProjectID    string `json:"projectId"`
	ItemID       string `json:"itemId"`
}

type ClaimRef struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

type RuntimeRef struct {
	ConnectionID string `json:"connectionId"`
	MachineID    string `json:"machineId,omitempty"`
	TerminalID   string `json:"terminalId"`
}

type ResultLink struct {
	Kind string `json:"kind"`
	URL  string `json:"url"`
}

type Execution struct {
	SchemaVersion int          `json:"schemaVersion"`
	ID            string       `json:"id"`
	Attempt       int          `json:"attempt"`
	State         State        `json:"state"`
	ReportedState State        `json:"reportedState,omitempty"`
	Work          WorkRef      `json:"work"`
	Claim         *ClaimRef    `json:"claim,omitempty"`
	Parent        *RuntimeRef  `json:"parent,omitempty"`
	Helpers       []RuntimeRef `json:"helpers"`
	Results       []ResultLink `json:"results"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

func (e Execution) Validate() error {
	var validationErrors []error
	if e.SchemaVersion != SchemaVersion {
		validationErrors = append(validationErrors, fmt.Errorf("unsupported schema version %d", e.SchemaVersion))
	}
	if blank(e.ID) {
		validationErrors = append(validationErrors, errors.New("execution ID is required"))
	}
	if e.Attempt < 1 {
		validationErrors = append(validationErrors, errors.New("attempt must be positive"))
	}
	if !validState(e.State) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid execution state %q", e.State))
	}
	if blank(e.Work.ConnectionID) || blank(e.Work.ProjectID) || blank(e.Work.ItemID) {
		validationErrors = append(validationErrors, errors.New("work connection, project and item IDs are required"))
	}
	if e.Claim != nil && blank(e.Claim.ID) {
		validationErrors = append(validationErrors, errors.New("claim ID is required"))
	}
	if e.Parent != nil && (blank(e.Parent.ConnectionID) || blank(e.Parent.TerminalID)) {
		validationErrors = append(validationErrors, errors.New("parent connection and terminal IDs are required"))
	}
	seen := make(map[string]struct{}, len(e.Helpers))
	for _, helper := range e.Helpers {
		if blank(helper.ConnectionID) || blank(helper.TerminalID) {
			validationErrors = append(validationErrors, errors.New("helper connection and terminal IDs are required"))
			continue
		}
		key := helper.ConnectionID + "\x00" + helper.TerminalID
		if _, exists := seen[key]; exists {
			validationErrors = append(validationErrors, fmt.Errorf("duplicate helper %s", helper.TerminalID))
		}
		seen[key] = struct{}{}
	}
	if e.CreatedAt.IsZero() || e.UpdatedAt.IsZero() || e.UpdatedAt.Before(e.CreatedAt) {
		validationErrors = append(validationErrors, errors.New("valid ordered timestamps are required"))
	}
	return errors.Join(validationErrors...)
}

func (e Execution) Terminal() bool {
	return e.State == StateCompleted || e.State == StateFailed || e.State == StateReleased
}

func validState(state State) bool {
	switch state {
	case StatePending, StateClaimed, StateRunning, StateBlocked, StateReview, StateCompleted, StateFailed, StateReleased:
		return true
	default:
		return false
	}
}

func blank(value string) bool { return strings.TrimSpace(value) == "" }
