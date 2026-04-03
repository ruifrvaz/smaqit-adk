package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	binaryPath = os.Getenv("SMAQIT_ADK_BIN")
	if binaryPath == "" {
		panic("SMAQIT_ADK_BIN is not set — run tests via 'make test' from installer/, or set SMAQIT_ADK_BIN manually to the smaqit-adk binary path")
	}
	os.Exit(m.Run())
}

// runBinary invokes the binary in the given working directory and returns
// combined stdout+stderr output and the exit code.
func runBinary(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out), cmd.ProcessState.ExitCode()
}

// mustInit runs init in dir and fails the test if it exits non-zero.
func mustInit(t *testing.T, dir string) {
	t.Helper()
	out, code := runBinary(t, dir, "init")
	if code != 0 {
		t.Fatalf("init failed (exit %d):\n%s", code, out)
	}
}

func TestCmdInit(t *testing.T) {
	dir := t.TempDir()
	mustInit(t, dir)

	// init creates only .github/agents/
	agentDir := filepath.Join(dir, ".github", "agents")
	info, err := os.Stat(agentDir)
	if err != nil || !info.IsDir() {
		t.Errorf("expected directory .github/agents/ after init")
	}

	// Only the two lite-tier agents and two routing skills are installed.
	files := []string{
		".github/agents/smaqit.create-agent.agent.md",
		".github/agents/smaqit.create-skill.agent.md",
		".github/skills/smaqit.create-agent/SKILL.md",
		".github/skills/smaqit.create-skill/SKILL.md",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected file %s after init: %v", f, err)
		}
	}

	// Framework, templates, and Level agents are NOT installed.
	notPresent := []string{
		".smaqit",
		".github/agents/smaqit.L0.agent.md",
		".github/agents/smaqit.L1.agent.md",
		".github/agents/smaqit.L2.agent.md",
	}
	for _, f := range notPresent {
		if _, err := os.Stat(filepath.Join(dir, f)); !os.IsNotExist(err) {
			t.Errorf("expected %s to NOT be present after lite-tier init", f)
		}
	}
}

func TestCmdInit_Idempotent(t *testing.T) {
	// Two independent inits on clean directories produce identical installed files.
	dir1, dir2 := t.TempDir(), t.TempDir()
	mustInit(t, dir1)
	mustInit(t, dir2)

	probes := []string{
		".github/agents/smaqit.create-agent.agent.md",
		".github/agents/smaqit.create-skill.agent.md",
		".github/skills/smaqit.create-agent/SKILL.md",
		".github/skills/smaqit.create-skill/SKILL.md",
	}
	for _, f := range probes {
		c1, err1 := os.ReadFile(filepath.Join(dir1, f))
		c2, err2 := os.ReadFile(filepath.Join(dir2, f))
		if err1 != nil || err2 != nil {
			t.Errorf("file %s missing after init", f)
			continue
		}
		if string(c1) != string(c2) {
			t.Errorf("file %s differs across two init runs — non-deterministic output", f)
		}
	}
}

func TestCmdInit_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	mustInit(t, dir)

	out, code := runBinary(t, dir, "init")
	if code == 0 {
		t.Fatal("second init on already-initialized dir should exit non-zero")
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' message in output, got: %s", out)
	}
	// Non-destructive: existing installation must still be intact.
	for _, f := range []string{
		".github/agents/smaqit.create-agent.agent.md",
		".github/agents/smaqit.create-skill.agent.md",
		".github/skills/smaqit.create-agent/SKILL.md",
		".github/skills/smaqit.create-skill/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("file %s removed by failed second init: %v", f, err)
		}
	}
}

func TestCmdUninstall(t *testing.T) {
	dir := t.TempDir()
	mustInit(t, dir)

	cmd := exec.Command(binaryPath, "uninstall")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y\n")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("uninstall failed:\n%s", string(out))
	}

	for _, removed := range []string{
		".github/agents/smaqit.create-agent.agent.md",
		".github/agents/smaqit.create-skill.agent.md",
		".github/skills/smaqit.create-agent/SKILL.md",
		".github/skills/smaqit.create-skill/SKILL.md",
		".github/agents",
		".github/skills/smaqit.create-agent",
		".github/skills/smaqit.create-skill",
	} {
		if _, err := os.Stat(filepath.Join(dir, removed)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed after uninstall", removed)
		}
	}
}

func TestCmdUninstall_NotInitialized(t *testing.T) {
	dir := t.TempDir()
	_, code := runBinary(t, dir, "uninstall")
	if code != 0 {
		t.Fatalf("uninstall on non-initialized dir should exit 0, got exit code %d", code)
	}
}

func TestCmdVersion(t *testing.T) {
	dir := t.TempDir()
	out, code := runBinary(t, dir, "version")
	if code != 0 {
		t.Fatalf("version failed (exit %d)", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("version output should not be empty")
	}
}
