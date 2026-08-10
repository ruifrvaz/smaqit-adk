// orchestrates planned runs, workspace cleanup, evidence capture, and grading.
package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NullableMetrics struct {
	InputTokens   *int64   `json:"inputTokens"`
	OutputTokens  *int64   `json:"outputTokens"`
	TotalTokens   *int64   `json:"totalTokens"`
	ToolCalls     *int64   `json:"toolCalls"`
	EstimatedCost *float64 `json:"estimatedCost"`
}
type RunResult struct {
	SchemaVersion     int               `json:"schemaVersion"`
	ExperimentID      string            `json:"experimentId"`
	RunID             string            `json:"runId"`
	CaseID            string            `json:"caseId"`
	VariantID         string            `json:"variantId"`
	Repetition        int               `json:"repetition"`
	Status            string            `json:"status"`
	StartedAt         time.Time         `json:"startedAt"`
	CompletedAt       time.Time         `json:"completedAt"`
	DurationMS        int64             `json:"durationMs"`
	Harness           HarnessResult     `json:"harness"`
	Usage             NullableMetrics   `json:"usage"`
	Repository        RepositoryMetrics `json:"repository"`
	RequiredPassed    bool              `json:"requiredPassed"`
	OptionalScore     *float64          `json:"optionalScore"`
	Grades            []GradeResult     `json:"grades"`
	Failure           *Failure          `json:"failure"`
	ArtifactDirectory string            `json:"artifactDirectory"`
}
type Failure struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
}
type Experiment struct {
	SchemaVersion int                 `json:"schemaVersion"`
	ID            string              `json:"id"`
	PlanID        string              `json:"planId"`
	Name          string              `json:"name"`
	CreatedAt     time.Time           `json:"createdAt"`
	Complete      bool                `json:"complete"`
	Results       []RunResult         `json:"results"`
	Statistics    []VariantStatistics `json:"statistics"`
	Comparison    ComparisonResult    `json:"comparison"`
	Warnings      []string            `json:"warnings,omitempty"`
	Directory     string              `json:"directory"`
}

func RunPlan(ctx context.Context, p *Plan) (*Experiment, error) {
	return RunPlanWithOptions(ctx, p, RunOptions{})
}

func RunPlanFiltered(ctx context.Context, p *Plan, caseID, variantID string, repetition int) (*Experiment, error) {
	return RunPlanWithOptions(ctx, p, RunOptions{CaseID: caseID, VariantID: variantID, Repetition: repetition})
}

// RunPlanWithOptions executes a full or filtered plan and reports ordered
// lifecycle state transitions to the optional observer.
func RunPlanWithOptions(ctx context.Context, p *Plan, options RunOptions) (*Experiment, error) {
	if err := VerifyPlan(p); err != nil {
		return nil, &DriftError{err}
	}
	var selected []PlannedRun
	for _, run := range p.Runs {
		if options.CaseID != "" && run.CaseID != options.CaseID {
			continue
		}
		if options.VariantID != "" && run.VariantID != options.VariantID {
			continue
		}
		if options.Repetition > 0 && run.Repetition != options.Repetition {
			continue
		}
		selected = append(selected, run)
	}
	if len(selected) == 0 {
		return nil, &SelectionError{"run filters selected no attempts"}
	}
	return runPlanRuns(ctx, p, selected, len(selected) == len(p.Runs), options.Observer)
}

type SelectionError struct{ Message string }

func (e *SelectionError) Error() string { return e.Message }

