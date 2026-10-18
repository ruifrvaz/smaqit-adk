// tests freeze handling of non-regular files, guaranteed cleanup reporting,
// and the per-attempt harness home.
package bench

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func requirePOSIX(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell and special files are unavailable on Windows")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh is unavailable")
	}
}

// processManifest returns a one-case, one-variant manifest whose variant runs
// `sh -c <script> sh <arguments...>` with PATH inherited.
func processManifest(script string, arguments ...string) string {
	quoted := []string{`"-c"`, yamlQuote(script), `"sh"`}
	for _, argument := range arguments {
		quoted = append(quoted, yamlQuote(argument))
	}
	return `schemaVersion: 2
name: freeze
cases:
  - id: case-1
    given:
      prompt: {text: hello}
    expect:
      - {id: ran, type: runtime, actual: exitCode, exitCode: 0}
variants:
  - id: shell
    adapter: process
    process:
      executable: sh
      arguments: [` + strings.Join(quoted, ", ") + `]
      inputMode: argument
      environment: {inherit: [PATH]}
execution: {repetitions: 1, timeoutSeconds: 20}
output: {directory: ./results}
`
}

func yamlQuote(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}

func runManifest(t *testing.T, root, body string) *Experiment {
	t.Helper()
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, body)
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	experiment, err := RunPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	return experiment
}

func TestSnapshotTreeRecordsUnixSocketAsSpecial(t *testing.T) {
	requirePOSIX(t)
	// Unix socket paths are length-limited, so keep the directory short.
	root, err := os.MkdirTemp("", "sock")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	write(t, filepath.Join(root, "regular.txt"), "content")
	listener, err := net.Listen("unix", filepath.Join(root, "ipc.sock"))
	if err != nil {
		t.Skipf("unix sockets unavailable: %v", err)
	}
	defer listener.Close()
	tree, err := snapshotTree(root)
	if err != nil {
		t.Fatalf("snapshot must not fail on a socket: %v", err)
	}
	kinds := map[string]TreeEntry{}
	for _, item := range tree {
		kinds[item.Path] = item
	}
	if item := kinds["ipc.sock"]; item.Kind != "special" || item.Type != "socket" || item.SHA256 != "" {
		t.Fatalf("socket entry = %+v", item)
	}
	if item := kinds["regular.txt"]; item.Kind != "file" || item.SHA256 == "" {
		t.Fatalf("regular entry = %+v", item)
	}
	if metrics := compareTrees(nil, tree); metrics.SpecialFiles != 1 || metrics.FinalFiles != 1 {
		t.Fatalf("metrics = %+v", metrics)
	}
}

func TestFreezeExcludesHarnessCreatedFIFOFromSubmission(t *testing.T) {
	requirePOSIX(t)
	if _, err := exec.LookPath("mkfifo"); err != nil {
		t.Skip("mkfifo is unavailable")
	}
	root := t.TempDir()
	experiment := runManifest(t, root, processManifest(`mkfifo "$1/pipe" && echo kept > "$1/kept.txt"`, "{workspace}"))
	result := experiment.Results[0]
	if result.Failure != nil || !result.RequiredPassed {
		t.Fatalf("run with a FIFO must succeed: %+v", result.Failure)
	}
	if result.Repository.SpecialFiles != 1 {
		t.Fatalf("special file not counted: %+v", result.Repository)
	}
	submission := filepath.Join(result.ArtifactDirectory, "submission")
	if _, err := os.Lstat(filepath.Join(submission, "pipe")); !os.IsNotExist(err) {
		t.Fatalf("FIFO must be excluded from the submission: %v", err)
	}
	if _, err := os.Stat(filepath.Join(submission, "kept.txt")); err != nil {
		t.Fatalf("regular output missing from submission: %v", err)
	}
	finalTree, err := os.ReadFile(filepath.Join(result.ArtifactDirectory, "repository", "final-tree.json"))
	if err != nil || !strings.Contains(string(finalTree), `"type": "fifo"`) {
		t.Fatalf("final tree must record the FIFO as special: %v %s", err, finalTree)
	}
}

func TestCleanupFailureIsReportedAlongsideEarlierFailure(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	body := strings.Replace(validManifest("prompt:\n        text: hello"), "    given:", "    prepare:\n      - {executable: false}\n    given:", 1)
	original := cleanupAttempt
	cleanupAttempt = func(w *Workspace) error {
		return errors.Join(removeAttemptWorkspace(w), errors.New("forced cleanup failure"))
	}
	defer func() { cleanupAttempt = original }()
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, body)
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Skipf("false is unavailable: %v", err)
	}
	var cleanupEvent *Failure
	experiment, err := RunPlanWithOptions(context.Background(), plan, RunOptions{Observer: func(event RunEvent) {
		if event.Type == "attempt.completed" {
			cleanupEvent = event.Cleanup
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	result := experiment.Results[0]
	if result.Failure == nil || result.Failure.Phase != "prepare" {
		t.Fatalf("earlier failure must be preserved: %+v", result.Failure)
	}
	if result.Cleanup == nil || !strings.Contains(result.Cleanup.Message, "forced cleanup failure") || result.RequiredPassed {
		t.Fatalf("cleanup failure must be reported despite the earlier failure: %+v", result)
	}
	if cleanupEvent == nil {
		t.Fatal("attempt.completed event must carry the cleanup failure")
	}
	persisted, err := os.ReadFile(filepath.Join(result.ArtifactDirectory, "result.json"))
	if err != nil || !strings.Contains(string(persisted), "forced cleanup failure") {
		t.Fatalf("result.json must record the cleanup failure: %v", err)
	}
	report, err := os.ReadFile(filepath.Join(experiment.Directory, "report.md"))
	if err != nil || !strings.Contains(string(report), "## Cleanup failures") {
		t.Fatalf("report must list cleanup failures: %v\n%s", err, report)
	}
}

func TestHarnessHomeIsOutsideWorkspaceAndRemoved(t *testing.T) {
	requirePOSIX(t)
	root := t.TempDir()
	experiment := runManifest(t, root, processManifest(`echo "$1" > "$2/home-path.txt" && echo secret-state > "$1/state"`, "{harnessHome}", "{workspace}"))
	result := experiment.Results[0]
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %+v", result.Failure)
	}
	data, err := os.ReadFile(filepath.Join(result.ArtifactDirectory, "submission", "home-path.txt"))
	if err != nil {
		t.Fatal(err)
	}
	home := strings.TrimSpace(string(data))
	if !strings.Contains(filepath.Base(home), "smaqit-bench-home-") {
		t.Fatalf("unexpected harness home %q", home)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("harness home must be removed after the attempt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.ArtifactDirectory, "submission", "state")); !os.IsNotExist(err) {
		t.Fatalf("harness-home state must never reach the submission: %v", err)
	}
}

func TestHarnessHomePlaceholderIsLimitedToVariantProcessArguments(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	body := strings.Replace(validManifest("prompt:\n        text: hello"), "      - id: says-ok", "      - id: grade\n        type: command\n        actual: submission\n        command: {executable: sh, arguments: [\"{harnessHome}\"]}\n      - id: says-ok", 1)
	path := filepath.Join(root, "bench.yaml")
	write(t, path, body)
	if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), "unsupported placeholder {harnessHome}") {
		t.Fatalf("expected {harnessHome} to be rejected outside variant process arguments, got %v", err)
	}
}
