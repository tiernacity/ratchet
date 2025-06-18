//go:build windows
// +build windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
)

func setProcAttr(cmd *exec.Cmd) {
	// No special process attributes needed on Windows
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}

	// Windows - kill the process
	if err := cmd.Process.Kill(); err != nil {
		// Process may have already exited
		fmt.Fprintf(os.Stderr, "Warning: failed to kill process: %v\n", err)
	}
}
