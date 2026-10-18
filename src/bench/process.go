// executes local process harnesses with bounded traces and explicit environments.
package bench

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

const maxCapturedOutput = 16 << 20

type processAdapter struct{ tracePrefix string }

func (a processAdapter) Execute(ctx context.Context, request RunRequest) (HarnessResult, error) {
	config := request.Variant.Process
	arguments, err := renderArguments(config.Arguments, request)
	if err != nil {
		return HarnessResult{}, err
	}
	workingDirectory := request.Workspace.Root
	if config.WorkingDirectory != "" && config.WorkingDirectory != "." {
		workingDirectory, err = containedPath(request.Workspace.Root, config.WorkingDirectory)
		if err != nil {
			return HarnessResult{}, err
		}
		if err := os.MkdirAll(workingDirectory, 0755); err != nil {
			return HarnessResult{}, err
		}
	}
	prefix := a.tracePrefix
	if prefix == "" {
		prefix = "harness"
	}
	spec := processSpec{Executable: config.Executable, Arguments: arguments, Directory: workingDirectory, Environment: explicitEnvironment(config.Environment), TraceDir: request.TraceDir, TracePrefix: prefix}
	if config.InputMode == "stdin" {
		spec.Stdin = request.CaseBrief
		spec.UseStdin = true
	}
	result, err := runProcess(ctx, spec)
	result.Adapter = "process"
	return result, err
}

// processSpec is one fully resolved local process launch shared by every
// adapter that runs a real executable.
type processSpec struct {
	Executable  string
	Arguments   []string
	Directory   string
	Environment []string
	Stdin       string
	UseStdin    bool
	TraceDir    string
	TracePrefix string
}

// runProcess launches spec in its own process group, tees bounded output into
// exclusive trace files, and terminates the whole tree on timeout or
// cancellation.
func runProcess(ctx context.Context, spec processSpec) (HarnessResult, error) {
	if err := os.MkdirAll(spec.TraceDir, 0755); err != nil {
		return HarnessResult{}, err
	}
	stdoutFile, err := os.OpenFile(filepath.Join(spec.TraceDir, spec.TracePrefix+".stdout.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return HarnessResult{}, err
	}
	defer stdoutFile.Close()
	stderrFile, err := os.OpenFile(filepath.Join(spec.TraceDir, spec.TracePrefix+".stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return HarnessResult{}, err
	}
	defer stderrFile.Close()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(spec.Executable, spec.Arguments...)
	cmd.Dir = spec.Directory
	cmd.Env = spec.Environment
	if spec.UseStdin {
		cmd.Stdin = bytes.NewBufferString(spec.Stdin)
	}
	cmd.Stdout = &limitedWriter{Writer: io.MultiWriter(stdoutFile, &stdout), Remaining: maxCapturedOutput}
	cmd.Stderr = &limitedWriter{Writer: io.MultiWriter(stderrFile, &stderr), Remaining: maxCapturedOutput}
	configureProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		return HarnessResult{}, fmt.Errorf("start process: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	result := HarnessResult{Executable: spec.Executable}
	select {
	case waitErr := <-done:
		result.ExitCode = exitCode(waitErr)
	case <-ctx.Done():
		result.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
		result.Cancelled = !result.TimedOut
		terminateProcessTree(cmd)
		waitErr := <-done
		result.ExitCode = exitCode(waitErr)
	}
	result.Stdout, result.Stderr = stdout.String(), stderr.String()
	return result, nil
}

func explicitEnvironment(environment Environment) []string {
	values := map[string]string{}
	for _, name := range environment.Inherit {
		if value, ok := os.LookupEnv(name); ok {
			values[name] = value
		}
	}
	for name, value := range environment.Set {
		values[name] = value
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]string, 0, len(names))
	for _, name := range names {
		result = append(result, name+"="+values[name])
	}
	return result
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}

type limitedWriter struct {
	Writer    io.Writer
	Remaining int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	original := len(p)
	if w.Remaining <= 0 {
		return original, nil
	}
	if int64(len(p)) > w.Remaining {
		p = p[:w.Remaining]
	}
	n, err := w.Writer.Write(p)
	w.Remaining -= int64(n)
	if err != nil {
		return n, err
	}
	return original, nil
}
