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
// combined stdout+stderr output and the exit code.
func runBinary(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out), cmd.ProcessState.ExitCode()
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

