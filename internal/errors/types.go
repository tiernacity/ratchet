package errors

import (
	"fmt"
	"strings"
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

// ValidationError is used for configuration or input validation failures
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
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
	return fmt.Sprintf("%v", e.Wrapped)
}

func (e *GitError) ExitCode() int {
	return 2 // Git operation error
}

func (e *GitError) Unwrap() error {
	return e.Wrapped
}

// CancelledError represents a cancelled operation
type CancelledError struct{}

func (e *CancelledError) Error() string {
	return "cancelled"
}

func (e *CancelledError) ExitCode() int {
	return 2 // Execution error
}

// PhaseError represents a command execution failure in a specific phase and branch
type PhaseError struct {
	Phase   string // "pre", "metric", "post"
	Branch  string // branch name
	Command string // the command that failed
	Reason  string // concise reason for failure
}

func (e *PhaseError) Error() string {
	return fmt.Sprintf("%s '%s' %s", e.Phase, e.Command, e.Reason)
}

func (e *PhaseError) ExitCode() int {
	return 2 // Command execution error
}

// CommandError is used for backward compatibility with legacy command execution failures
type CommandError struct {
	Command string
	Wrapped error
}

func (e *CommandError) Error() string {
	// Check if this is a cancellation error and provide cleaner message
	if e.Wrapped != nil && e.Wrapped.Error() == "cancelled" {
		return "cancelled"
	}
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
	cleanOutput := strings.TrimSpace(e.Output)
	if cleanOutput == "" {
		return "no metric output provided"
	}
	return fmt.Sprintf("invalid metric output '%s': %v", cleanOutput, e.Wrapped)
}

func (e *ParseError) ExitCode() int {
	return 2 // Parse error
}

func (e *ParseError) Unwrap() error {
	return e.Wrapped
}

// ExitError is used to return a specific exit code without showing an error message
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return "" // Empty message to prevent double output
}

func (e *ExitError) ExitCode() int {
	return e.Code
}
