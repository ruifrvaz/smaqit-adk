package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mustInstallGlobal runs --install-global with HOME set to a fresh, isolated
// temp directory and fails the test if it exits non-zero. Returns the HOME
// path used, for inspection.
func mustInstallGlobal(t *testing.T) string {
	t.Helper()
	out, code, home := runBinaryHome(t, "--install-global")
	if code != 0 {
		t.Fatalf("--install-global failed (exit %d):\n%s", code, out)
	}
	return home
}

func TestInstallGlobal(t *testing.T) {
	home := mustInstallGlobal(t)

	// Agents installed to Claude and Codex global paths, never Copilot.
	agentFiles := []string{
		filepath.Join(home, ".claude", "agents", "smaqit-L0.md"),
		filepath.Join(home, ".claude", "agents", "smaqit-L1.md"),
		filepath.Join(home, ".claude", "agents", "smaqit-L2.md"),
		filepath.Join(home, ".codex", "agents", "smaqit.L0.toml"),
		filepath.Join(home, ".codex", "agents", "smaqit.L1.toml"),
		filepath.Join(home, ".codex", "agents", "smaqit.L2.toml"),
	}
	for _, f := range agentFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected agent file %s after --install-global: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".copilot", "agents")); !os.IsNotExist(err) {
		t.Error("no Copilot agent directory should be created — Copilot is not an authored target")
	}

	// All 5 skills installed to both the shared (Copilot+Codex) and Claude paths.
	skillNames := []string{
		"smaqit.create-agent",
		"smaqit.create-skill",
		"smaqit.new-principle",
		"smaqit.bench-run",
		"smaqit.bench-scaffold",
	}
	for _, base := range []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "skills"),
	} {
		for _, name := range skillNames {
			f := filepath.Join(base, name, "SKILL.md")
			if _, err := os.Stat(f); err != nil {
				t.Errorf("expected skill file %s after --install-global: %v", f, err)
			}
		}
	}

	// Templates and framework installed to the namespaced global data dir.
	dataFiles := []string{
		filepath.Join(home, ".agents", "smaqit-adk", "templates", "agents", "base-agent.template.md"),
		filepath.Join(home, ".agents", "smaqit-adk", "framework", "SMAQIT.md"),
	}
	for _, f := range dataFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected data file %s after --install-global: %v", f, err)
		}
	}
}

func TestInstallGlobal_Idempotent(t *testing.T) {
	// Two independent global installs into two fresh HOMEs produce identical
	// installed files — a re-run (e.g. to upgrade) is safe and deterministic.
	home1, home2 := t.TempDir(), t.TempDir()
	for _, home := range []string{home1, home2} {
		cmd := exec.Command(binaryPath, "--install-global")
		cmd.Env = append(os.Environ(), "HOME="+home)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("--install-global failed for %s: %v\n%s", home, err, out)
		}
	}

	probes := []string{
		filepath.Join(".claude", "agents", "smaqit-L2.md"),
		filepath.Join(".codex", "agents", "smaqit.L2.toml"),
		filepath.Join(".agents", "skills", "smaqit.create-agent", "SKILL.md"),
	}
	for _, rel := range probes {
		c1, err1 := os.ReadFile(filepath.Join(home1, rel))
		c2, err2 := os.ReadFile(filepath.Join(home2, rel))
		if err1 != nil || err2 != nil {
			t.Errorf("file %s missing after --install-global", rel)
			continue
		}
		if string(c1) != string(c2) {
			t.Errorf("file %s differs across two --install-global runs — non-deterministic output", rel)
		}
	}
}

func TestInstallGlobal_RerunOverwritesCleanly(t *testing.T) {
	// Re-running --install-global against an already-installed HOME must not
	// error — this is how a version upgrade refreshes files in place.
	home := mustInstallGlobal(t)

	cmd := exec.Command(binaryPath, "--install-global")
	cmd.Env = append(os.Environ(), "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("second --install-global on already-installed HOME should succeed: %v\n%s", err, out)
	}
}

