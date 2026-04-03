package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mustLite runs lite in dir and fails the test if it exits non-zero.
func mustLite(t *testing.T, dir string) {
	t.Helper()
	out, code := runBinary(t, dir, "lite")
	if code != 0 {
		t.Fatalf("lite failed (exit %d):\n%s", code, out)
	}
}

func TestCmdLite(t *testing.T) {
	dir := t.TempDir()
	mustLite(t, dir)

	// lite creates only .github/agents/
	agentDir := filepath.Join(dir, ".github", "agents")
	info, err := os.Stat(agentDir)
	if err != nil || !info.IsDir() {
		t.Errorf("expected directory .github/agents/ after lite")
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
			t.Errorf("expected file %s after lite: %v", f, err)
		}
	}

	// Framework, templates, and Level agents are NOT installed.
	notPresent := []string{
		".smaqit/agents/smaqit.L0.agent.md",
		".smaqit/agents/smaqit.L1.agent.md",
		".smaqit/agents/smaqit.L2.agent.md",
		".smaqit/framework",
		".smaqit/templates",
	}
	for _, f := range notPresent {
		if _, err := os.Stat(filepath.Join(dir, f)); !os.IsNotExist(err) {
			t.Errorf("expected %s to NOT be present after lite install", f)
		}
	}
}

func TestCmdLite_Idempotent(t *testing.T) {
	// Two independent lite installs on clean directories produce identical installed files.
	dir1, dir2 := t.TempDir(), t.TempDir()
	mustLite(t, dir1)
	mustLite(t, dir2)

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
			t.Errorf("file %s missing after lite", f)
			continue
		}
		if string(c1) != string(c2) {
			t.Errorf("file %s differs across two lite runs — non-deterministic output", f)
		}
	}
}

func TestCmdLite_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	mustLite(t, dir)

	out, code := runBinary(t, dir, "lite")
	if code == 0 {
		t.Fatal("second lite on already-initialized dir should exit non-zero")
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
			t.Errorf("file %s removed by failed second lite: %v", f, err)
		}
	}
}

func TestCmdUninstall_Lite(t *testing.T) {
	dir := t.TempDir()
	mustLite(t, dir)

	cmd := exec.Command(binaryPath, "uninstall", "lite")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y\n")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("uninstall lite failed:\n%s", string(out))
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
			t.Errorf("expected %s to be removed after uninstall lite", removed)
		}
	}
}

func TestCmdAdvanced(t *testing.T) {
	dir := t.TempDir()
	out, code := runBinary(t, dir, "advanced")
	if code != 0 {
		t.Fatalf("advanced failed (exit %d):\n%s", code, out)
	}

	// Level agents installed into .smaqit/agents/
	agentFiles := []string{
		".smaqit/agents/smaqit.L0.agent.md",
		".smaqit/agents/smaqit.L1.agent.md",
		".smaqit/agents/smaqit.L2.agent.md",
	}
	for _, f := range agentFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected file %s after advanced: %v", f, err)
		}
	}

	// Framework installed into .smaqit/framework/
	frameworkDir := filepath.Join(dir, ".smaqit", "framework")
	if info, err := os.Stat(frameworkDir); err != nil || !info.IsDir() {
		t.Errorf("expected directory .smaqit/framework/ after advanced")
	}

	// Templates installed into .smaqit/templates/
	templatesDir := filepath.Join(dir, ".smaqit", "templates")
	if info, err := os.Stat(templatesDir); err != nil || !info.IsDir() {
		t.Errorf("expected directory .smaqit/templates/ after advanced")
	}

	// Advanced skills installed into .smaqit/skills/
	skillFiles := []string{
		".smaqit/skills/smaqit.new-agent/SKILL.md",
		".smaqit/skills/smaqit.new-skill/SKILL.md",
	}
	for _, f := range skillFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected file %s after advanced: %v", f, err)
		}
	}

	// Lite-tier .github/ content is NOT installed by advanced
	notPresent := []string{
		".github/agents/smaqit.create-agent.agent.md",
		".github/agents/smaqit.create-skill.agent.md",
	}
	for _, f := range notPresent {
		if _, err := os.Stat(filepath.Join(dir, f)); !os.IsNotExist(err) {
			t.Errorf("expected %s to NOT be present after advanced-only install", f)
		}
	}
}

func TestCmdAdvanced_Idempotent(t *testing.T) {
	// Two independent advanced installs on clean directories produce identical installed files.
	dir1, dir2 := t.TempDir(), t.TempDir()
	for _, dir := range []string{dir1, dir2} {
		out, code := runBinary(t, dir, "advanced")
		if code != 0 {
			t.Fatalf("advanced failed (exit %d):\n%s", code, out)
		}
	}

	probes := []string{
		".smaqit/agents/smaqit.L0.agent.md",
		".smaqit/agents/smaqit.L1.agent.md",
		".smaqit/agents/smaqit.L2.agent.md",
		".smaqit/skills/smaqit.new-agent/SKILL.md",
		".smaqit/skills/smaqit.new-skill/SKILL.md",
	}
	for _, f := range probes {
		c1, err1 := os.ReadFile(filepath.Join(dir1, f))
		c2, err2 := os.ReadFile(filepath.Join(dir2, f))
		if err1 != nil || err2 != nil {
			t.Errorf("file %s missing after advanced", f)
			continue
		}
		if string(c1) != string(c2) {
			t.Errorf("file %s differs across two advanced runs — non-deterministic output", f)
		}
	}
}

func TestCmdAdvanced_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	out, code := runBinary(t, dir, "advanced")
	if code != 0 {
		t.Fatalf("advanced failed (exit %d):\n%s", code, out)
	}

	out, code = runBinary(t, dir, "advanced")
	if code == 0 {
		t.Fatal("second advanced on already-installed dir should exit non-zero")
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' message in output, got: %s", out)
	}
	// Non-destructive: existing installation must still be intact.
	for _, f := range []string{
		".smaqit/agents/smaqit.L0.agent.md",
		".smaqit/skills/smaqit.new-agent/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("file %s removed by failed second advanced: %v", f, err)
		}
	}
}

func TestCmdUninstall_Advanced(t *testing.T) {
	dir := t.TempDir()
	out, code := runBinary(t, dir, "advanced")
	if code != 0 {
		t.Fatalf("advanced failed (exit %d):\n%s", code, out)
	}

	cmd := exec.Command(binaryPath, "uninstall", "advanced")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y\n")
	out2, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("uninstall advanced failed:\n%s", string(out2))
	}

	for _, removed := range []string{
		".smaqit/agents",
		".smaqit/framework",
		".smaqit/templates",
		".smaqit/skills",
	} {
		if _, err := os.Stat(filepath.Join(dir, removed)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed after uninstall advanced", removed)
		}
	}
}

func TestCmdInit_MigrationMessage(t *testing.T) {
	dir := t.TempDir()
	out, code := runBinary(t, dir, "init")
	if code == 0 {
		t.Fatal("init should exit non-zero (migration message)")
	}
	if !strings.Contains(out, "smaqit-adk lite") {
		t.Errorf("expected migration message mentioning 'smaqit-adk lite', got: %s", out)
	}
}
