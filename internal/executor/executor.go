package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// Execute runs a command and returns its stdout output
func Execute(command string, workingDir string) (string, error) {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling to cancel context on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()
	defer signal.Stop(sigChan)

	// Use the system shell to execute the command
	var cmd *exec.Cmd
	if os.PathSeparator == '\\' {
		// Windows
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		// Unix-like systems
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Set working directory if provided
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set process group so we can kill child processes (Unix only)
	setProcAttr(cmd)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for command to complete or context to be cancelled
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			// Include stderr in error message for debugging
			stderrStr := strings.TrimSpace(stderr.String())
			if stderrStr != "" {
				return "", fmt.Errorf("command failed: %w\nstderr: %s", err, stderrStr)
			}
			return "", fmt.Errorf("command failed: %w", err)
		}
	case <-ctx.Done():
		// Context was cancelled (likely due to signal), kill the process group
		killProcess(cmd)
		<-done // Wait for cmd.Wait() to return
		return "", fmt.Errorf("command interrupted")
	}

	// Return stdout output
	return strings.TrimSpace(stdout.String()), nil
}
