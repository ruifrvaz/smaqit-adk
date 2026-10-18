// tests the claude-code adapter against a fake headless Claude Code script.
package bench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const fakeClaudeToken = "fake-oauth-token-SENTINEL-0123456789"

// fakeClaude imitates `claude -p --output-format stream-json`: it reports the
// skills and agents it finds in CLAUDE_CONFIG_DIR, writes a session transcript,
// and records what it observed into report.txt in its working directory.
// FAKE_MODE selects a misbehavior: hide (omit discovered treatments), error
// (is_error result), or leak (echo the credential into output and a file).
const fakeClaude = `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.9.9 (Fake Claude)"; exit 0; fi
session=""; prev=""
for a in "$@"; do [ "$prev" = "--session-id" ] && session="$a"; prev="$a"; done
brief=$(cat)
printf '%s\n' "$@" > argv.txt
{
  echo "home=$HOME"
  echo "config=$CLAUDE_CONFIG_DIR"
  echo "cwd=$(pwd)"
  [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ] && echo "token=present"
  [ -n "$CLAUDE_CODE_SESSION_ID" ] && echo "leaked=session"
  [ -n "$ANTHROPIC_API_KEY" ] && echo "leaked=apikey"
  case "$brief" in *"Case brief"*) echo "brief=stdin";; esac
} > report.txt
skills=""; agents=""
if [ "$FAKE_MODE" != "hide" ]; then
  for d in "$CLAUDE_CONFIG_DIR"/skills/*/; do [ -d "$d" ] || continue; n=${d%/}; n=${n##*/}; skills="$skills${skills:+,}\"$n\""; done
  for f in "$CLAUDE_CONFIG_DIR"/commands/*.md; do [ -f "$f" ] || continue; n=${f##*/}; n=${n%.md}; skills="$skills${skills:+,}\"$n\""; done
  for f in "$CLAUDE_CONFIG_DIR"/agents/*.md; do [ -f "$f" ] || continue; n=${f##*/}; n=${n%.md}; agents="$agents${agents:+,}\"$n\""; done
fi
mkdir -p "$CLAUDE_CONFIG_DIR/projects/encoded-cwd"
echo '{"type":"user","transcript":true}' > "$CLAUDE_CONFIG_DIR/projects/encoded-cwd/$session.jsonl"
echo '{"type":"system","subtype":"init","skills":['"$skills"'],"agents":['"$agents"'],"slash_commands":['"$skills"']}'
echo '{"type":"assistant","message":{"content":[{"type":"text"},{"type":"tool_use","name":"Read"}]}}'
iserr=false; [ "$FAKE_MODE" = "error" ] && iserr=true
if [ "$FAKE_MODE" = "leak" ]; then echo "token is $CLAUDE_CODE_OAUTH_TOKEN"; echo "$CLAUDE_CODE_OAUTH_TOKEN" > leaked.txt; fi
echo '{"type":"result","subtype":"success","is_error":'"$iserr"',"total_cost_usd":0.25,"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":2,"cache_read_input_tokens":3}}'
`

type claudeFixture struct {
	root, executable string
}

func newClaudeFixture(t *testing.T) claudeFixture {
	t.Helper()
	requirePOSIX(t)
	root := t.TempDir()
	executable := filepath.Join(root, "bin", "claude")
	write(t, executable, fakeClaude)
	if err := os.Chmod(executable, 0755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "treatments", "skill", "SKILL.md"), "---\nname: probe-skill\ndescription: Probe skill.\n---\n\nBody.\n")
	write(t, filepath.Join(root, "treatments", "probe-agent.md"), "---\nname: probe-agent\ndescription: Probe agent.\n---\n\nBody.\n")
	write(t, filepath.Join(root, "treatments", "probe-command.md"), "Run the probe.\n")
	write(t, filepath.Join(root, "fixture", "README.md"), "fixture\n")
	managed := filepath.Join(root, "managed")
	original := managedClaudeRoots
	managedClaudeRoots = func() []string { return []string{managed} }
	t.Cleanup(func() { managedClaudeRoots = original })
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", fakeClaudeToken)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "parent-session")
	t.Setenv("ANTHROPIC_API_KEY", "parent-api-key-must-not-pass")
	return claudeFixture{root: root, executable: executable}
}

