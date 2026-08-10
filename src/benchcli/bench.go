// implements the bench command line, result formats, and lifecycle rendering.
package benchcli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ruifrvaz/smaqit-adk/src/bench"
)

const (
	benchExitSuccess        = 0
	benchExitIneligible     = 2
	benchExitInvalid        = 3
	benchExitDrift          = 4
	benchExitInfrastructure = 5
)

func Run(args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printBenchUsage(os.Stdout)
		return benchExitSuccess
	}
	switch args[0] {
	case "validate":
		return benchValidate(args[1:])
	case "plan":
		return benchPlan(args[1:])
	case "run":
		return benchRun(args[1:])
	case "grade":
		return benchGrade(args[1:])
	case "compare":
		return benchCompare(args[1:])
	case "report":
		return benchReport(args[1:])
	case "suite":
		return benchSuite(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown bench command %q\n", args[0])
		printBenchUsage(os.Stderr)
		return benchExitInvalid
	}
}

func printBenchUsage(w io.Writer) {
	fmt.Fprintln(w, `smaqit-adk bench - local agent evaluation and benchmarking

Usage: smaqit-adk bench <command> [options] <path>

Commands:
  validate <manifest.yaml>       Strictly validate without executing
  plan [-out path] <manifest>    Resolve, hash, and save the run plan
  run <manifest-or-plan>         Verify a plan and execute its run matrix
  grade <experiment-directory>   Regrade frozen submissions without rerunning
  compare <experiment-directory> Recompute comparison from existing evidence
  report <experiment-directory>  Render a report from existing evidence
  suite <validate|plan|run> <directory>
                                  Discover every bench.yaml under a directory
                                  tree and validate, plan, or run each in turn

Every command accepts -json for machine-readable final output. bench run
and bench suite run also accept -events plain|jsonl|quiet for lifecycle state
updates. Exit codes: 0 success, 2 completed but ineligible, 3 invalid
input/configuration, 4 plan drift, and 5 infrastructure failure.`)
}

func commandFlags(name, usage string) (*flag.FlagSet, *bool) {
	fs := flag.NewFlagSet("bench "+name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, usage); fs.PrintDefaults() }
	jsonMode := fs.Bool("json", false, "emit machine-readable JSON")
	return fs, jsonMode
}
func onePath(fs *flag.FlagSet, args []string) (string, error) {
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 1 {
		return "", fmt.Errorf("expected exactly one path")
	}
	return fs.Arg(0), nil
}

