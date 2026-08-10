// defines the harness adapter contract and built-in mock execution.
package bench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RunRequest struct {
	Run       PlannedRun
	Variant   Variant
	Workspace *Workspace
	Task      string
	TraceDir  string
}

type HarnessResult struct {
	Adapter    string `json:"adapter"`
	Executable string `json:"executable,omitempty"`
	ExitCode   int    `json:"exitCode"`
	TimedOut   bool   `json:"timedOut"`
	Cancelled  bool   `json:"cancelled"`
	Stdout     string `json:"-"`
	Stderr     string `json:"-"`
}

type adapter interface {
	Execute(context.Context, RunRequest) (HarnessResult, error)
}

func adapterFor(v Variant) (adapter, error) {
	switch v.Adapter {
	case "mock":
		return mockAdapter{}, nil
	case "process":
		return processAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported adapter %q", v.Adapter)
	}
}

func renderArguments(arguments []string, request RunRequest) ([]string, error) {
	values := map[string]string{
		"{task}": request.Task, "{taskFile}": request.Workspace.TaskFile,
		"{inputRoot}": request.Workspace.InputRoot, "{workspace}": request.Workspace.Root,
		"{caseId}": request.Run.CaseID, "{variantId}": request.Variant.ID,
	}
	for id, path := range request.Workspace.Inputs {
		values["{input:"+id+"}"] = path
	}
	out := make([]string, len(arguments))
	for i, argument := range arguments {
		out[i] = argument
		for placeholder, value := range values {
			out[i] = strings.ReplaceAll(out[i], placeholder, value)
		}
		if unresolved := placeholderPattern.FindString(out[i]); unresolved != "" {
			return nil, fmt.Errorf("unresolved placeholder %s", unresolved)
		}
	}
	return out, nil
}

type mockAdapter struct{}

func (mockAdapter) Execute(_ context.Context, request RunRequest) (HarnessResult, error) {
	config := request.Variant.Mock
	for rel, content := range config.Files {
		target, err := containedPath(request.Workspace.Root, rel)
		if err != nil {
			return HarnessResult{}, err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return HarnessResult{}, err
		}
		if err := os.WriteFile(target, []byte(content), 0644); err != nil {
			return HarnessResult{}, err
		}
	}
	if err := os.MkdirAll(request.TraceDir, 0755); err != nil {
		return HarnessResult{}, err
	}
	if err := os.WriteFile(filepath.Join(request.TraceDir, "harness.stdout.log"), []byte(config.Stdout), 0600); err != nil {
		return HarnessResult{}, err
	}
	if err := os.WriteFile(filepath.Join(request.TraceDir, "harness.stderr.log"), []byte(config.Stderr), 0600); err != nil {
		return HarnessResult{}, err
	}
	return HarnessResult{Adapter: "mock", ExitCode: config.ExitCode, Stdout: config.Stdout, Stderr: config.Stderr}, nil
}

func executeCommand(ctx context.Context, command Command, request RunRequest, tracePrefix string) (HarnessResult, error) {
	variant := request.Variant
	variant.Process = &ProcessConfig{Executable: command.Executable, Arguments: command.Arguments, InputMode: "argument", WorkingDirectory: ".", Environment: command.Environment}
	request.Variant = variant
	if command.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(command.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	return processAdapter{tracePrefix: tracePrefix}.Execute(ctx, request)
}
