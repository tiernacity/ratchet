package errors

import (
	"fmt"
)

// RatchetError is the base interface for all ratchet errors
type RatchetError interface {
	error
	ExitCode() int
}

// MetricTestError is used when a metric comparison fails - this is an expected failure scenario
type MetricTestError struct {
	Current  float64
	Base     float64
	Operator string
	Branch   string
}

func (e *MetricTestError) Error() string {
	return fmt.Sprintf("HEAD metric (%.4g) is NOT %s %s (%.4g)",
		e.Current, e.Operator, e.Branch, e.Base)
}

func (e *MetricTestError) ExitCode() int {
	return 1 // Expected failure
}

// ExecutionError is used for unexpected errors during execution
type ExecutionError struct {
	Phase   string // Which phase failed (setup, execution, cleanup)
	Wrapped error  // The underlying error
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("%s: %v", e.Phase, e.Wrapped)
}

func (e *ExecutionError) ExitCode() int {
	return 2 // Unexpected error
}

func (e *ExecutionError) Unwrap() error {
	return e.Wrapped
}

// ValidationError is used for configuration or input validation failures
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

func (e *ValidationError) ExitCode() int {
	return 2 // Configuration error
}

// GitError is used for Git-related failures
type GitError struct {
	Operation string
	Wrapped   error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git %s: %v", e.Operation, e.Wrapped)
}

func (e *GitError) ExitCode() int {
	return 2 // Git operation error
}

func (e *GitError) Unwrap() error {
	return e.Wrapped
}

// CommandError is used for command execution failures
type CommandError struct {
	Command string
	Wrapped error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command '%s' failed: %v", e.Command, e.Wrapped)
}

func (e *CommandError) ExitCode() int {
	return 2 // Command execution error
}

func (e *CommandError) Unwrap() error {
	return e.Wrapped
}

// ParseError is used for metric parsing failures
type ParseError struct {
	Output  string
	Wrapped error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("invalid metric output '%s': %v", e.Output, e.Wrapped)
}

func (e *ParseError) ExitCode() int {
	return 2 // Parse error
}

func (e *ParseError) Unwrap() error {
	return e.Wrapped
}
