// Eval runner for smaqit-adk behavioral evaluations.
// Drives Copilot SDK sessions against ADK artifacts (skills and agents),
// then grades the resulting transcript against per-eval criteria.
//
// Auth: requires COPILOT_GITHUB_TOKEN, GH_TOKEN, or GITHUB_TOKEN.
// Without an explicit token the Copilot CLI reads auth from shared XDG config
// dirs, routing sessions through VS Code and loading smaqit-adk's workspace
// context — which invalidates results. Use 'make evals' to auto-detect a token
// via gh auth token.
//
// Usage:
//
// go run ./evals/runner/... <evals-dir>
//
// Output: tests/evals/runs/<YYYYMMDD-HHMMSS>/
//
// report.md    — human-readable run report
// results.json — machine-readable results (CI-friendly)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// ── Eval file format ──────────────────────────────────────────────────────────

type Turn struct {
	UserInput string `json:"user_input"`
	// Trigger, when set to "ask_user", indicates this answer is fed via
	// OnUserInputRequest rather than a direct session.SendAndWait call.
	Trigger string `json:"trigger,omitempty"`
}

type EvalFile struct {
	Type              string   `json:"type"`          // "skill" or "agent"
	ArtifactFile      string   `json:"artifact_file"` // path relative to repo root
	Description       string   `json:"description"`
	Turns             []Turn   `json:"turns"`
	ExpectedBehavior  []string `json:"expected_behavior"`
	ForbiddenBehavior []string `json:"forbidden_behavior"`
}

// ── Result types ──────────────────────────────────────────────────────────────

type CriterionResult struct {
	Criterion string `json:"criterion"`
	Pass      bool   `json:"pass"`
	Reason    string `json:"reason"`
	Forbidden bool   `json:"forbidden"` // true → criterion must NOT be present
}

type EvalResult struct {
	EvalFile    string            `json:"file"`
	Description string            `json:"description"`
	Pass        bool              `json:"pass"`
	Criteria    []CriterionResult `json:"criteria"`
	Error       string            `json:"error,omitempty"`
}

// JSONReport is the top-level results.json structure.
type JSONReport struct {
	RunID     string       `json:"run_id"`
	Timestamp string       `json:"timestamp"`
	Total     int          `json:"total"`
	Passed    int          `json:"passed"`
	Failed    int          `json:"failed"`
	Evals     []EvalResult `json:"evals"`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: eval-runner <evals-dir>")
		os.Exit(1)
	}

	evalsDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving evals dir: %v\n", err)
		os.Exit(1)
	}

	// repo root is two levels up from tests/evals/
	repoRoot := filepath.Join(evalsDir, "..", "..")

	// Run directory: <evals>/runs/<YYYYMMDD-HHMMSS>
	now := time.Now()
	runID := now.Format("20060102-150405")
	runDir := filepath.Join(evalsDir, "runs", runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating run dir: %v\n", err)
		os.Exit(1)
	}

	evalFiles, err := collectEvalFiles(evalsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting eval files: %v\n", err)
		os.Exit(1)
	}
	if len(evalFiles) == 0 {
		fmt.Println("no eval files found")
		os.Exit(0)
	}

	token := resolveToken()
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: no GitHub token found.")
		fmt.Fprintln(os.Stderr, "  The Copilot CLI reads auth from shared XDG config dirs, so without an")
		fmt.Fprintln(os.Stderr, "  explicit token eval sessions will load smaqit-adk's VS Code context,")
		fmt.Fprintln(os.Stderr, "  invalidating results. Run via 'make evals' (auto-detects gh auth token)")
		fmt.Fprintln(os.Stderr, "  or set GH_TOKEN explicitly.")
		os.Exit(1)
	}

	fmt.Printf("running %d evals...\n\n", len(evalFiles))

	var results []EvalResult
	for i, path := range evalFiles {
		rel, _ := filepath.Rel(evalsDir, path)
		fmt.Printf("[%d/%d] %s ... ", i+1, len(evalFiles), rel)
		evalCtx, evalCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		result := runEval(evalCtx, path, evalsDir, repoRoot, runDir, token)
		evalCancel()
		results = append(results, result)
		printResultLine(result) // live progress
	}

	// Tally.
	passed, failed := 0, 0
	for _, r := range results {
		if r.Pass {
			passed++
		} else {
			failed++
		}
	}

	// Write reports.
	report := JSONReport{
		RunID:     runID,
		Timestamp: now.UTC().Format(time.RFC3339),
		Total:     len(results),
		Passed:    passed,
		Failed:    failed,
		Evals:     results,
	}
	if err := writeJSONReport(runDir, report); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write results.json: %v\n", err)
	}
	if err := writeMarkdownReport(runDir, report); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write report.md: %v\n", err)
	}

	fmt.Printf("\n%d/%d passed — run report: %s\n", passed, len(results), runDir)

	if failed > 0 {
		os.Exit(1)
	}
}

