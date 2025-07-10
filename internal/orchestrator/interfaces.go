package orchestrator

import (
	"context"
	"io"

	"github.com/tiernacity/ratchet/internal/config"
)

// Orchestrator defines the main interface for coordinating the ratchet workflow
type Orchestrator interface {
	Run(ctx context.Context, cfg *config.Config) error
}

// GitOperations defines the interface for Git operations
type GitOperations interface {
	IsGitRepository() error
	CreateWorktree(path, branch string) error
	RemoveWorktree(path string) error
	ResolveBranch(branch string) (string, error)
}

// CommandExecutor defines the interface for executing shell commands
type CommandExecutor interface {
	Execute(ctx context.Context, dir, command string) (string, error)
}

// MetricParser defines the interface for parsing metric output
type MetricParser interface {
	Parse(output string) (float64, error)
}

// ProgressReporter defines the interface for reporting progress
type ProgressReporter interface {
	Start(baseRef, headRef string, verbose bool)
	UpdateBranch(branch string, phase string, completed bool)
	Success(current, base float64, operator, branch string)
	Failure(current, base float64, operator, branch string)
	Error(message string)
	NoComparison(value float64)
	Info(message string)
}

// OutputWriter defines the interface for writing output
type OutputWriter interface {
	Stdout() io.Writer
	Stderr() io.Writer
}