func (f claudeFixture) manifest(extraEnvironment string) string {
	return `schemaVersion: 2
name: claude
cases:
  - id: case-1
    fixture: {source: ./fixture}
    given:
      prompt: {text: Use the probe treatments.}
    expect:
      - {id: report, type: file, actual: submission, path: report.txt, operator: exists}
variants:
  - id: with-treatment
    adapter: claude-code
    claudeCode:
      executable: ` + f.executable + `
      model: claude-haiku-5-5
      permissionMode: bypassPermissions
      maxTurns: 3
      allowedTools: [Read, Edit]
      credential: CLAUDE_CODE_OAUTH_TOKEN
      install:
        - {treatment: skill, as: skill}
        - {treatment: agent, as: agent}
        - {treatment: command, as: command}` + extraEnvironment + `
    treatment:
      - {id: skill, source: ./treatments/skill}
      - {id: agent, source: ./treatments/probe-agent.md}
      - {id: command, source: ./treatments/probe-command.md}
    intendedDifferences: [probe treatments installed]
execution: {repetitions: 1, timeoutSeconds: 20}
output: {directory: ./results}
`
}

func fakeMode(mode string) string {
	return "\n      environment: {set: {FAKE_MODE: " + mode + "}}"
}

func readSubmission(t *testing.T, result RunResult, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(result.ArtifactDirectory, "submission", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertNoSecretInDirectory(t *testing.T, directory, secret string) {
	t.Helper()
	_ = filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil && strings.Contains(string(data), secret) {
			t.Errorf("credential found in evidence file %s", path)
		}
		return nil
	})
}

