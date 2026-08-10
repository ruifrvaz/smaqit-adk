//go:build windows

// terminates timed-out process harness trees on Windows.
package bench

import (
	"os/exec"
	"strconv"
)

func configureProcessGroup(command *exec.Cmd) {}
func terminateProcessTree(command *exec.Cmd) {
	if command.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(command.Process.Pid)).Run()
}
