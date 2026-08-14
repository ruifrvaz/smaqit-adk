// re-derives grades, comparisons, and reports from frozen experiment artifacts.
package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RevisionResult struct {
	Revision         int    `json:"revision"`
	Path             string `json:"path"`
	RunCount         int    `json:"runCount"`
	RequiredFailures int    `json:"requiredFailures"`
}

func readExperiment(directory string) (*Experiment, error) {
	data, err := os.ReadFile(filepath.Join(directory, "experiment.json"))
	if err != nil {
		return nil, err
	}
	var experiment Experiment
	if err := json.Unmarshal(data, &experiment); err != nil {
		return nil, err
	}
	return &experiment, nil
}

func Regrade(ctx context.Context, directory string) (RevisionResult, error) {
	experiment, err := readExperiment(directory)
	if err != nil {
		return RevisionResult{}, err
	}
	plan, err := ReadPlan(filepath.Join(directory, "run-plan.json"))
	if err != nil {
		return RevisionResult{}, err
	}
	revision := nextRevision(filepath.Join(directory, "grades"), "revision-", ".json")
	type revisedRun struct {
		RunID          string        `json:"runId"`
		RequiredPassed bool          `json:"requiredPassed"`
		OptionalScore  *float64      `json:"optionalScore"`
		Grades         []GradeResult `json:"grades"`
		Failure        *Failure      `json:"failure,omitempty"`
	}
	payload := struct {
		SchemaVersion int          `json:"schemaVersion"`
		Revision      int          `json:"revision"`
		Runs          []revisedRun `json:"runs"`
	}{SchemaVersion: 1, Revision: revision}
	for _, run := range experiment.Results {
		caseConfig, ok := findCase(plan.Manifest, run.CaseID)
		if !ok {
			return RevisionResult{}, fmt.Errorf("case %s missing from plan", run.CaseID)
		}
		variant, _ := findVariant(plan.Manifest, run.VariantID)
		runDirectory := filepath.Join(directory, "runs", run.RunID)
		submission := filepath.Join(runDirectory, "submission")
		stdout, _ := os.ReadFile(filepath.Join(runDirectory, "traces", "harness.stdout.log"))
		stderr, _ := os.ReadFile(filepath.Join(runDirectory, "traces", "harness.stderr.log"))
		workspace := &Workspace{Root: submission, InputRoot: "<not-exposed>", BriefFile: "<not-exposed>", Inputs: map[string]string{}}
		request := RunRequest{Run: PlannedRun{RunID: run.RunID, CaseID: run.CaseID, VariantID: run.VariantID, Repetition: run.Repetition}, Variant: variant, Workspace: workspace, TraceDir: filepath.Join(runDirectory, "grades", fmt.Sprintf("revision-%03d", revision), "logs")}
		harness := run.Harness
		harness.Stdout = string(stdout)
		harness.Stderr = string(stderr)
		grades, gradeErr := gradeExpectations(ctx, caseConfig.Expect, harness, submission, request)
		revised := revisedRun{RunID: run.RunID, Grades: grades, RequiredPassed: true}
		if gradeErr != nil {
			revised.RequiredPassed = false
			revised.Failure = &Failure{Phase: "regrade", Message: gradeErr.Error()}
		} else {
			for _, grade := range grades {
				if !grade.Passed {
					revised.RequiredPassed = false
				}
			}
		}
		if len(plan.Manifest.Graders) > 0 {
			score := 0.0
			for _, grader := range plan.Manifest.Graders {
				copyRoot, copyErr := os.MkdirTemp("", "smaqit-bench-regrade-")
				if copyErr != nil {
					return RevisionResult{}, copyErr
				}
				copyErr = copyDirectory(submission, copyRoot, nil)
				if copyErr != nil {
					os.RemoveAll(copyRoot)
					return RevisionResult{}, copyErr
				}
				if copyErr = makeWritable(copyRoot); copyErr != nil {
					os.RemoveAll(copyRoot)
					return RevisionResult{}, copyErr
				}
				gradeRequest := request
				gradeRequest.Workspace = &Workspace{Root: copyRoot, InputRoot: "<not-exposed>", BriefFile: "<not-exposed>", Inputs: map[string]string{}}
				graderResult, graderErr := executeCommand(ctx, grader.Command, gradeRequest, "grader-"+grader.ID)
				os.RemoveAll(copyRoot)
				if graderErr != nil {
					return RevisionResult{}, graderErr
				}
				value := 0.0
				if graderResult.ExitCode == 0 {
					value = 1
				}
				score += grader.Weight * value
				revised.Grades = append(revised.Grades, GradeResult{ID: grader.ID, Type: "optional-command", Passed: graderResult.ExitCode == 0, Score: value})
			}
			revised.OptionalScore = &score
		}
		payload.Runs = append(payload.Runs, revised)
	}
	path := filepath.Join(directory, "grades", fmt.Sprintf("revision-%03d.json", revision))
	if err := writeJSONAtomic(path, payload); err != nil {
		return RevisionResult{}, err
	}
	failures := 0
	for _, run := range payload.Runs {
		if !run.RequiredPassed {
			failures++
		}
	}
	return RevisionResult{Revision: revision, Path: path, RunCount: len(payload.Runs), RequiredFailures: failures}, nil
}

func Recompare(directory string) (ComparisonResult, string, error) {
	experiment, err := readExperiment(directory)
	if err != nil {
		return ComparisonResult{}, "", err
	}
	plan, err := ReadPlan(filepath.Join(directory, "run-plan.json"))
	if err != nil {
		return ComparisonResult{}, "", err
	}
	stats := Summarize(experiment.Results, plan.Manifest)
	comparison := Compare(stats, experiment.Complete, plan.Manifest)
	revision := nextRevision(filepath.Join(directory, "comparisons"), "revision-", ".json")
	path := filepath.Join(directory, "comparisons", fmt.Sprintf("revision-%03d.json", revision))
	payload := struct {
		SchemaVersion int                 `json:"schemaVersion"`
		Revision      int                 `json:"revision"`
		Statistics    []VariantStatistics `json:"statistics"`
		Comparison    ComparisonResult    `json:"comparison"`
	}{1, revision, stats, comparison}
	if err := writeJSONAtomic(path, payload); err != nil {
		return ComparisonResult{}, "", err
	}
	return comparison, path, nil
}

func RenderExistingReport(directory, format string) (string, error) {
	experiment, err := readExperiment(directory)
	if err != nil {
		return "", err
	}
	revision := nextRevision(filepath.Join(directory, "reports"), "report-", extensionFor(format))
	switch format {
	case "markdown", "md":
		path := filepath.Join(directory, "reports", fmt.Sprintf("report-%03d.md", revision))
		return path, WriteReport(path, experiment)
	case "json":
		path := filepath.Join(directory, "reports", fmt.Sprintf("report-%03d.json", revision))
		return path, writeJSONAtomic(path, experiment)
	default:
		return "", fmt.Errorf("format must be markdown or json")
	}
}
func extensionFor(format string) string {
	if format == "json" {
		return ".json"
	}
	return ".md"
}
func nextRevision(directory, prefix, suffix string) int {
	entries, _ := os.ReadDir(directory)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), suffix) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return len(names) + 1
}
