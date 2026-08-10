// tests end-to-end run orchestration, event delivery, and workspace cleanup.
package bench

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunPlanEmitsAndPersistsOrderedLifecycleEvents(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	var observed []RunEvent
	experiment, err := RunPlanWithOptions(context.Background(), plan, RunOptions{Observer: func(event RunEvent) {
		observed = append(observed, event)
	}})
	if err != nil {
		t.Fatal(err)
	}
	expectedTypes := []string{"run.started", "attempt.started", "attempt.completed", "run.progress", "run.completed"}
	if len(observed) != len(expectedTypes) {
		t.Fatalf("expected %d events, got %d: %+v", len(expectedTypes), len(observed), observed)
	}
	for i, expectedType := range expectedTypes {
		if observed[i].Type != expectedType || observed[i].Sequence != i+1 {
			t.Fatalf("event %d = %+v, want type %s and sequence %d", i, observed[i], expectedType, i+1)
		}
		if observed[i].Timestamp.IsZero() || observed[i].ExperimentID != experiment.ID {
			t.Fatalf("event %d lacks run identity or timestamp: %+v", i, observed[i])
		}
	}
	if observed[2].RequiredPassed == nil || !*observed[2].RequiredPassed {
		t.Fatalf("attempt completion lacks grading state: %+v", observed[2])
	}
	if observed[4].Outcome != "evaluation-passed" || observed[4].Report != filepath.Join(experiment.Directory, "report.md") {
		t.Fatalf("run completion lacks final state: %+v", observed[4])
	}

	file, err := os.Open(filepath.Join(experiment.Directory, "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var persisted []RunEvent
	for scanner.Scan() {
		var event RunEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("invalid persisted event: %v", err)
		}
		persisted = append(persisted, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(persisted) != len(observed) {
		t.Fatalf("persisted %d events, observed %d", len(persisted), len(observed))
	}
	for i := range persisted {
		if persisted[i].Sequence != observed[i].Sequence || persisted[i].Type != observed[i].Type {
			t.Fatalf("persisted event %d does not match observer: %+v vs %+v", i, persisted[i], observed[i])
		}
	}
}

func TestRunPlanEndsLifecycleWithFailureEvent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell failure fixture is Unix-only")
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Variants[0].Setup = []Command{{Executable: "sh", Arguments: []string{"-c", "exit 9"}}}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	var observed []RunEvent
	experiment, err := RunPlanWithOptions(context.Background(), plan, RunOptions{Observer: func(event RunEvent) {
		observed = append(observed, event)
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(experiment.Results) != 1 || experiment.Results[0].Failure == nil {
		t.Fatalf("expected failed attempt: %+v", experiment.Results)
	}
	final := observed[len(observed)-1]
	if final.Type != "run.failed" || final.Failure == nil || final.Failure.Phase != "setup" {
		t.Fatalf("unexpected final lifecycle event: %+v", final)
	}
}

func TestRunPlanMockSingleEvaluation(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	fixture := filepath.Join(root, "fixture")
	write(t, filepath.Join(fixture, "original.txt"), "unchanged")
	before, _ := digestPath(fixture)
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Cases[0].Fixture = &SourceRef{Source: fixture}
	m.Variants[0].Mock.Files = map[string]string{"created.txt": "created"}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	experiment, err := RunPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if experiment.Comparison.Outcome != "evaluation-passed" {
		t.Fatalf("unexpected outcome: %+v", experiment.Comparison)
	}
	if len(experiment.Results) != 1 || !experiment.Results[0].RequiredPassed {
		t.Fatalf("unexpected results: %+v", experiment.Results)
	}
	submission := filepath.Join(experiment.Results[0].ArtifactDirectory, "submission")
	if _, err := os.Stat(filepath.Join(submission, inputDirectoryName)); !os.IsNotExist(err) {
		t.Fatal("agent-visible inputs leaked into frozen submission")
	}
	if _, err := os.Stat(filepath.Join(experiment.Directory, "report.md")); err != nil {
		t.Fatal(err)
	}
	after, _ := digestPath(fixture)
	if before != after {
		t.Fatal("source fixture changed")
	}
	if experiment.Results[0].Repository.FilesCreated != 1 {
		t.Fatalf("unexpected repository metrics: %+v", experiment.Results[0].Repository)
	}
	first, err := Regrade(context.Background(), experiment.Directory)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Regrade(context.Background(), experiment.Directory)
	if err != nil {
		t.Fatal(err)
	}
	if first.Path == second.Path || first.Revision != 1 || second.Revision != 2 {
		t.Fatalf("regrading did not create revisions: %+v %+v", first, second)
	}
}

func TestRemoveWorkspaceDeletesReadOnlyStagedInputs(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	write(t, input, "visible")
	workspace, err := prepareWorkspace(Case{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}, Files: []InputAsset{{ID: "input", Source: input}}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := removeWorkspace(workspace.Root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workspace.Root); !os.IsNotExist(err) {
		t.Fatalf("workspace still exists after cleanup: %v", err)
	}
}

func TestProcessAdapterStdinAndNamedInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	write(t, input, "named")
	script := filepath.Join(root, "harness.sh")
	write(t, script, "#!/bin/sh\nread task\nprintf '%s:%s' \"$task\" \"$(cat \"$1\")\"\n")
	if err := os.Chmod(script, 0755); err != nil {
		t.Fatal(err)
	}
	workspace, err := prepareWorkspace(Case{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}, Files: []InputAsset{{ID: "sample", Source: input}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer removeWorkspace(workspace.Root)
	request := RunRequest{Run: PlannedRun{CaseID: "case"}, Variant: Variant{ID: "v", Adapter: "process", Process: &ProcessConfig{Executable: script, Arguments: []string{"{input:sample}"}, InputMode: "stdin"}}, Workspace: workspace, Task: "hello", TraceDir: filepath.Join(root, "traces")}
	result, err := (processAdapter{}).Execute(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 || !strings.Contains(result.Stdout, "hello:named") {
		t.Fatalf("unexpected process result: %+v", result)
	}
}

func TestCompareRequiredGateDominatesScore(t *testing.T) {
	one := 1.0
	zero := 0.0
	manifest := Manifest{Variants: []Variant{{ID: "correct"}, {ID: "wrong"}}, Comparison: Comparison{MinimumRequiredPassRate: 1, TieThreshold: .01}}
	results := []RunResult{{VariantID: "correct", Status: "completed", RequiredPassed: true, OptionalScore: &zero}, {VariantID: "wrong", Status: "completed", RequiredPassed: false, OptionalScore: &one}}
	comparison := Compare(Summarize(results, manifest), true, manifest)
	if comparison.Winner != "correct" {
		t.Fatalf("required gate did not dominate: %+v", comparison)
	}
}

func TestFilteredRunIsInconclusive(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Variants = append(m.Variants, Variant{ID: "second", Adapter: "mock", Mock: &MockConfig{Stdout: "ok"}, IntendedDifferences: []string{"second treatment"}})
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	experiment, err := RunPlanFiltered(context.Background(), plan, "", "mock", 0)
	if err != nil {
		t.Fatal(err)
	}
	if experiment.Complete || experiment.Comparison.Outcome != "inconclusive" {
		t.Fatalf("filtered matrix became conclusive: %+v", experiment.Comparison)
	}
}

func TestFullMockComparisonSelectsEligibleVariant(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Variants = append(m.Variants, Variant{ID: "failing", Adapter: "mock", Mock: &MockConfig{Stdout: "wrong"}, IntendedDifferences: []string{"known failing response"}})
	m.Execution.Repetitions = 2
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	experiment, err := RunPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if experiment.Comparison.Outcome != "winner" || experiment.Comparison.Winner != "mock" {
		t.Fatalf("unexpected comparison: %+v", experiment.Comparison)
	}
}

func TestOptionalCommandGraderProducesWeightedScore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("true command fixture is Unix-only")
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "visible")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Graders = []OptionalGrader{{ID: "quality", Type: "command", Weight: 1, Command: Command{Executable: "true"}}}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	experiment, err := RunPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	score := experiment.Results[0].OptionalScore
	if score == nil || *score != 1 {
		t.Fatalf("unexpected optional score: %v", score)
	}
}

func TestProcessTimeoutTerminatesGroup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix process-group test")
	}
	root := t.TempDir()
	workspace := &Workspace{Root: root, InputRoot: root, TaskFile: filepath.Join(root, "task"), Inputs: map[string]string{}}
	request := RunRequest{Run: PlannedRun{CaseID: "case"}, Variant: Variant{ID: "v", Process: &ProcessConfig{Executable: "sh", Arguments: []string{"-c", "sleep 30 & wait"}, InputMode: "argument"}}, Workspace: workspace, TraceDir: filepath.Join(root, "traces")}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	result, err := (processAdapter{}).Execute(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if !result.TimedOut {
		t.Fatal("expected timeout")
	}
	if time.Since(started) > 3*time.Second {
		t.Fatal("process tree was not terminated promptly")
	}
}
