package errors

import (
	"context"
	"errors"
	"strings"
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

// extractConciseReason extracts a concise reason from command execution errors
func extractConciseReason(err error) string {
	if err == nil {
		return "unknown error"
	}
	
	errStr := err.Error()
	
	// Handle exit code errors (most common case)
	if strings.HasPrefix(errStr, "exit status ") {
		code := strings.TrimPrefix(errStr, "exit status ")
		return "exit code " + code
	}
	
	// Handle command not found
	if strings.Contains(errStr, "executable file not found") {
		// Extract command name from error like: exec: "npm": executable file not found in $PATH
		if strings.Contains(errStr, "exec: \"") {
			start := strings.Index(errStr, "exec: \"") + 7
			end := strings.Index(errStr[start:], "\"")
			if end > 0 {
				cmd := errStr[start : start+end]
				return cmd + " not found"
			}
		}
		return "command not found"
	}
	
	// Handle process killed
	if strings.Contains(errStr, "killed") {
		return "process killed"
	}
	
	// Handle signal errors
	if strings.Contains(errStr, "signal:") {
		return "interrupted"
	}
	
	// Fallback to original error message
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
