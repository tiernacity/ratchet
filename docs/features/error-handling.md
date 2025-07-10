# Error Handling Feature

## Overview
The Error Handling module defines custom error types that map to specific exit codes and provide consistent error reporting throughout the application. It distinguishes between expected failures (metric tests) and unexpected errors.

## Error Types

### Interface
```go
package errors

// RatchetError is the base interface for all ratchet errors
type RatchetError interface {
    error
    ExitCode() int
}
```

### MetricTestError
Used when a metric comparison fails - this is an expected failure scenario.

```go
type MetricTestError struct {
    Current  float64
    Base     float64
    Operator string
}

func (e *MetricTestError) Error() string {
    return fmt.Sprintf("metric test failed: %.2f %s %.2f", 
        e.Current, e.Operator, e.Base)
}

func (e *MetricTestError) ExitCode() int {
    return 1 // Expected failure
}
```

### ExecutionError
Used for unexpected errors during execution.

```go
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
```

### ValidationError
Used for configuration or input validation failures.

```go
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
```

## Exit Codes

| Code | Type | Description | Example |
|------|------|-------------|---------|
| 0 | Success | Metric test passed | Current metric satisfies comparison |
| 1 | MetricTestError | Metric test failed (expected) | Current < Base when --gt specified |
| 2 | ExecutionError | Unexpected error | Command not found, Git error, etc. |

## Error Creation Helpers

```go
// NewMetricTestError creates a metric test failure
func NewMetricTestError(current, base float64, operator string) *MetricTestError {
    return &MetricTestError{
        Current:  current,
        Base:     base,
        Operator: operator,
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
```

## Error Checking Utilities

```go
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
```

## Usage in Main

```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        // Log error to stderr
        fmt.Fprintln(os.Stderr, err)
        
        // Exit with appropriate code
        os.Exit(errors.GetExitCode(err))
    }
}
```

## Error Context and Wrapping

### Adding Context
```go
// In orchestrator
func (o *orchestrator) Run(ctx context.Context, cfg *config.Config) error {
    // Check Git repository
    if err := o.git.IsGitRepository(); err != nil {
        return errors.WrapExecutionError("git validation", err)
    }
    
    // Create worktree
    if err := o.git.CreateWorktree(path, cfg.BaseBranch); err != nil {
        return errors.WrapExecutionError("worktree creation", 
            fmt.Errorf("failed to create worktree for %s: %w", cfg.BaseBranch, err))
    }
    
    // Compare metrics
    if !o.compareMetrics(current, base, cfg.Operator) {
        return errors.NewMetricTestError(current, base, cfg.Operator.String())
    }
}
```

### Error Chain Example
```
ExecutionError: worktree creation: failed to create worktree for main: exit status 128: fatal: 'main' is already checked out at '/home/user/project'
```

## Testing Error Handling

### Unit Tests
```go
func TestErrorTypes(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        wantCode int
        wantMsg  string
    }{
        {
            name:     "metric test error",
            err:      NewMetricTestError(5.0, 10.0, "<"),
            wantCode: 1,
            wantMsg:  "metric test failed: 5.00 < 10.00",
        },
        {
            name:     "execution error",
            err:      WrapExecutionError("setup", fmt.Errorf("command failed")),
            wantCode: 2,
            wantMsg:  "setup: command failed",
        },
        {
            name:     "validation error",
            err:      NewValidationError("metric-cmd", "cannot be empty"),
            wantCode: 2,
            wantMsg:  "validation error: metric-cmd: cannot be empty",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.wantMsg, tt.err.Error())
            assert.Equal(t, tt.wantCode, GetExitCode(tt.err))
        })
    }
}
```

### Error Type Checking
```go
func TestErrorTypeChecking(t *testing.T) {
    metricErr := NewMetricTestError(1, 2, ">")
    execErr := WrapExecutionError("test", fmt.Errorf("failed"))
    
    assert.True(t, IsMetricTestError(metricErr))
    assert.False(t, IsMetricTestError(execErr))
    
    assert.True(t, IsExecutionError(execErr))
    assert.False(t, IsExecutionError(metricErr))
}
```

## Error Messages Best Practices

### Do's
1. **Be Specific**: Include what failed and why
2. **Add Context**: Include relevant values (branch names, commands, etc.)
3. **Be Actionable**: Suggest how to fix the issue when possible
4. **Use Lowercase**: Start error messages with lowercase
5. **No Punctuation**: Don't end with periods

### Don'ts
1. **Generic Messages**: Avoid "something went wrong"
2. **Technical Jargon**: Keep messages user-friendly
3. **Stack Traces**: Don't expose internal implementation
4. **Sensitive Data**: Never include passwords or tokens

### Examples
```go
// Good
"metric command failed: exit status 1: npm: command not found"
"failed to create worktree: branch 'feature-xyz' does not exist"
"validation error: exactly one comparison operator must be specified"

// Bad
"Error!"
"Failed"
"Internal error: panic in goroutine 7"
```

## Integration with Progress Reporter

```go
func (o *orchestrator) handleError(err error, reporter progress.Reporter) {
    if err == nil {
        return
    }
    
    // Show error in progress
    reporter.Error(err.Error())
    
    // Provide additional context for common errors
    switch {
    case strings.Contains(err.Error(), "not in a git repository"):
        reporter.Info("Run 'git init' to initialize a repository")
    case strings.Contains(err.Error(), "command not found"):
        reporter.Info("Ensure the command is installed and in PATH")
    }
}
```

## Common Error Scenarios

### Git Errors
```go
// Not in a git repository
if err := git.IsGitRepository(); err != nil {
    return WrapExecutionError("git validation", 
        fmt.Errorf("not in a git repository. Run 'git init' first"))
}

// Branch doesn't exist
if err := git.CreateWorktree(path, branch); err != nil {
    return WrapExecutionError("worktree creation",
        fmt.Errorf("branch '%s' not found. Did you mean 'origin/%s'?", branch, branch))
}
```

### Command Errors
```go
// Command not found
if err := executor.Execute(ctx, dir, cmd); err != nil {
    if strings.Contains(err.Error(), "executable file not found") {
        return WrapExecutionError("command execution",
            fmt.Errorf("command not found: %s. Is it installed?", cmd))
    }
    return WrapExecutionError("command execution", err)
}
```

### Parse Errors
```go
// Invalid metric output
metric, err := parser.Parse(output)
if err != nil {
    return WrapExecutionError("metric parsing",
        fmt.Errorf("invalid metric output '%s': %w", output, err))
}
```

## Future Enhancements

1. **Structured Errors**: Include machine-readable error codes
2. **Error Recovery**: Suggestions for automatic recovery
3. **Error Reporting**: Send errors to monitoring service
4. **Localization**: Support error messages in multiple languages
5. **Error Templates**: Reusable error message templates