func runPlanRuns(ctx context.Context, p *Plan, runs []PlannedRun, completeMatrix bool, observer func(RunEvent)) (completedExperiment *Experiment, runErr error) {
	id := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + p.PlanID[:8]
	directory := filepath.Join(p.Manifest.Output.Directory, "experiment-"+id)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}
	experiment := &Experiment{SchemaVersion: 1, ID: id, PlanID: p.PlanID, Name: p.Name, CreatedAt: time.Now().UTC(), Complete: true, Warnings: p.Warnings, Directory: directory}
	journal, err := newEventJournal(directory, observer)
	if err != nil {
		return nil, err
	}
	defer func() {
		if runErr != nil {
			_ = journal.emit(RunEvent{Type: "run.failed", ExperimentID: id, PlanID: p.PlanID, Name: p.Name, Directory: directory, Failure: &Failure{Phase: "run", Message: runErr.Error()}})
		}
		_ = journal.close()
	}()
	if err := journal.emit(RunEvent{Type: "run.started", ExperimentID: id, PlanID: p.PlanID, Name: p.Name, Directory: directory, TotalAttempts: len(runs)}); err != nil {
		return nil, err
	}
	if err := WritePlan(filepath.Join(directory, "run-plan.json"), p); err != nil {
		return nil, err
	}
	if err := writeJSONAtomic(filepath.Join(directory, "resolved-manifest.json"), sanitizedManifest(p.Manifest)); err != nil {
		return nil, err
	}
	for index, planned := range runs {
		if err := journal.emit(RunEvent{Type: "attempt.started", ExperimentID: id, PlanID: p.PlanID, RunID: planned.RunID, CaseID: planned.CaseID, VariantID: planned.VariantID, Repetition: planned.Repetition, CompletedAttempts: index, TotalAttempts: len(runs)}); err != nil {
			return nil, err
		}
		result := runAttempt(ctx, experiment.ID, directory, p, planned)
		experiment.Results = append(experiment.Results, result)
		passed := result.RequiredPassed
		if err := journal.emit(RunEvent{Type: "attempt.completed", ExperimentID: id, PlanID: p.PlanID, RunID: planned.RunID, CaseID: planned.CaseID, VariantID: planned.VariantID, Repetition: planned.Repetition, Status: result.Status, RequiredPassed: &passed, DurationMS: result.DurationMS, CompletedAttempts: index + 1, TotalAttempts: len(runs), Failure: result.Failure}); err != nil {
			return nil, err
		}
		if err := journal.emit(RunEvent{Type: "run.progress", ExperimentID: id, PlanID: p.PlanID, CompletedAttempts: index + 1, TotalAttempts: len(runs)}); err != nil {
			return nil, err
		}
		if result.Failure != nil && p.Manifest.Execution.FailFast {
			experiment.Complete = false
			break
		}
	}
	experiment.Statistics = Summarize(experiment.Results, p.Manifest)
	experiment.Complete = completeMatrix && len(experiment.Results) == len(runs)
	experiment.Comparison = Compare(experiment.Statistics, experiment.Complete, p.Manifest)
	if err := writeJSONAtomic(filepath.Join(directory, "experiment.json"), experiment); err != nil {
		return nil, err
	}
	if err := writeJSONAtomic(filepath.Join(directory, "comparison.json"), experiment.Comparison); err != nil {
		return nil, err
	}
	if err := WriteReport(filepath.Join(directory, "report.md"), experiment); err != nil {
		return nil, err
	}
	complete := experiment.Complete
	finalEvent := RunEvent{Type: "run.completed", ExperimentID: id, PlanID: p.PlanID, Name: p.Name, Directory: directory, Report: filepath.Join(directory, "report.md"), CompletedAttempts: len(experiment.Results), TotalAttempts: len(runs), Complete: &complete, Outcome: experiment.Comparison.Outcome, Winner: experiment.Comparison.Winner}
	for _, result := range experiment.Results {
		if result.Failure != nil {
			finalEvent.Type = "run.failed"
			finalEvent.Failure = result.Failure
			break
		}
	}
	if err := journal.emit(finalEvent); err != nil {
		return nil, err
	}
	return experiment, nil
}