func TestCmdUninstall_Global(t *testing.T) {
	home := mustInstallGlobal(t)

	cmd := exec.Command(binaryPath, "uninstall")
	cmd.Env = append(os.Environ(), "HOME="+home)
	cmd.Stdin = strings.NewReader("y\n")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("uninstall failed:\n%s", string(out))
	}

	removed := []string{
		filepath.Join(home, ".claude", "agents", "smaqit-L0.md"),
		filepath.Join(home, ".claude", "agents", "smaqit-L1.md"),
		filepath.Join(home, ".claude", "agents", "smaqit-L2.md"),
		filepath.Join(home, ".codex", "agents", "smaqit.L0.toml"),
		filepath.Join(home, ".codex", "agents", "smaqit.L1.toml"),
		filepath.Join(home, ".codex", "agents", "smaqit.L2.toml"),
		filepath.Join(home, ".agents", "skills", "smaqit.create-agent"),
		filepath.Join(home, ".claude", "skills", "smaqit.create-agent"),
		filepath.Join(home, ".agents", "smaqit-adk"),
	}
	for _, f := range removed {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed after uninstall", f)
		}
	}

	// Shared parent directories must survive — other tools/products may
	// still have content there.
	for _, kept := range []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "agents"),
		filepath.Join(home, ".codex", "agents"),
	} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("expected shared directory %s to survive uninstall: %v", kept, err)
		}
	}
}

func TestCmdUninstall_DoesNotTouchUnrelatedSharedContent(t *testing.T) {
	home := mustInstallGlobal(t)

	// Simulate another tool/product's own content sitting in the same
	// shared directories smaqit-adk installs into.
	otherSkill := filepath.Join(home, ".agents", "skills", "some-other-product.skill")
	if err := os.MkdirAll(otherSkill, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(otherSkill, "SKILL.md"), []byte("unrelated"), 0644); err != nil {
		t.Fatal(err)
	}
	otherAgent := filepath.Join(home, ".claude", "agents", "some-other-product.md")
	if err := os.WriteFile(otherAgent, []byte("unrelated"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binaryPath, "uninstall")
	cmd.Env = append(os.Environ(), "HOME="+home)
	cmd.Stdin = strings.NewReader("y\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("uninstall failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(otherSkill); err != nil {
		t.Errorf("uninstall must not remove unrelated shared content: %s: %v", otherSkill, err)
	}
	if _, err := os.Stat(otherAgent); err != nil {
		t.Errorf("uninstall must not remove unrelated shared content: %s: %v", otherAgent, err)
	}
}

func TestCmdUninstall_Cancelled(t *testing.T) {
	home := mustInstallGlobal(t)

	cmd := exec.Command(binaryPath, "uninstall")
	cmd.Env = append(os.Environ(), "HOME="+home)
	cmd.Stdin = strings.NewReader("n\n")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("declined uninstall should exit 0:\n%s", string(out))
	}
	if !strings.Contains(string(out), "cancelled") {
		t.Errorf("expected cancellation message, got: %s", out)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "agents", "smaqit-L2.md")); err != nil {
		t.Error("declined uninstall must not remove any files")
	}
}

func TestCmdNoArgs_PrintsHelp(t *testing.T) {
	cmd := exec.Command(binaryPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("bare invocation with no args should exit non-zero")
	}
	if !strings.Contains(string(out), "Usage: smaqit-adk") {
		t.Errorf("expected usage/help output, got: %s", out)
	}
}

func TestCmdHelp_DoesNotMentionInstallGlobalFlag(t *testing.T) {
	out, code, _ := runBinaryHome(t, "help")
	if code != 0 {
		t.Fatalf("help failed (exit %d):\n%s", code, out)
	}
	if strings.Contains(out, "--install-global") {
		t.Error("--install-global must stay undocumented — install.sh is the sole global-install entry point")
	}
}

func TestCmdVersion_Simple(t *testing.T) {
	out, code, _ := runBinaryHome(t, "version")
	if code != 0 {
		t.Fatalf("version failed (exit %d)", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("version output should not be empty")
	}
}
