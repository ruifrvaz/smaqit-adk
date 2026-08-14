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
	if err := os.MkdirAll(request.TraceDir, 0755); err != nil {
		return HarnessResult{}, err
	}
	prefix := a.tracePrefix
	if prefix == "" {
		prefix = "harness"
	}
	stdoutFile, err := os.OpenFile(filepath.Join(request.TraceDir, prefix+".stdout.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return HarnessResult{}, err
	}
	defer stdoutFile.Close()
	stderrFile, err := os.OpenFile(filepath.Join(request.TraceDir, prefix+".stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return HarnessResult{}, err
	}
	defer stderrFile.Close()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(config.Executable, arguments...)
	cmd.Dir = workingDirectory
	cmd.Env = explicitEnvironment(config.Environment)
	if config.InputMode == "stdin" {
		cmd.Stdin = bytes.NewBufferString(request.CaseBrief)
	}
	cmd.Stdout = &limitedWriter{Writer: io.MultiWriter(stdoutFile, &stdout), Remaining: maxCapturedOutput}
	cmd.Stderr = &limitedWriter{Writer: io.MultiWriter(stderrFile, &stderr), Remaining: maxCapturedOutput}
	configureProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		return HarnessResult{}, fmt.Errorf("start process: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	result := HarnessResult{Adapter: "process", Executable: config.Executable}
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
