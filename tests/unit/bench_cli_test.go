package unit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchRunJSONLLifecycleEvents(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "bench.yaml")
	content := `schemaVersion: 1
name: cli-events
cases:
  - id: case
    given:
      prompt: {text: say ok}
    expect:
      - {id: output, type: text, actual: stdout, operator: exact, value: ok}
variants:
  - id: mock
    adapter: mock
    mock: {stdout: ok}
execution: {repetitions: 1, timeoutSeconds: 5}
output: {directory: ./results}
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	output, code := runBinary(t, root, "bench", "run", "-events", "jsonl", manifest)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, output)
	}
	var events []struct {
		Sequence     int    `json:"sequence"`
		Type         string `json:"type"`
		Directory    string `json:"directory"`
		Outcome      string `json:"outcome"`
		Total        int    `json:"totalAttempts"`
		Completed    int    `json:"completedAttempts"`
		RequiredPass *bool  `json:"requiredPassed"`
	}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		var event struct {
			Sequence     int    `json:"sequence"`
			Type         string `json:"type"`
			Directory    string `json:"directory"`
			Outcome      string `json:"outcome"`
			Total        int    `json:"totalAttempts"`
			Completed    int    `json:"completedAttempts"`
			RequiredPass *bool  `json:"requiredPassed"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("invalid JSONL event: %v: %s", err, scanner.Text())
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	expected := []string{"run.started", "attempt.started", "attempt.completed", "run.progress", "run.completed"}
	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %d: %s", len(expected), len(events), output)
	}
	for i := range expected {
		if events[i].Type != expected[i] || events[i].Sequence != i+1 {
			t.Fatalf("event %d = %+v, want %s sequence %d", i, events[i], expected[i], i+1)
		}
	}
	if events[2].RequiredPass == nil || !*events[2].RequiredPass {
		t.Fatalf("attempt event lacks required result: %+v", events[2])
	}
	final := events[len(events)-1]
	if final.Outcome != "evaluation-passed" || final.Directory == "" || final.Completed != 1 || final.Total != 1 {
		t.Fatalf("unexpected final event: %+v", final)
	}
	if _, err := os.Stat(filepath.Join(final.Directory, "events.jsonl")); err != nil {
		t.Fatal(err)
	}
}

func TestBenchRunReportsHumanProgressByDefault(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "bench.yaml")
	content := `schemaVersion: 1
name: cli-human-progress
cases:
  - id: case
    given:
      prompt: {text: say ok}
    expect:
      - {id: output, type: text, actual: stdout, operator: exact, value: ok}
variants:
  - id: mock
    adapter: mock
    mock: {stdout: ok}
execution: {repetitions: 1, timeoutSeconds: 5}
output: {directory: ./results}
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	output, code := runBinary(t, root, "bench", "run", manifest)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, output)
	}
	for _, state := range []string{"bench: run started", "bench: attempt 1/1 started", "bench: progress 1/1", "bench: run completed", "evaluation-passed: cli-human-progress"} {
		if !strings.Contains(output, state) {
			t.Fatalf("missing state %q from output: %s", state, output)
		}
	}
}

func TestBenchRunJSONLReportsConfigurationFailure(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "bad.yaml")
	if err := os.WriteFile(manifest, []byte("schemaVersion: 1\nunknown: true\n"), 0644); err != nil {
		t.Fatal(err)
	}
	output, code := runBinary(t, root, "bench", "run", "-events", "jsonl", manifest)
	if code != 3 {
		t.Fatalf("expected invalid exit 3, got %d: %s", code, output)
	}
	var event struct {
		Sequence int    `json:"sequence"`
		Type     string `json:"type"`
		Failure  struct {
			Phase string `json:"phase"`
		} `json:"failure"`
		Diagnostics []struct {
			Path string `json:"path"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(output), &event); err != nil {
		t.Fatalf("invalid failure event: %v: %s", err, output)
	}
	if event.Sequence != 1 || event.Type != "run.failed" || event.Failure.Phase != "invalid-configuration" {
		t.Fatalf("unexpected failure event: %+v", event)
	}
	if len(event.Diagnostics) == 0 || event.Diagnostics[0].Path != "unknown" {
		t.Fatalf("missing field diagnostic: %+v", event.Diagnostics)
	}
}

func TestBenchNestedHelp(t *testing.T) {
	for _, command := range []string{"validate", "plan", "run", "grade", "compare", "report"} {
		output, code := runBinary(t, t.TempDir(), "bench", command, "--help")
		if code != 0 {
			t.Errorf("%s help exited %d: %s", command, code, output)
		}
		if !strings.Contains(output, "Usage: smaqit-adk bench "+command) {
			t.Errorf("%s help missing usage: %s", command, output)
		}
	}
}

func TestBenchValidateAndRunJSON(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "bench.yaml")
	content := `schemaVersion: 1
name: cli-smoke
cases:
  - id: case
    given:
      prompt: {text: say ok}
    expect:
      - {id: output, type: text, actual: stdout, operator: exact, value: ok}
variants:
  - id: mock
    adapter: mock
    mock: {stdout: ok}
execution: {repetitions: 1, timeoutSeconds: 5}
output: {directory: ./results}
`
	if err := os.WriteFile(manifest, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	output, code := runBinary(t, root, "bench", "validate", "-json", manifest)
	if code != 0 {
		t.Fatalf("validate exited %d: %s", code, output)
	}
	var validation map[string]any
	if err := json.Unmarshal([]byte(output), &validation); err != nil {
		t.Fatalf("invalid validation JSON: %v: %s", err, output)
	}
	if validation["valid"] != true {
		t.Fatalf("unexpected validation: %v", validation)
	}
	output, code = runBinary(t, root, "bench", "run", "-json", manifest)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, output)
	}
	var experiment struct {
		Comparison struct {
			Outcome string `json:"outcome"`
		} `json:"comparison"`
		Directory string `json:"directory"`
	}
	if err := json.Unmarshal([]byte(output), &experiment); err != nil {
		t.Fatalf("invalid run JSON: %v: %s", err, output)
	}
	if experiment.Comparison.Outcome != "evaluation-passed" {
		t.Fatalf("unexpected outcome: %s", experiment.Comparison.Outcome)
	}
	if _, err := os.Stat(filepath.Join(experiment.Directory, "run-plan.json")); err != nil {
		t.Fatal(err)
	}
}

func TestBenchInvalidManifestJSON(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "bad.yaml")
	if err := os.WriteFile(manifest, []byte("schemaVersion: 1\nunknown: true\n"), 0644); err != nil {
		t.Fatal(err)
	}
	output, code := runBinary(t, root, "bench", "validate", "-json", manifest)
	if code != 3 {
		t.Fatalf("expected invalid exit 3, got %d: %s", code, output)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("invalid error JSON: %v", err)
	}
}
