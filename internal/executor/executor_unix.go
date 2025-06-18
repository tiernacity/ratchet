//go:build !windows
// +build !windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func setProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}

	// Unix-like - kill the process group
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			// Process group may have already exited
			fmt.Fprintf(os.Stderr, "Warning: failed to send SIGTERM to process group: %v\n", err)
		}
		// Give it a moment to terminate gracefully
		time.Sleep(100 * time.Millisecond)
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			// Process group may have already exited
			fmt.Fprintf(os.Stderr, "Warning: failed to send SIGKILL to process group: %v\n", err)
		}
	} else {
		if err := cmd.Process.Kill(); err != nil {
			// Process may have already exited
			fmt.Fprintf(os.Stderr, "Warning: failed to kill process: %v\n", err)
		}
	}
}