func benchValidate(args []string) int {
	fs, jsonMode := commandFlags("validate", "Usage: smaqit-adk bench validate [-json] <manifest.yaml>")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	m, err := bench.LoadManifest(path)
	if err != nil {
		return benchConfigError(err, *jsonMode)
	}
	payload := map[string]any{"valid": true, "name": m.Name, "cases": len(m.Cases), "variants": len(m.Variants)}
	emit(payload, *jsonMode, fmt.Sprintf("valid: %s (%d case(s), %d variant(s))", m.Name, len(m.Cases), len(m.Variants)))
	return 0
}
func benchPlan(args []string) int {
	fs, jsonMode := commandFlags("plan", "Usage: smaqit-adk bench plan [-json] [-out plan.json] <manifest.yaml>")
	out := fs.String("out", "", "saved plan path")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	m, err := bench.LoadManifest(path)
	if err != nil {
		return benchConfigError(err, *jsonMode)
	}
	plan, err := bench.BuildPlan(path, m)
	if err != nil {
		return benchConfigError(err, *jsonMode)
	}
	if *out == "" {
		*out = filepath.Join(m.Output.Directory, safeName(m.Name)+".plan.json")
	}
	if err := bench.WritePlan(*out, plan); err != nil {
		return benchInfrastructureError(err, *jsonMode)
	}
	payload := map[string]any{"planned": true, "planId": plan.PlanID, "path": *out, "seed": plan.Seed, "runs": len(plan.Runs), "warnings": plan.Warnings}
	emit(payload, *jsonMode, fmt.Sprintf("planned %d run(s): %s", len(plan.Runs), *out))
	return 0
}
func benchRun(args []string) int {
	fs, jsonMode := commandFlags("run", "Usage: smaqit-adk bench run [-json] [-events plain|jsonl|quiet] <manifest.yaml|plan.json>")
	caseID := fs.String("case", "", "run only this case ID")
	variantID := fs.String("variant", "", "run only this variant ID")
	repetition := fs.Int("repetition", 0, "run only this repetition (1-based)")
	events := fs.String("events", "auto", "lifecycle updates: plain, jsonl, or quiet")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	eventMode, err := resolveRunEventMode(*events, *jsonMode)
	if err != nil {
		return benchCLIError(err, *jsonMode)
	}
	renderer := &runEventRenderer{mode: eventMode, writer: os.Stdout}
	if *repetition < 0 {
		return benchRunError("invalid-cli", fmt.Errorf("repetition must be positive"), *jsonMode, renderer, benchExitInvalid)
	}
	var plan *bench.Plan
	if looksLikePlan(path) {
		plan, err = bench.ReadPlan(path)
	} else {
		var m *bench.Manifest
		m, err = bench.LoadManifest(path)
		if err == nil {
			plan, err = bench.BuildPlan(path, m)
			if err == nil {
				planPath := filepath.Join(m.Output.Directory, safeName(m.Name)+".plan.json")
				err = bench.WritePlan(planPath, plan)
			}
		}
	}
	if err != nil {
		return benchRunError("invalid-configuration", err, *jsonMode, renderer, benchExitInvalid)
	}
	experiment, err := bench.RunPlanWithOptions(fsContext(), plan, bench.RunOptions{CaseID: *caseID, VariantID: *variantID, Repetition: *repetition, Observer: renderer.observe})
	if err != nil {
		var drift *bench.DriftError
		if errors.As(err, &drift) {
			return benchRunError("drift", err, *jsonMode, renderer, benchExitDrift)
		}
		var selection *bench.SelectionError
		if errors.As(err, &selection) {
			return benchRunError("invalid-cli", err, *jsonMode, renderer, benchExitInvalid)
		}
		return benchRunError("infrastructure", err, *jsonMode, renderer, benchExitInfrastructure)
	}
	if eventMode != "jsonl" {
		emit(experiment, *jsonMode, fmt.Sprintf("%s: %s (%s)", experiment.Comparison.Outcome, experiment.Name, experiment.Directory))
	}
	for _, result := range experiment.Results {
		if result.Failure != nil {
			return benchExitInfrastructure
		}
	}
	if experiment.Comparison.Outcome == "evaluation-failed" || experiment.Comparison.Outcome == "inconclusive" {
		return benchExitIneligible
	}
	return 0
}

func resolveRunEventMode(value string, jsonMode bool) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "auto" {
		if jsonMode {
			return "quiet", nil
		}
		return "plain", nil
	}
	if value != "plain" && value != "jsonl" && value != "quiet" {
		return "", fmt.Errorf("events must be one of plain, jsonl, or quiet")
	}
	if value != "quiet" && jsonMode {
		return "", fmt.Errorf("-json requires -events quiet; use -events jsonl for streaming machine-readable state")
	}
	return value, nil
}

type runEventRenderer struct {
	mode     string
	writer   io.Writer
	sequence int
	lastType string
}

func (r *runEventRenderer) observe(event bench.RunEvent) {
	r.sequence = event.Sequence
	r.lastType = event.Type
	switch r.mode {
	case "jsonl":
		_ = json.NewEncoder(r.writer).Encode(event)
	case "plain":
		switch event.Type {
		case "run.started":
			fmt.Fprintf(r.writer, "bench: run started (%d attempt(s), %s)\n", event.TotalAttempts, event.ExperimentID)
		case "attempt.started":
			fmt.Fprintf(r.writer, "bench: attempt %d/%d started (%s, %s, repetition %d)\n", event.CompletedAttempts+1, event.TotalAttempts, event.CaseID, event.VariantID, event.Repetition)
		case "attempt.completed":
			fmt.Fprintf(r.writer, "bench: attempt %d/%d %s in %dms (required passed: %t)\n", event.CompletedAttempts, event.TotalAttempts, event.Status, event.DurationMS, event.RequiredPassed != nil && *event.RequiredPassed)
		case "run.progress":
			fmt.Fprintf(r.writer, "bench: progress %d/%d attempt(s) complete\n", event.CompletedAttempts, event.TotalAttempts)
		case "run.completed":
			fmt.Fprintf(r.writer, "bench: run completed (%s, report: %s)\n", event.Outcome, event.Report)
		case "run.failed":
			fmt.Fprintf(r.writer, "bench: run failed (%s: %s)\n", event.Failure.Phase, event.Failure.Message)
		}
	}
}

