package executor

import "context"

// CommandExecutor defines the interface for executing shell commands
type CommandExecutor interface {
	// Execute runs a command in the specified directory and returns stdout and error
	Execute(ctx context.Context, dir, command string) (string, error)
}
