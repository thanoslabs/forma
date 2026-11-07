// Package herdr implements the bridge runtime-provider port using herdr-core's
// machine-readable hc CLI. It never invokes a shell or parses terminal text.
package herdr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type commandRunner struct{}

func (commandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

type Config struct {
	Binary       string
	ConnectionID string
	Machine      string
	WorkingDir   string
	Command      []string
	Runner       Runner
}

type Provider struct {
	binary       string
	connectionID string
	machine      string
	workingDir   string
	command      []string
	runner       Runner
}

func New(config Config) (*Provider, error) {
	if config.Binary == "" {
		config.Binary = "hc"
	}
	if strings.TrimSpace(config.ConnectionID) == "" || len(config.Command) == 0 || strings.TrimSpace(config.Command[0]) == "" {
		return nil, errors.New("connection ID and agent command are required")
	}
	if config.Runner == nil {
		config.Runner = commandRunner{}
	}
	return &Provider{binary: config.Binary, connectionID: config.ConnectionID, machine: config.Machine, workingDir: config.WorkingDir, command: append([]string(nil), config.Command...), runner: config.Runner}, nil
}

func (p *Provider) Start(ctx context.Context, request bridge.RuntimeStart) (orchestration.RuntimeRef, error) {
	if strings.TrimSpace(request.ExecutionID) == "" || strings.TrimSpace(request.Prompt) == "" {
		return orchestration.RuntimeRef{}, errors.New("execution ID and prompt are required")
	}
	args := []string{"--json", "run", "--bg", "--name", request.ExecutionID, "--no-parent"}
	if p.machine != "" {
		args = append(args, "--on", p.machine)
	}
	if p.workingDir != "" {
		args = append(args, "--cwd", p.workingDir)
	}
	args = append(args, "--")
	args = append(args, p.command...)
	args = append(args, request.Prompt)
	output, err := p.runner.Run(ctx, p.binary, args...)
	if err != nil {
		return orchestration.RuntimeRef{}, fmt.Errorf("start herdr terminal: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var response struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return orchestration.RuntimeRef{}, fmt.Errorf("decode hc response: %w", err)
	}
	if strings.TrimSpace(response.Address) == "" {
		return orchestration.RuntimeRef{}, errors.New("hc returned no terminal address")
	}
	return orchestration.RuntimeRef{ConnectionID: p.connectionID, MachineID: p.machine, TerminalID: response.Address}, nil
}

type terminal struct {
	Address string `json:"address"`
	Parent  string `json:"parent"`
	State   string `json:"state"`
	Status  string `json:"status"`
}

func (p *Provider) Observe(ctx context.Context, parent orchestration.RuntimeRef) (bridge.RuntimeObservation, error) {
	output, err := p.runner.Run(ctx, p.binary, "--json", "get", parent.TerminalID)
	if err != nil {
		return bridge.RuntimeObservation{}, fmt.Errorf("inspect herdr terminal: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var root terminal
	if err := json.Unmarshal(output, &root); err != nil {
		return bridge.RuntimeObservation{}, fmt.Errorf("decode hc get: %w", err)
	}
	if root.Address == "" {
		root.Address = parent.TerminalID
	}
	list, err := p.runner.Run(ctx, p.binary, "--json", "ls")
	if err != nil {
		return bridge.RuntimeObservation{}, fmt.Errorf("list herdr terminals: %w", err)
	}
	var response struct {
		Terminals []terminal `json:"terminals"`
	}
	if err := json.Unmarshal(list, &response); err != nil {
		return bridge.RuntimeObservation{}, fmt.Errorf("decode hc ls: %w", err)
	}
	helpers := make([]orchestration.RuntimeRef, 0)
	for _, item := range response.Terminals {
		if item.Parent == root.Address {
			helpers = append(helpers, orchestration.RuntimeRef{ConnectionID: p.connectionID, TerminalID: item.Address})
		}
	}
	return bridge.RuntimeObservation{State: mapState(root), Parent: orchestration.RuntimeRef{ConnectionID: p.connectionID, MachineID: p.machine, TerminalID: root.Address}, Helpers: helpers}, nil
}

func mapState(value terminal) orchestration.State {
	if value.State != "live" {
		return orchestration.StateReview
	}
	switch value.Status {
	case "blocked":
		return orchestration.StateBlocked
	case "error":
		return orchestration.StateFailed
	case "done":
		return orchestration.StateReview
	default:
		return orchestration.StateRunning
	}
}