// ── Eval execution ────────────────────────────────────────────────────────────

func runEval(ctx context.Context, evalPath, evalsDir, repoRoot, runDir, token string) EvalResult {
	// Store a relative path in the report for readability.
	rel, err := filepath.Rel(evalsDir, evalPath)
	if err != nil {
		rel = evalPath
	}
	result := EvalResult{EvalFile: rel}

	ef, err := loadEvalFile(evalPath)
	if err != nil {
		result.Error = fmt.Sprintf("load: %v", err)
		return result
	}
	result.Description = ef.Description

	// Read artifact file relative to repo root.
	artifactPath := filepath.Join(repoRoot, ef.ArtifactFile)
	artifactBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		result.Error = fmt.Sprintf("read artifact %s: %v", ef.ArtifactFile, err)
		return result
	}

	// Create isolated workspace OUTSIDE the project tree. Placing it inside the
	// smaqit-adk repo would cause git-root discovery to expose the full project
	// context to the agent instead of the scaffold workspace only.
	safeName := strings.NewReplacer("/", "__", "\\", "__", ".json", "").Replace(rel)
	tmpDir, err := os.MkdirTemp("", "smaqit-eval-*")
	if err != nil {
		result.Error = fmt.Sprintf("mkdir workspace: %v", err)
		return result
	}
	// After the eval, copy the workspace into the run dir for inspection, then
	// remove the temp dir.
	defer func() {
		preserveDir := filepath.Join(runDir, "workspaces", safeName)
		_ = copyDir(tmpDir, preserveDir)
		_ = os.RemoveAll(tmpDir)
	}()
	if err := setupWorkspace(tmpDir, repoRoot); err != nil {
		result.Error = fmt.Sprintf("setup workspace: %v", err)
		return result
	}

	// Separate ask_user answers from direct send turns.
	var askAnswers []string
	var sendTurns []string
	for _, t := range ef.Turns {
		if t.Trigger == "ask_user" {
			askAnswers = append(askAnswers, t.UserInput)
		} else {
			sendTurns = append(sendTurns, t.UserInput)
		}
	}

	// ask_user queue — consumed by OnUserInputRequest.
	// pairs are also accumulated into askUserTranscript so the grader can see them.
	var queueIdx atomic.Int32
	var askUserTranscript strings.Builder

	client := copilot.NewClient(&copilot.ClientOptions{
		Cwd:         tmpDir,
		GitHubToken: token, // empty → UseLoggedInUser (default true) handles local auth
	})

	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "replace",
			Content: string(artifactBytes),
		},
		WorkingDirectory: tmpDir,
		OnPermissionRequest: func(req copilot.PermissionRequest, _ copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
			// Deny shell execution; approve all other tool access.
			if req.Kind == copilot.PermissionRequestKindShell {
				return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindDeniedByRules}, nil
			}
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindApproved}, nil
		},
		OnUserInputRequest: func(req copilot.UserInputRequest, _ copilot.UserInputInvocation) (copilot.UserInputResponse, error) {
			idx := int(queueIdx.Add(1)) - 1
			if idx >= len(askAnswers) {
				return copilot.UserInputResponse{}, fmt.Errorf(
					"ask_user queue exhausted at index %d (eval provided %d answers)", idx, len(askAnswers))
			}
			answer := askAnswers[idx]
			// Record the Q&A for the grader transcript.
			if req.Question != "" {
				fmt.Fprintf(&askUserTranscript, "ASSISTANT (ask_user): %s\n\nUSER (ask_user): %s\n\n", req.Question, answer)
			} else {
				fmt.Fprintf(&askUserTranscript, "USER (ask_user): %s\n\n", answer)
			}
			return copilot.UserInputResponse{Answer: answer}, nil
		},
	})
	if err != nil {
		result.Error = fmt.Sprintf("create session: %v", err)
		return result
	}

	// Drive non-triggered turns sequentially.
	for _, prompt := range sendTurns {
		if _, err := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: prompt}); err != nil {
			result.Error = fmt.Sprintf("send: %v", err)
			return result
		}
	}

	// Continuation: skills sometimes emit an intermediate text response between
	// gathering sections, causing SendAndWait to return before all ask_user
	// answers are consumed. Send remaining queued answers as regular messages to
	// keep the conversation alive. ask_user callbacks take priority — if one fires
	// during a continuation SendAndWait, it atomically claims the next answer
	// before the continuation loop's post-send check.
	const maxContinuations = 30
	for cont := 0; cont < maxContinuations && int(queueIdx.Load()) < len(askAnswers); cont++ {
		idx := int(queueIdx.Add(1)) - 1
		if idx >= len(askAnswers) {
			break
		}
		answer := askAnswers[idx]
		if _, err := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: answer}); err != nil {
			break
		}
	}

	// Collect full transcript.
	// Prepend ask_user Q&A pairs (invisible to GetMessages) so the grader
	// has the complete interaction history.
	events, err := session.GetMessages(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("get messages: %v", err)
		return result
	}
	workspaceFiles := collectWorkspaceFiles(tmpDir)
	transcript := askUserTranscript.String() + buildTranscript(events) + workspaceFiles

	// Grade each criterion.
	for _, criterion := range ef.ExpectedBehavior {
		pass, reason := grade(ctx, token, transcript, criterion)
		result.Criteria = append(result.Criteria, CriterionResult{
			Criterion: criterion,
			Pass:      pass,
			Reason:    reason,
			Forbidden: false,
		})
	}
	for _, criterion := range ef.ForbiddenBehavior {
		present, reason := grade(ctx, token, transcript, criterion)
		result.Criteria = append(result.Criteria, CriterionResult{
			Criterion: criterion,
			Pass:      !present, // forbidden → must NOT be present
			Reason:    reason,
			Forbidden: true,
		})
	}

	result.Pass = true
	for _, c := range result.Criteria {
		if !c.Pass {
			result.Pass = false
			break
		}
	}
	return result
}