func runAttempt(parent context.Context, experimentID, experimentDirectory string, p *Plan, planned PlannedRun) (result RunResult) {
	started := time.Now().UTC()
	runDirectory := filepath.Join(experimentDirectory, "runs", planned.RunID)
	traceDirectory := filepath.Join(runDirectory, "traces")
	result = RunResult{SchemaVersion: 1, ExperimentID: experimentID, RunID: planned.RunID, CaseID: planned.CaseID, VariantID: planned.VariantID, Repetition: planned.Repetition, StartedAt: started, Status: "failed", ArtifactDirectory: runDirectory}
	caseConfig, _ := findCase(p.Manifest, planned.CaseID)
	variant, _ := findVariant(p.Manifest, planned.VariantID)
	workspace, err := prepareWorkspace(caseConfig)
	if err != nil {
		return finishFailure(result, "workspace", err)
	}
	defer func() {
		if cleanupErr := removeWorkspace(workspace.Root); cleanupErr != nil && result.Failure == nil {
			result = finishFailure(result, "cleanup", cleanupErr)
		}
	}()
	taskBytes, err := os.ReadFile(workspace.TaskFile)
	if err != nil {
		return finishFailure(result, "workspace", err)
	}
	task := string(taskBytes)
	request := RunRequest{Run: planned, Variant: variant, Workspace: workspace, Task: task, TraceDir: traceDirectory}
	if err := os.MkdirAll(runDirectory, 0755); err != nil {
		return finishFailure(result, "artifacts", err)
	}
	requestArtifact := map[string]any{"schemaVersion": 1, "runId": planned.RunID, "caseId": planned.CaseID, "variantId": planned.VariantID, "workspace": "<ephemeral>", "task": task, "inputIds": inputIDs(workspace.Inputs), "environmentNames": environmentNames(variant)}
	if variant.Process != nil {
		resolvedArguments, renderErr := renderArguments(variant.Process.Arguments, request)
		if renderErr != nil {
			return finishFailure(result, "request", renderErr)
		}
		requestArtifact["executable"] = variant.Process.Executable
		requestArtifact["arguments"] = resolvedArguments
		requestArtifact["inputMode"] = variant.Process.InputMode
	}
	if err := writeJSONAtomic(filepath.Join(runDirectory, "request.json"), requestArtifact); err != nil {
		return finishFailure(result, "artifacts", err)
	}
	baseline, err := snapshotTree(workspace.Root)
	if err != nil {
		return finishFailure(result, "workspace", err)
	}
	if err := writeJSONAtomic(filepath.Join(runDirectory, "repository", "baseline-tree.json"), baseline); err != nil {
		return finishFailure(result, "artifacts", err)
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(p.Manifest.Execution.TimeoutSeconds)*time.Second)
	defer cancel()
	for i, command := range variant.Setup {
		setupResult, err := executeCommand(ctx, command, request, fmt.Sprintf("setup-%02d", i+1))
		if err != nil {
			return finishFailure(result, "setup", err)
		}
		if setupResult.ExitCode != 0 {
			return finishFailure(result, "setup", fmt.Errorf("command exited %d", setupResult.ExitCode))
		}
	}
	selectedAdapter, err := adapterFor(variant)
	if err != nil {
		return finishFailure(result, "adapter", err)
	}
	harness, err := selectedAdapter.Execute(ctx, request)
	result.Harness = harness
	if err != nil {
		return finishFailure(result, "adapter", err)
	}
	if harness.TimedOut {
		result.Status = "timedOut"
	} else if harness.Cancelled {
		result.Status = "cancelled"
	} else {
		result.Status = "completed"
	}
	finalTree, err := snapshotTree(workspace.Root)
	if err != nil {
		return finishFailure(result, "freeze", err)
	}
	result.Repository = compareTrees(baseline, finalTree)
	if err := writeJSONAtomic(filepath.Join(runDirectory, "repository", "final-tree.json"), finalTree); err != nil {
		return finishFailure(result, "artifacts", err)
	}
	if err := writeJSONAtomic(filepath.Join(runDirectory, "repository", "metrics.json"), result.Repository); err != nil {
		return finishFailure(result, "artifacts", err)
	}
	submission := filepath.Join(runDirectory, "submission")
	stagedSubmission, err := os.MkdirTemp(runDirectory, ".submission-")
	if err != nil {
		return finishFailure(result, "freeze", err)
	}
	defer os.RemoveAll(stagedSubmission)
	if err := copyDirectory(workspace.Root, stagedSubmission, func(rel string) bool {
		return rel == inputDirectoryName || strings.HasPrefix(rel, inputDirectoryName+"/")
	}); err != nil {
		return finishFailure(result, "freeze", err)
	}
	if err := makeEvidenceReadOnly(stagedSubmission); err != nil {
		return finishFailure(result, "freeze", err)
	}
	if err := os.Rename(stagedSubmission, submission); err != nil {
		return finishFailure(result, "freeze", err)
	}
	gradingRequest := request
	gradingRequest.TraceDir = filepath.Join(runDirectory, "grades", "revision-001", "logs")
	grades, err := gradeExpectations(ctx, caseConfig.Expect, harness, submission, gradingRequest)
	if err != nil {
		return finishFailure(result, "grade", err)
	}
	result.Grades = grades
	result.RequiredPassed = !harness.TimedOut && !harness.Cancelled
	for _, grade := range grades {
		if !grade.Passed {
			result.RequiredPassed = false
		}
	}
	if len(p.Manifest.Graders) > 0 {
		score := 0.0
		for _, grader := range p.Manifest.Graders {
			copyRoot, err := os.MkdirTemp("", "smaqit-bench-grader-")
			if err != nil {
				return finishFailure(result, "grader", err)
			}
			if err := copyDirectory(submission, copyRoot, nil); err != nil {
				os.RemoveAll(copyRoot)
				return finishFailure(result, "grader", err)
			}
			if err := makeWritable(copyRoot); err != nil {
				os.RemoveAll(copyRoot)
				return finishFailure(result, "grader", err)
			}
			gradeRequest := gradingRequest
			gradeRequest.Workspace = &Workspace{Root: copyRoot, InputRoot: workspace.InputRoot, TaskFile: workspace.TaskFile, Inputs: workspace.Inputs}
			gr, err := executeCommand(ctx, grader.Command, gradeRequest, "grader-"+grader.ID)
			os.RemoveAll(copyRoot)
			if err != nil {
				return finishFailure(result, "grader", err)
			}
			value := 0.0
			if gr.ExitCode == 0 {
				value = 1
			}
			score += grader.Weight * value
			result.Grades = append(result.Grades, GradeResult{ID: grader.ID, Type: "optional-command", Passed: gr.ExitCode == 0, Score: value})
		}
		result.OptionalScore = &score
	}
	result.CompletedAt = time.Now().UTC()
	result.DurationMS = result.CompletedAt.Sub(started).Milliseconds()
	_ = writeJSONAtomic(filepath.Join(runDirectory, "grades", "revision-001.json"), result.Grades)
	_ = writeJSONAtomic(filepath.Join(runDirectory, "result.json"), result)
	return result
}

