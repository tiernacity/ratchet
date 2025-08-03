# Design Patterns and Best Practices

This document captures the essential design decisions, error handling patterns, and architectural principles for ongoing development of the ratchet CLI tool.

## Core Error Handling Principles

**CRITICAL**: Always use proper Go error types with `errors.Is()` and `errors.As()`. Never use string matching on error messages.

### Error Type Strategy

All ratchet errors implement the `RatchetError` interface:

```go
type RatchetError interface {
    error
    ExitCode() int
}
```

### Exit Code Mapping

| Code | Type | Description | Example |
|------|------|-------------|---------|
| 0 | Success | Metric test passed | Current metric satisfies comparison |
| 1 | MetricTestError | Metric test failed (expected) | Current < Base when --gt specified |
| 2 | All other errors | Configuration, Git, Command, Parse errors | Invalid flags, branch not found, command failed |

### Error Types and Usage

#### MetricTestError - Expected Failures
```go
type MetricTestError struct {
    Current  float64
    Base     float64
    Operator string
    Branch   string
}

// Usage
return errors.NewMetricTestError(headMetric, baseMetric, op.HumanString(), baseBranch)
```

#### ValidationError - Configuration Issues
```go
type ValidationError struct {
    Field   string
    Message string
}

// Usage
return errors.NewValidationError("operators", "multiple comparison operators specified")
```

#### GitError - Git Operation Failures
```go
type GitError struct {
    Operation string
    Wrapped   error
}

// Usage
return errors.WrapGitError("branch resolution", fmt.Errorf("branch '%s' not found", cfg.BaseBranch))
```

#### CommandError - Command Execution Failures
```go
type CommandError struct {
    Command string
    Wrapped error
}

// Usage
return errors.WrapCommandError(cfg.MetricCmd, fmt.Errorf("metric command failed in %s: %w", branchName, err))
```

#### ParseError - Metric Parsing Failures
```go
type ParseError struct {
    Output  string
    Wrapped error
}

// Usage (typically from parser)
return errors.WrapParseError(output, fmt.Errorf("no numeric value found"))
```

### Proper Error Detection Patterns

#### ✅ Correct Approach - Use Error Types
```go
func handleRatchetError(err error) int {
    var metricErr *errors.MetricTestError
    if errors.As(err, &metricErr) {
        return 1 // Expected failure
    }
    
    var validationErr *errors.ValidationError
    if errors.As(err, &validationErr) {
        return 2 // Configuration error
    }
    
    // Check standard library error types
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        return 2 // Command execution error
    }
    
    return 2 // Default
}
```

#### ❌ Wrong Approach - String Matching
```go
// NEVER DO THIS - fragile and unreliable
func badHandleError(err error) {
    if strings.Contains(err.Error(), "exit status") {
        // Breaks with locale/version changes
    }
}
```

### Help Display Strategy

Use `errors.ShouldSuppressHelp()` to control when CLI help is shown:

- **Show Help**: ValidationError, CLI parsing errors (user configuration mistakes)
- **Suppress Help**: MetricTestError, GitError, CommandError, ParseError (runtime failures)

## Orchestration Patterns

### Resource Management

Always use defer for cleanup and handle interruption signals:

```go
func (o *orchestrator) Run(ctx context.Context, cfg *config.Config) error {
    // Setup with cleanup
    cleanup, err := o.setupEnvironment(ctx, cfg)
    if err != nil {
        return err
    }
    defer cleanup() // CRITICAL: Always cleanup
    
    // Setup signal handling
    o.setupSignalHandler(cleanup)
    
    // Execute workflow...
}
```

### Worktree Management

- Use unique temporary directory names to avoid conflicts
- Always use `--force` flag for robust Git operations
- Clean up in ALL error conditions (use defer immediately after creation)

```go
func (o *orchestrator) setupEnvironment() (func(), error) {
    // Create temp directory
    tempDir, err := o.createTempDir()
    if err != nil {
        return nil, err
    }
    
    // Create worktree
    worktreePath := filepath.Join(tempDir, "ratchet-base")
    if err := o.git.CreateWorktree(worktreePath, cfg.BaseBranch); err != nil {
        os.RemoveAll(tempDir) // Cleanup on error
        return nil, errors.WrapGitError("worktree creation", err)
    }
    
    // Return cleanup function
    cleanup := func() {
        o.git.RemoveWorktree(worktreePath)
        os.RemoveAll(tempDir)
    }
    
    return cleanup, nil
}
```

### Dependency Injection

All components receive dependencies through interfaces to enable testing:

```go
type orchestrator struct {
    git      GitOperations      // Interface, not concrete type
    executor CommandExecutor    // Interface, not concrete type
    parser   MetricParser       // Interface, not concrete type
    progress ProgressReporter   // Interface, not concrete type
}
```

### Error Context Propagation

Add meaningful context when wrapping errors:

```go
// Good - includes what failed and why
return errors.WrapGitError("worktree creation", 
    fmt.Errorf("failed to create worktree for %s: %w", baseBranch, err))

// Good - includes command and phase context
return errors.WrapCommandError(cfg.MetricCmd,
    fmt.Errorf("metric command failed in %s: %w", branchName, err))
```

## Configuration Patterns

### Field Naming Convention

- CLI flags use kebab-case: `--metric`, `--pre`, `--post`
- Config file fields use snake_case: `metric:`, `pre:`, `post:`
- Go struct fields use PascalCase: `Metric`, `Pre`, `Post`

### Validation Strategy

Validate configuration at the boundary (CLI/config file parsing):

```go
func (c *Config) Validate() error {
    // Check required fields
    if c.Metric == "" {
        return errors.NewValidationError("metric", "metric command is required")
    }
    
    // Check exclusive options
    operators := []bool{c.GreaterThan, c.LessThan, c.Equal, c.GreaterEqual, c.LessEqual}
    count := 0
    for _, op := range operators {
        if op {
            count++
        }
    }
    if count != 1 {
        return errors.NewValidationError("operators", "exactly one comparison operator required")
    }
    
    return nil
}
```

## Interface Design Principles

### Minimal Interfaces

Keep interfaces focused on single responsibilities:

```go
// Good - focused responsibility
type GitOperations interface {
    IsGitRepository() error
    ResolveBranch(branch string) (string, error)
    CreateWorktree(path, branch string) error
    RemoveWorktree(path string) error
}

// Good - single responsibility
type CommandExecutor interface {
    Execute(ctx context.Context, dir, command string) (stdout, stderr string, err error)
}
```

### Error Handling in Interfaces

Interfaces should return structured errors, not strings:

```go
// Good - returns structured error
type MetricParser interface {
    Parse(output string) (float64, error) // Returns ParseError on failure
}

// Bad - returns string error messages
type BadParser interface {
    Parse(output string) (float64, string)
}
```

## Testing Patterns

### Table-Driven Tests

Use table-driven tests for validation and error scenarios:

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr error
    }{
        {
            name:    "missing metric command",
            config:  Config{},
            wantErr: errors.NewValidationError("metric", "metric command is required"),
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            assert.Equal(t, tt.wantErr, err)
        })
    }
}
```

### Mock Interfaces

Use testify/mock for interface testing:

```go
type mockGitOperations struct {
    mock.Mock
}

func (m *mockGitOperations) CreateWorktree(path, branch string) error {
    args := m.Called(path, branch)
    return args.Error(0)
}

func TestOrchestrator(t *testing.T) {
    gitMock := &mockGitOperations{}
    gitMock.On("CreateWorktree", mock.Anything, "main").Return(nil)
    
    orch := NewOrchestrator(gitMock, ...)
    // ... test orchestrator
    
    gitMock.AssertExpectations(t)
}
```

### Error Type Testing

Test error types using `errors.As()`:

```go
func TestErrorTypes(t *testing.T) {
    err := someFunction()
    
    var metricErr *errors.MetricTestError
    require.True(t, errors.As(err, &metricErr))
    assert.Equal(t, 1, metricErr.ExitCode())
}
```

## Platform Considerations

### Temporary Directories

Respect CI environment variables:

```go
func createTempDir() (string, error) {
    tempBase := os.TempDir()
    if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
        tempBase = runnerTemp
    }
    return os.MkdirTemp(tempBase, "ratchet-*")
}
```

### Signal Handling

Handle interruption for cleanup:

```go
func setupSignalHandler(cleanup func()) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        cleanup()
        os.Exit(2)
    }()
}
```

## Code Organization Principles

### Module Boundaries

- `internal/cli/`: Thin CLI layer (argument parsing only)
- `internal/orchestrator/`: Core workflow coordination
- `internal/config/`: Configuration types and validation
- `internal/errors/`: Custom error types and helpers
- `internal/git/`: Git operations interface and implementation
- `internal/executor/`: Command execution interface and implementation
- `internal/parser/`: Metric parsing and validation
- `internal/progress/`: Progress reporting interface and implementation

### Import Organization

- Standard library imports first
- Third-party imports second
- Internal imports last

### Error Propagation

- Errors flow up to main() for consistent exit code handling
- Each layer adds appropriate context through error wrapping
- Main() determines exit codes using error types

This document should be updated when core patterns change, but implementation details that can be found in the code should not be duplicated here.