func TestClaudeCodeAdapterRunsIsolatedWithInstalledTreatments(t *testing.T) {
	f := newClaudeFixture(t)
	experiment := runManifest(t, f.root, f.manifest(""))
	result := experiment.Results[0]
	if result.Failure != nil || !result.RequiredPassed {
		t.Fatalf("run failed: %+v %+v", result.Failure, result.Grades)
	}
	report := readSubmission(t, result, "report.txt")
	for _, want := range []string{"token=present", "brief=stdin"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "leaked=") {
		t.Errorf("parent CLAUDE_*/ANTHROPIC_* variables reached the harness:\n%s", report)
	}
	values := map[string]string{}
	for _, line := range strings.Split(report, "\n") {
		if name, value, ok := strings.Cut(line, "="); ok {
			values[name] = value
		}
	}
	if values["config"] != filepath.Join(values["home"], ".claude") || !strings.Contains(values["home"], "smaqit-bench-home-") {
		t.Errorf("HOME/CLAUDE_CONFIG_DIR not isolated in the harness home: %v", values)
	}
	if strings.HasPrefix(values["home"], values["cwd"]) {
		t.Errorf("harness home must be outside the workspace: %v", values)
	}
	if _, err := os.Stat(values["home"]); !os.IsNotExist(err) {
		t.Errorf("harness home must be removed after the attempt: %v", err)
	}
	argv := readSubmission(t, result, "argv.txt")
	for _, want := range []string{"-p\n", "--output-format\nstream-json\n", "--verbose\n", "--permission-mode\nbypassPermissions\n", "--model\nclaude-haiku-5-5\n", "--max-turns\n3\n", "--allowedTools\nRead\nEdit\n"} {
		if !strings.Contains(argv, want) {
			t.Errorf("argv lacks %q:\n%s", want, argv)
		}
	}
	if result.Harness.Name != "claude-code" || result.Harness.Version != "9.9.9 (Fake Claude)" {
		t.Errorf("harness identity = %+v", result.Harness)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(result.Harness.SessionID) || !strings.Contains(argv, result.Harness.SessionID) {
		t.Errorf("session id %q not a v4 UUID passed to the harness", result.Harness.SessionID)
	}
	usage := result.Usage
	if usage.InputTokens == nil || *usage.InputTokens != 15 || *usage.OutputTokens != 5 || *usage.TotalTokens != 20 || *usage.EstimatedCost != 0.25 || *usage.ToolCalls != 1 {
		t.Errorf("usage not parsed from the result event: %+v", usage)
	}
	if _, err := os.Stat(filepath.Join(result.ArtifactDirectory, "traces", "transcript.jsonl")); err != nil {
		t.Errorf("session transcript not captured: %v", err)
	}
	request, err := os.ReadFile(filepath.Join(result.ArtifactDirectory, "request.json"))
	if err != nil || !strings.Contains(string(request), `"credential": "CLAUDE_CODE_OAUTH_TOKEN"`) {
		t.Errorf("request.json lacks harness identity: %v", err)
	}
	reportMD, err := os.ReadFile(filepath.Join(experiment.Directory, "report.md"))
	if err != nil || !strings.Contains(string(reportMD), "## Harnesses") || !strings.Contains(string(reportMD), "9.9.9 (Fake Claude)") {
		t.Errorf("report.md lacks the harness line: %v", err)
	}
	assertNoSecretInDirectory(t, experiment.Directory, fakeClaudeToken)
	assertNoSecretInDirectory(t, experiment.Directory, "parent-api-key-must-not-pass")
}

func TestClaudeCodeFixtureDefinitionShadowingFailsRun(t *testing.T) {
	f := newClaudeFixture(t)
	write(t, filepath.Join(f.root, "fixture", ".claude", "agents", "probe-agent.md"), "---\nname: probe-agent\n---\n")
	result := runManifest(t, f.root, f.manifest("")).Results[0]
	if result.Failure == nil || !strings.Contains(result.Failure.Message, `agent "probe-agent" is also defined under`) {
		t.Fatalf("expected a shadowing failure, got %+v", result.Failure)
	}
}

func TestClaudeCodeManagedDefinitionShadowingFailsRun(t *testing.T) {
	f := newClaudeFixture(t)
	write(t, filepath.Join(f.root, "managed", ".claude", "skills", "other-dir", "SKILL.md"), "---\nname: probe-skill\n---\n")
	result := runManifest(t, f.root, f.manifest("")).Results[0]
	if result.Failure == nil || !strings.Contains(result.Failure.Message, `skill "probe-skill" is also defined under`) {
		t.Fatalf("expected a managed shadowing failure, got %+v", result.Failure)
	}
}

func TestClaudeCodeUndiscoveredTreatmentFailsRun(t *testing.T) {
	f := newClaudeFixture(t)
	result := runManifest(t, f.root, f.manifest(fakeMode("hide"))).Results[0]
	if result.Failure == nil || !strings.Contains(result.Failure.Message, "not discovered by Claude Code") {
		t.Fatalf("expected a discovery failure, got %+v", result.Failure)
	}
}

func TestClaudeCodeReportedErrorFailsRequiredExpectations(t *testing.T) {
	f := newClaudeFixture(t)
	result := runManifest(t, f.root, f.manifest(fakeMode("error"))).Results[0]
	if !result.Harness.ReportedError || result.RequiredPassed {
		t.Fatalf("is_error must fail the attempt even with exit code 0: %+v", result)
	}
}

func TestClaudeCodeCredentialCanaryRedactsEvidence(t *testing.T) {
	f := newClaudeFixture(t)
	experiment := runManifest(t, f.root, f.manifest(fakeMode("leak")))
	result := experiment.Results[0]
	if result.Failure == nil || result.Failure.Phase != "evidence" || result.RequiredPassed {
		t.Fatalf("leaked credential must fail the attempt in phase evidence: %+v", result.Failure)
	}
	for _, leaked := range []string{"submission/leaked.txt", "traces/harness.stdout.log"} {
		if !strings.Contains(result.Failure.Message, leaked) {
			t.Errorf("failure message does not name %s: %s", leaked, result.Failure.Message)
		}
	}
	if got := readSubmission(t, result, "leaked.txt"); !strings.Contains(got, "<redacted>") {
		t.Errorf("leaked file not redacted: %q", got)
	}
	assertNoSecretInDirectory(t, experiment.Directory, fakeClaudeToken)
}

func TestClaudeCodeMissingCredentialFailsRun(t *testing.T) {
	f := newClaudeFixture(t)
	os.Unsetenv("CLAUDE_CODE_OAUTH_TOKEN")
	result := runManifest(t, f.root, f.manifest("")).Results[0]
	if result.Failure == nil || !strings.Contains(result.Failure.Message, "credential variable CLAUDE_CODE_OAUTH_TOKEN is not set") {
		t.Fatalf("expected a missing-credential failure, got %+v", result.Failure)
	}
}

func TestClaudeCodePlanRecordsHarnessVersion(t *testing.T) {
	f := newClaudeFixture(t)
	manifestPath := filepath.Join(f.root, "bench.yaml")
	write(t, manifestPath, f.manifest(""))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Harnesses) != 1 || plan.Harnesses[0].Version != "9.9.9 (Fake Claude)" || plan.Harnesses[0].Name != "claude-code" {
		t.Fatalf("plan harnesses = %+v", plan.Harnesses)
	}
	executableHashed := false
	for _, asset := range plan.Assets {
		if asset.Kind == "executable" && asset.Path == plan.Harnesses[0].Executable {
			executableHashed = true
		}
	}
	if !executableHashed {
		t.Fatal("claude executable must be hashed into the plan")
	}
	data, err := json.Marshal(plan)
	if err != nil || strings.Contains(string(data), fakeClaudeToken) {
		t.Fatalf("plan must not contain the credential: %v", err)
	}
}

