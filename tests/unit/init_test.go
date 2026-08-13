package unit_test

import (
	"os"
	"os/exec"
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
// combined stdout+stderr output and the exit code. HOME is sandboxed to dir
// itself so no test can ever write into the real developer machine's global
// agent/skill directories, even if a command reads $HOME incidentally.
func runBinary(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+dir)
	out, _ := cmd.CombinedOutput()
	return string(out), cmd.ProcessState.ExitCode()
}

// runBinaryHome invokes the binary with HOME set to a fresh, isolated temp
// directory (never the real developer machine) and returns combined output,
// exit code, and the sandboxed HOME path for inspection.
func runBinaryHome(t *testing.T, args ...string) (out string, code int, home string) {
	t.Helper()
	home = t.TempDir()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+home)
	o, _ := cmd.CombinedOutput()
	return string(o), cmd.ProcessState.ExitCode(), home
}

func TestCmdUninstall_NotInitialized(t *testing.T) {
	_, code, _ := runBinaryHome(t, "uninstall")
	if code != 0 {
		t.Fatalf("uninstall with nothing installed should exit 0, got exit code %d", code)
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

