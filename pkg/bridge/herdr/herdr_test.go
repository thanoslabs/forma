package herdr

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

type fakeRunner struct {
	name   string
	args   []string
	output []byte
	err    error
}

type sequenceRunner struct {
	outputs [][]byte
	calls   int
}

func (r *sequenceRunner) Run(context.Context, string, ...string) ([]byte, error) {
	out := r.outputs[r.calls]
	r.calls++
	return out, nil
}
func TestProviderObservesParentAndHelpers(t *testing.T) {
	runner := &sequenceRunner{outputs: [][]byte{[]byte(`{"address":"t1","state":"live","status":"blocked"}`), []byte(`{"terminals":[{"address":"t1"},{"address":"t2","parent":"t1"}]}`)}}
	provider, _ := New(Config{ConnectionID: "herdr", Command: []string{"codex"}, Runner: runner})
	observation, err := provider.Observe(context.Background(), orchestration.RuntimeRef{ConnectionID: "herdr", TerminalID: "t1"})
	if err != nil || observation.State != orchestration.StateBlocked || len(observation.Helpers) != 1 || observation.Helpers[0].TerminalID != "t2" {
		t.Fatalf("observation=%#v err=%v", observation, err)
	}
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.name, r.args = name, append([]string(nil), args...)
	return r.output, r.err
}

func TestProviderStartsNamedDetachedTerminal(t *testing.T) {
	runner := &fakeRunner{output: []byte(`{"address":"build-box/t7"}`)}
	provider, err := New(Config{Binary: "/usr/bin/hc", ConnectionID: "herdr-1", Machine: "build-box", WorkingDir: "/workspace/forma", Command: []string{"codex", "--model", "gpt"}, Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	ref, err := provider.Start(context.Background(), bridge.RuntimeStart{ExecutionID: "execution-1", Prompt: "Implement issue"})
	if err != nil || ref.TerminalID != "build-box/t7" || ref.ConnectionID != "herdr-1" {
		t.Fatalf("ref=%#v err=%v", ref, err)
	}
	want := []string{"--json", "run", "--bg", "--name", "execution-1", "--no-parent", "--on", "build-box", "--cwd", "/workspace/forma", "--", "codex", "--model", "gpt", "Implement issue"}
	if runner.name != "/usr/bin/hc" || !reflect.DeepEqual(runner.args, want) {
		t.Fatalf("command=%s args=%q", runner.name, runner.args)
	}
}

func TestProviderRejectsMissingAddressAndPreservesFailure(t *testing.T) {
	provider, _ := New(Config{ConnectionID: "herdr", Command: []string{"codex"}, Runner: &fakeRunner{output: []byte(`{}`)}})
	if _, err := provider.Start(context.Background(), bridge.RuntimeStart{ExecutionID: "execution", Prompt: "work"}); err == nil {
		t.Fatal("expected missing address error")
	}
	provider, _ = New(Config{ConnectionID: "herdr", Command: []string{"codex"}, Runner: &fakeRunner{output: []byte("offline"), err: errors.New("exit 1")}})
	if _, err := provider.Start(context.Background(), bridge.RuntimeStart{ExecutionID: "execution", Prompt: "work"}); err == nil {
		t.Fatal("expected command error")
	}
}