// ── Workspace file collector ─────────────────────────────────────────────────

// collectWorkspaceFiles walks the workspace directory and returns the contents
// of every file that was not placed there by setupWorkspace. This lets the
// grader evaluate file outputs (e.g. compiled agent files) even when the agent
// wrote them to disk without echoing them in the chat transcript.
func collectWorkspaceFiles(dir string) string {
	// Paths created by setupWorkspace — exclude from grader input.
	skip := map[string]bool{
		filepath.Join(dir, "templates"): true,
		filepath.Join(dir, "framework"): true,
		filepath.Join(dir, ".github", "agents", "smaqit.L2.agent.md"): true,
	}

	var sb strings.Builder
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error { //nolint:errcheck
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip entire subtrees created by setupWorkspace.
			if skip[path] {
				return filepath.SkipDir
			}
			return nil
		}
		if skip[path] {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			fmt.Fprintf(&sb, "\n--- workspace file: %s ---\n%s\n", rel, string(data))
		}
		return nil
	})
	if sb.Len() == 0 {
		return ""
	}
	return "\n\n=== Files written to workspace ===\n" + sb.String()
}

// ── Grader ────────────────────────────────────────────────────────────────────

// grade asks a second Copilot session whether the transcript satisfies criterion.
// Returns (present, reason): present=true means the criterion was observed in the transcript.
func grade(ctx context.Context, token, transcript, criterion string) (bool, string) {
	client := copilot.NewClient(&copilot.ClientOptions{
		GitHubToken: token,
	})
	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		OnPermissionRequest: func(_ copilot.PermissionRequest, _ copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindApproved}, nil
		},
	})
	if err != nil {
		return false, fmt.Sprintf("grader session: %v", err)
	}

	prompt := fmt.Sprintf(
		"You are an objective evaluator. Read the following conversation transcript, then answer whether the stated criterion was satisfied.\n\n"+
			"Transcript:\n---\n%s\n---\n\n"+
			"Criterion: %q\n\n"+
			"Reply with YES or NO on the first line, then one sentence explaining why.",
		transcript, criterion)

	if _, err := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: prompt}); err != nil {
		return false, fmt.Sprintf("grader send: %v", err)
	}

	events, err := session.GetMessages(ctx)
	if err != nil {
		return false, fmt.Sprintf("grader get messages: %v", err)
	}
	response := lastAssistantContent(events)
	present := strings.HasPrefix(strings.ToUpper(strings.TrimSpace(response)), "YES")
	reason := firstSentenceAfterVerdict(response)
	return present, reason
}

// ── Report writers ────────────────────────────────────────────────────────────

func writeJSONReport(runDir string, report JSONReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, "results.json"), data, 0644)
}

