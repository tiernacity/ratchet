package executor

import (
	"context"
	"os/exec"
	"runtime"
)

// executorImpl implements the CommandExecutor interface
type executorImpl struct{}

// New creates a new command executor implementation
func New() CommandExecutor {
	return &executorImpl{}
}

// Execute runs a command in the specified directory
func (e *executorImpl) Execute(ctx context.Context, dir, command string) (string, error) {
	var cmd *exec.Cmd

	// Use appropriate shell based on platform
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Set working directory
	cmd.Dir = dir

	// Execute and capture output
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a context cancellation first
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		// Return the raw Go error - let the errors package handle formatting
		return "", err
	}

	return string(output), nil
}
