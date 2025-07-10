package errors

import (
	"errors"
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

// WrapExecutionError wraps an error with execution context
func WrapExecutionError(phase string, err error) *ExecutionError {
	if err == nil {
		return nil
	}
	return &ExecutionError{
		Phase:   phase,
		Wrapped: err,
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

// IsExecutionError checks if an error is an execution error
func IsExecutionError(err error) bool {
	var ee *ExecutionError
	return errors.As(err, &ee)
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
