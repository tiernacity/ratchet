package errors

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
)

// NewMetricTestError creates a metric test failure
func NewMetricTestError(current, base float64, operator, branch string) *MetricTestError {
	return &MetricTestError{
		Current:  current,
		Base:     base,
		Operator: operator,
		Branch:   branch,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// WrapGitError wraps an error with Git operation context
func WrapGitError(operation string, err error) *GitError {
	if err == nil {
		return nil
	}
	return &GitError{
		Operation: operation,
		Wrapped:   err,
	}
}

// WrapCommandError wraps an error with command execution context
func WrapCommandError(command string, err error) *CommandError {
	if err == nil {
		return nil
	}
	return &CommandError{
		Command: command,
		Wrapped: err,
	}
}

// NewCancelledError creates a cancellation error
func NewCancelledError() *CancelledError {
	return &CancelledError{}
}

// NewPhaseError creates a phase execution error with context
func NewPhaseError(phase, branch, reason string) *PhaseError {
	return &PhaseError{
		Phase:  phase,
		Branch: branch,
		Reason: reason,
	}
}

// NewPhaseErrorFromExecutorError creates a phase error from an executor error
func NewPhaseErrorFromExecutorError(phase, branch string, err error) RatchetError {
	if err == nil {
		return nil
	}

	// Handle different error types directly
	errStr := err.Error()

	// Handle cancellation - should return CancelledError, not PhaseError
	if errStr == "cancelled" || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return NewCancelledError()
	}

	// Extract concise reason from error
	reason := extractConciseReason(err)
	return NewPhaseError(phase, branch, reason)
}

// extractConciseReason extracts a concise reason from command execution errors using proper error types
func extractConciseReason(err error) string {
	if err == nil {
		return "unknown error"
	}

	// Check for exec.ExitError (command ran but failed)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Get the actual exit code from the process state
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				return "interrupted by signal " + status.Signal().String()
			}
			return "exit code " + exitErr.String()
		}
		return "exit code " + exitErr.String()
	}

	// Check for exec.Error (command couldn't be started)
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		if execErr.Err == exec.ErrNotFound {
			return execErr.Name + " not found"
		}
		return "failed to start " + execErr.Name + ": " + execErr.Err.Error()
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}

	// Check for context timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	// Check for syscall errors
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		switch syscallErr {
		case syscall.ENOENT:
			return "command not found"
		case syscall.EACCES:
			return "permission denied"
		case syscall.EINTR:
			return "interrupted"
		default:
			return "system error: " + syscallErr.Error()
		}
	}

	// Fallback to original error message (but truncate if too long)
	errStr := err.Error()
	if len(errStr) > 100 {
		return errStr[:97] + "..."
	}
	return errStr
}

// WrapParseError wraps an error with parsing context
func WrapParseError(output string, err error) *ParseError {
	if err == nil {
		return nil
	}
	return &ParseError{
		Output:  output,
		Wrapped: err,
	}
}

// IsMetricTestError checks if an error is a metric test failure
func IsMetricTestError(err error) bool {
	var mte *MetricTestError
	return errors.As(err, &mte)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// IsGitError checks if an error is a Git error
func IsGitError(err error) bool {
	var ge *GitError
	return errors.As(err, &ge)
}

// IsCommandError checks if an error is a command error
func IsCommandError(err error) bool {
	var ce *CommandError
	return errors.As(err, &ce)
}

// IsCancelledError checks if an error is a cancellation error
func IsCancelledError(err error) bool {
	var ce *CancelledError
	return errors.As(err, &ce)
}

// IsPhaseError checks if an error is a phase execution error
func IsPhaseError(err error) bool {
	var pe *PhaseError
	return errors.As(err, &pe)
}

// IsParseError checks if an error is a parse error
func IsParseError(err error) bool {
	var pe *ParseError
	return errors.As(err, &pe)
}

// GetExitCode returns the appropriate exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	var re RatchetError
	if errors.As(err, &re) {
		return re.ExitCode()
	}

	return 2 // Default to execution error
}

// ShouldSuppressHelp returns true if help should be suppressed for this error type
func ShouldSuppressHelp(err error) bool {
	if err == nil {
		return false
	}

	// Suppress help for runtime errors, but not validation/CLI errors
	return IsMetricTestError(err) ||
		IsCommandError(err) ||
		IsCancelledError(err) ||
		IsPhaseError(err) ||
		IsGitError(err) ||
		IsParseError(err)
}