func writeMarkdownReport(runDir string, report JSONReport) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Eval Run: %s\n\n", report.RunID)
	fmt.Fprintf(&sb, "**Date:** %s  \n", report.Timestamp)
	fmt.Fprintf(&sb, "**Result:** %d/%d passed", report.Passed, report.Total)
	if report.Failed > 0 {
		fmt.Fprintf(&sb, " — **%d FAILED**", report.Failed)
	}
	fmt.Fprintf(&sb, "\n\n")

	// Summary table.
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "| Result | Eval | Description |\n")
	fmt.Fprintf(&sb, "|--------|------|-------------|\n")
	for _, r := range report.Evals {
		status := "✓ PASS"
		if !r.Pass {
			status = "✗ FAIL"
		}
		if r.Error != "" {
			status = "✗ ERROR"
		}
		fmt.Fprintf(&sb, "| %s | `%s` | %s |\n", status, r.EvalFile, r.Description)
	}
	fmt.Fprintf(&sb, "\n")

	// Detail per eval.
	fmt.Fprintf(&sb, "## Details\n\n")
	for _, r := range report.Evals {
		status := "PASS"
		if !r.Pass {
			status = "FAIL"
		}
		fmt.Fprintf(&sb, "### %s — %s\n\n", r.EvalFile, status)
		if r.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", r.Description)
		}
		if r.Error != "" {
			fmt.Fprintf(&sb, "**ERROR:** %s\n\n", r.Error)
			continue
		}
		for _, c := range r.Criteria {
			mark := "✓"
			label := "expected"
			if !c.Pass {
				mark = "✗"
			}
			if c.Forbidden {
				label = "forbidden"
			}
			fmt.Fprintf(&sb, "- %s `[%s]` %s\n", mark, label, c.Criterion)
			if c.Reason != "" {
				fmt.Fprintf(&sb, "  > %s\n", c.Reason)
			}
		}
		fmt.Fprintf(&sb, "\n")
	}

	return os.WriteFile(filepath.Join(runDir, "report.md"), []byte(sb.String()), 0644)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildTranscript(events []copilot.SessionEvent) string {
	var sb strings.Builder
	for _, ev := range events {
		switch ev.Type {
		case copilot.SessionEventTypeUserMessage:
			if ev.Data.Content != nil {
				fmt.Fprintf(&sb, "USER: %s\n\n", *ev.Data.Content)
			}
		case copilot.SessionEventTypeAssistantMessage:
			if ev.Data.Content != nil {
				fmt.Fprintf(&sb, "ASSISTANT: %s\n\n", *ev.Data.Content)
			}
		}
	}
	return sb.String()
}

func lastAssistantContent(events []copilot.SessionEvent) string {
	last := ""
	for _, ev := range events {
		if ev.Type == copilot.SessionEventTypeAssistantMessage && ev.Data.Content != nil {
			last = *ev.Data.Content
		}
	}
	return last
}

func firstSentenceAfterVerdict(response string) string {
	lines := strings.SplitN(response, "\n", 2)
	if len(lines) > 1 {
		rest := strings.TrimSpace(lines[1])
		if idx := strings.IndexAny(rest, ".!?"); idx >= 0 {
			return rest[:idx+1]
		}
		return rest
	}
	return response
}

func setupWorkspace(dir, repoRoot string) error {
	for _, sub := range []string{
		filepath.Join(".github", "agents"),
		filepath.Join(".github", "skills"),
		filepath.Join(".smaqit", "definitions", "agents"),
		filepath.Join(".smaqit", "definitions", "skills"),
		filepath.Join(".smaqit", "logs"),
	} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			return err
		}
	}
	// Copy ADK artifacts so skills can invoke smaqit.L2 as a subagent and L2
	// can read templates and framework files during compilation.
	for _, entry := range []struct{ src, dst string }{
		{filepath.Join(repoRoot, "agents", "smaqit.L2.agent.md"), filepath.Join(dir, ".github", "agents", "smaqit.L2.agent.md")},
	} {
		if err := copyFile(entry.src, entry.dst); err != nil {
			return fmt.Errorf("copy %s: %w", filepath.Base(entry.src), err)
		}
	}
	for _, tree := range []string{"templates", "framework"} {
		if err := copyDir(filepath.Join(repoRoot, tree), filepath.Join(dir, tree)); err != nil {
			return fmt.Errorf("copy %s: %w", tree, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func loadEvalFile(path string) (EvalFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EvalFile{}, err
	}
	var ef EvalFile
	if err := json.Unmarshal(data, &ef); err != nil {
		return EvalFile{}, fmt.Errorf("%s: %w", path, err)
	}
	return ef, nil
}

func collectEvalFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the runner and runs directories.
		if info.IsDir() && (info.Name() == "runner" || info.Name() == "runs") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func resolveToken() string {
	for _, env := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return "" // UseLoggedInUser (SDK default) handles local Copilot auth
}

func printResultLine(r EvalResult) {
	status := "PASS"
	if !r.Pass {
		status = "FAIL"
	}
	if r.Error != "" {
		status = "ERROR"
	}
	fmt.Printf("[%s]\n", status)
	if r.Error != "" {
		fmt.Printf("  ERROR: %s\n", r.Error)
		return
	}
	for _, c := range r.Criteria {
		if !c.Pass {
			label := "expected"
			if c.Forbidden {
				label = "forbidden"
			}
			fmt.Printf("  ✗ [%s] %s\n", label, c.Criterion)
			if c.Reason != "" {
				fmt.Printf("        → %s\n", c.Reason)
			}
		}
	}
}