func TestClaudeCodeManifestValidation(t *testing.T) {
	f := newClaudeFixture(t)
	write(t, filepath.Join(f.root, "treatments", "nameless-agent.md"), "No frontmatter.\n")
	base := f.manifest("")
	cases := []struct{ name, old, new, want string }{
		{"credential", "credential: CLAUDE_CODE_OAUTH_TOKEN", "credential: GITHUB_TOKEN", "must be ANTHROPIC_API_KEY or CLAUDE_CODE_OAUTH_TOKEN"},
		{"permission mode", "permissionMode: bypassPermissions", "permissionMode: yolo", "permissionMode: must be one of"},
		{"bare flag", "maxTurns: 3", "maxTurns: 3\n      arguments: [--bare]", "flag --bare is controlled by the claude-code adapter"},
		{"session flag", "maxTurns: 3", "maxTurns: 3\n      arguments: [--session-id=x]", "flag --session-id=x is controlled"},
		{"inherit claude", "maxTurns: 3", "maxTurns: 3\n      environment: {inherit: [CLAUDE_CODE_SESSION_ID]}", "must not be inherited"},
		{"set anthropic", "maxTurns: 3", "maxTurns: 3\n      environment: {set: {ANTHROPIC_BASE_URL: x}}", "controlled by the claude-code adapter or is credential-shaped"},
		{"unknown treatment", "{treatment: command, as: command}", "{treatment: missing, as: command}", "must name one of this variant's treatments"},
		{"bad kind", "{treatment: command, as: command}", "{treatment: command, as: plugin}", "must be skill, agent, or command"},
		{"skill not dir", "{treatment: skill, as: skill}", "{treatment: command, as: skill}", "a skill treatment must be a directory"},
		{"agent without name", "./treatments/probe-agent.md", "./treatments/nameless-agent.md", "an agent treatment must declare a frontmatter name"},
		{"process too", "    adapter: claude-code\n", "    adapter: claude-code\n    process: {executable: sh}\n", "must not be set for claude-code adapter"},
		{"missing config", "    adapter: claude-code\n", "    adapter: process\n", "claudeCode: must not be set for process adapter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(base, tc.old) {
				t.Fatalf("test setup: %q not in manifest", tc.old)
			}
			path := filepath.Join(f.root, "bench.yaml")
			write(t, path, strings.Replace(base, tc.old, tc.new, 1))
			if _, err := LoadManifest(path); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}