func (r *runEventRenderer) failure(kind string, err error) {
	diagnostics := []bench.Diagnostic(nil)
	var validation *bench.ValidationError
	if errors.As(err, &validation) {
		diagnostics = validation.Diagnostics
	}
	r.observe(bench.RunEvent{SchemaVersion: 1, Sequence: r.sequence + 1, Type: "run.failed", Timestamp: time.Now().UTC(), Failure: &bench.Failure{Phase: kind, Message: err.Error()}, Diagnostics: diagnostics})
}

func benchRunError(kind string, err error, jsonMode bool, renderer *runEventRenderer, exitCode int) int {
	if renderer.mode == "jsonl" {
		if renderer.lastType != "run.failed" {
			renderer.failure(kind, err)
		}
		return exitCode
	}
	emitError(kind, err, jsonMode)
	return exitCode
}
func benchGrade(args []string) int {
	fs, jsonMode := commandFlags("grade", "Usage: smaqit-adk bench grade [-json] <experiment-directory>")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	revision, err := bench.Regrade(fsContext(), path)
	if err != nil {
		return benchInfrastructureError(err, *jsonMode)
	}
	emit(revision, *jsonMode, fmt.Sprintf("wrote grading revision: %s", revision.Path))
	if revision.RequiredFailures > 0 {
		return benchExitIneligible
	}
	return 0
}
func benchCompare(args []string) int {
	fs, jsonMode := commandFlags("compare", "Usage: smaqit-adk bench compare [-json] <experiment-directory>")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	result, revision, err := bench.Recompare(path)
	if err != nil {
		return benchInfrastructureError(err, *jsonMode)
	}
	emit(result, *jsonMode, fmt.Sprintf("%s: %s (revision %s)", result.Outcome, result.Reason, revision))
	if result.Outcome == "evaluation-failed" || result.Outcome == "inconclusive" {
		return benchExitIneligible
	}
	return 0
}
func benchReport(args []string) int {
	fs, jsonMode := commandFlags("report", "Usage: smaqit-adk bench report [-json] [-format markdown|json] <experiment-directory>")
	format := fs.String("format", "markdown", "report format")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	output, err := bench.RenderExistingReport(path, *format)
	if err != nil {
		return benchInfrastructureError(err, *jsonMode)
	}
	emit(map[string]any{"reported": true, "path": output, "format": *format}, *jsonMode, "report written: "+output)
	return 0
}