func finishFailure(result RunResult, phase string, err error) RunResult {
	result.Failure = &Failure{Phase: phase, Message: err.Error()}
	result.CompletedAt = time.Now().UTC()
	result.DurationMS = result.CompletedAt.Sub(result.StartedAt).Milliseconds()
	if result.ArtifactDirectory != "" {
		_ = writeJSONAtomic(filepath.Join(result.ArtifactDirectory, "result.json"), result)
	}
	return result
}
func findCase(m Manifest, id string) (Case, bool) {
	for _, c := range m.Cases {
		if c.ID == id {
			return c, true
		}
	}
	return Case{}, false
}
func findVariant(m Manifest, id string) (Variant, bool) {
	for _, v := range m.Variants {
		if v.ID == id {
			return v, true
		}
	}
	return Variant{}, false
}
func inputIDs(inputs map[string]string) []string {
	ids := make([]string, 0, len(inputs))
	for id := range inputs {
		ids = append(ids, id)
	}
	sortStrings(ids)
	return ids
}
func environmentNames(v Variant) []string {
	if v.Process == nil {
		return nil
	}
	names := append([]string{}, v.Process.Environment.Inherit...)
	for name := range v.Process.Environment.Set {
		names = append(names, name)
	}
	sortStrings(names)
	return names
}
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
func sanitizedManifest(m Manifest) Manifest {
	data, _ := json.Marshal(m)
	var clone Manifest
	_ = json.Unmarshal(data, &clone)
	for i := range clone.Variants {
		if clone.Variants[i].Process != nil {
			for name := range clone.Variants[i].Process.Environment.Set {
				clone.Variants[i].Process.Environment.Set[name] = "<redacted>"
			}
		}
	}
	return clone
}

type DriftError struct{ error }

func (e *DriftError) Unwrap() error { return e.error }