func benchSuite(args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(os.Stderr, "Usage: smaqit-adk bench suite <validate|plan|run> [options] <directory>")
		return benchExitInvalid
	}
	switch args[0] {
	case "validate":
		return benchSuiteValidate(args[1:])
	case "plan":
		return benchSuitePlan(args[1:])
	case "run":
		return benchSuiteRun(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown bench suite command %q\n", args[0])
		return benchExitInvalid
	}
}
func benchSuiteValidate(args []string) int {
	fs, jsonMode := commandFlags("suite validate", "Usage: smaqit-adk bench suite validate [-json] <directory>")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	result, err := bench.ValidateSuite(path)
	if err != nil {
		return benchConfigError(err, *jsonMode)
	}
	emit(result, *jsonMode, fmt.Sprintf("suite validate: %d manifest(s), valid=%t (%s)", len(result.Manifests), result.Valid, path))
	if !result.Valid {
		return benchExitInvalid
	}
	return 0
}
func benchSuitePlan(args []string) int {
	fs, jsonMode := commandFlags("suite plan", "Usage: smaqit-adk bench suite plan [-json] <directory>")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	result, err := bench.PlanSuite(path)
	if err != nil {
		return benchConfigError(err, *jsonMode)
	}
	emit(result, *jsonMode, fmt.Sprintf("suite plan: %d manifest(s), planned=%t (%s)", len(result.Manifests), result.Planned, path))
	if !result.Planned {
		return benchExitInvalid
	}
	return 0
}
func benchSuiteRun(args []string) int {
	fs, jsonMode := commandFlags("suite run", "Usage: smaqit-adk bench suite run [-json] [-events plain|jsonl|quiet] <directory>")
	events := fs.String("events", "auto", "lifecycle updates: plain, jsonl, or quiet")
	path, err := onePath(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return benchExitSuccess
		}
		return benchCLIError(err, *jsonMode)
	}
	eventMode, err := resolveRunEventMode(*events, *jsonMode)
	if err != nil {
		return benchCLIError(err, *jsonMode)
	}
	renderer := &suiteEventRenderer{mode: eventMode, writer: os.Stdout}
	result, err := bench.RunSuite(fsContext(), path, bench.SuiteOptions{Observer: renderer.observe})
	if err != nil {
		return benchInfrastructureError(err, *jsonMode)
	}
	if eventMode != "jsonl" {
		emit(result, *jsonMode, fmt.Sprintf("suite run: %d passed, %d failed, %d errored (%s)", result.Passed, result.Failed, result.Errored, path))
	}
	if result.Errored > 0 {
		return benchExitInfrastructure
	}
	if !result.Eligible() {
		return benchExitIneligible
	}
	return 0
}

type suiteEventRenderer struct {
	mode   string
	writer io.Writer
}

func (r *suiteEventRenderer) observe(manifestPath string, event bench.RunEvent) {
	label := filepath.Base(filepath.Dir(manifestPath))
	switch r.mode {
	case "jsonl":
		_ = json.NewEncoder(r.writer).Encode(map[string]any{"manifestPath": manifestPath, "event": event})
	case "plain":
		switch event.Type {
		case "run.started":
			fmt.Fprintf(r.writer, "bench[%s]: run started (%d attempt(s))\n", label, event.TotalAttempts)
		case "attempt.completed":
			fmt.Fprintf(r.writer, "bench[%s]: attempt %d/%d %s (required passed: %t)\n", label, event.CompletedAttempts, event.TotalAttempts, event.Status, event.RequiredPassed != nil && *event.RequiredPassed)
		case "run.completed":
			fmt.Fprintf(r.writer, "bench[%s]: run completed (%s)\n", label, event.Outcome)
		case "run.failed":
			fmt.Fprintf(r.writer, "bench[%s]: run failed (%s: %s)\n", label, event.Failure.Phase, event.Failure.Message)
		}
	}
}

func fsContext() context.Context { return context.Background() }
func looksLikePlan(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return strings.EqualFold(filepath.Ext(path), ".json")
	}
	var header struct {
		PlanID string `json:"planId"`
	}
	return json.Unmarshal(data, &header) == nil && header.PlanID != ""
}
func safeName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "bench"
	}
	return strings.Trim(b.String(), "-")
}
func benchCLIError(err error, jsonMode bool) int {
	emitError("invalid-cli", err, jsonMode)
	return benchExitInvalid
}
func benchConfigError(err error, jsonMode bool) int {
	emitError("invalid-configuration", err, jsonMode)
	return benchExitInvalid
}
func benchInfrastructureError(err error, jsonMode bool) int {
	emitError("infrastructure", err, jsonMode)
	return benchExitInfrastructure
}
func emitError(kind string, err error, jsonMode bool) {
	payload := map[string]any{"ok": false, "error": map[string]any{"kind": kind, "message": err.Error()}}
	var validation *bench.ValidationError
	if errors.As(err, &validation) {
		payload["diagnostics"] = validation.Diagnostics
	}
	if jsonMode {
		_ = json.NewEncoder(os.Stdout).Encode(payload)
	} else {
		fmt.Fprintf(os.Stderr, "%s: %v\n", kind, err)
	}
}
func emit(value any, jsonMode bool, human string) {
	if jsonMode {
		_ = json.NewEncoder(os.Stdout).Encode(value)
	} else {
		fmt.Fprintln(os.Stdout, human)
	}
}
