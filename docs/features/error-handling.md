# Error Handling Feature

## Overview
The Error Handling module defines custom error types that map to specific exit codes and provide consistent error reporting throughout the application. It distinguishes between expected failures (metric tests) and unexpected errors, and controls when help text is displayed to users.

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
    Branch   string
}

func (e *MetricTestError) Error() string {
    return fmt.Sprintf("HEAD metric (%.4g) is NOT %s %s (%.4g)", 
        e.Current, e.Operator, e.Branch, e.Base)
}

func (e *MetricTestError) ExitCode() int {
    return 1 // Expected failure
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

### GitError
Used for Git-related failures.

```go
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
```

### CommandError
Used for command execution failures.

```go
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
```

### ParseError
Used for metric parsing failures.

```go
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
```

## Exit Codes

| Code | Type | Description | Example |
|------|------|-------------|---------|
| 0 | Success | Metric test passed | Current metric satisfies comparison |
| 1 | MetricTestError | Metric test failed (expected) | Current < Base when --gt specified |
| 2 | ValidationError | Invalid configuration/input | Multiple operators specified |
| 2 | GitError | Git operation failed | Branch not found |
| 2 | CommandError | Command execution failed | Command returned non-zero exit |
| 2 | ParseError | Metric parsing failed | No numeric value in output |

## Error Creation Helpers

The errors package provides two patterns for creating errors:

### New* Functions - Creating Original Errors
```go
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
```

### Wrap* Functions - Wrapping Existing Errors
```go
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
```

## Error Checking Utilities

```go
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
        IsGitError(err) ||
        IsParseError(err)
}
```

## Help Display Behavior

The CLI uses `ShouldSuppressHelp()` to determine when to show help text:

### Shows Help (User Error)
- **ValidationError**: Configuration/input validation failures
- **CLI errors**: Missing arguments, unknown flags, flag parsing errors

### Suppresses Help (Runtime Error)
- **MetricTestError**: Expected test failures
- **GitError**: Git operation failures
- **CommandError**: Command execution failures
- **ParseError**: Metric parsing failures

## Usage in Main

```go
func runRatchet(cmd *cobra.Command, args []string) error {
    // ... implementation ...
    
    err = orch.Run(ctx, cfg)
    if err != nil && errors.ShouldSuppressHelp(err) {
        cmd.SilenceUsage = true
    }
    return err
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        // Exit with appropriate code (Cobra handles error printing)
        os.Exit(errors.GetExitCode(err))
    }
}
```

## Error Context and Wrapping

### Adding Context in Orchestrator
```go
// In orchestrator
func (o *orchestrator) Run(ctx context.Context, cfg *config.Config) error {
    // Check Git repository
    if err := o.git.IsGitRepository(); err != nil {
        return errors.WrapGitError("validation", fmt.Errorf("not in a git repository"))
    }
    
    // Resolve branch
    baseBranch, err := o.git.ResolveBranch(cfg.BaseBranch)
    if err != nil {
        return errors.WrapGitError("branch resolution",
            fmt.Errorf("branch '%s' not found", cfg.BaseBranch))
    }
    
    // Create worktree
    if err := o.git.CreateWorktree(worktreePath, baseBranch); err != nil {
        return errors.WrapGitError("worktree creation",
            fmt.Errorf("failed to create worktree for %s: %w", baseBranch, err))
    }
    
    // Compare metrics
    return o.compareMetrics(baseMetric, headMetric, cfg.Operator, baseBranch)
}

// compareMetrics returns appropriate error type
func (o *orchestrator) compareMetrics(baseMetric, headMetric float64, op config.ComparisonOperator, baseBranch string) error {
    passed := false
    
    switch op {
    case config.OpGreaterThan:
        passed = headMetric > baseMetric
    // ... other cases ...
    }
    
    if !passed {
        return errors.NewMetricTestError(headMetric, baseMetric, op.HumanString(), baseBranch)
    }
    
    o.reporter.Success(headMetric, baseMetric, op.HumanString(), baseBranch)
    return nil
}
```

### Error Chain Examples
```
git validation: not in a git repository
git branch resolution: branch 'feature-xyz' not found
command 'npm test' failed: command failed with exit code 1: npm: command not found
invalid metric output 'no tests found': no numeric value found
validation error: operators: multiple comparison operators specified
```

## Testing Error Handling

### Unit Tests
```go
func TestErrorCreationHelpers(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        wantCode int
        wantMsg  string
    }{
        {
            name:     "metric test error",
            err:      NewMetricTestError(5.0, 10.0, "greater than", "main"),
            wantCode: 1,
            wantMsg:  "HEAD metric (5) is NOT greater than main (10)",
        },
        {
            name:     "validation error",
            err:      NewValidationError("metric-cmd", "cannot be empty"),
            wantCode: 2,
            wantMsg:  "validation error: metric-cmd: cannot be empty",
        },
        {
            name:     "git error",
            err:      WrapGitError("checkout", fmt.Errorf("branch not found")),
            wantCode: 2,
            wantMsg:  "git checkout: branch not found",
        },
        {
            name:     "command error",
            err:      WrapCommandError("echo test", fmt.Errorf("exit status 1")),
            wantCode: 2,
            wantMsg:  "command 'echo test' failed: exit status 1",
        },
        {
            name:     "parse error",
            err:      WrapParseError("hello", fmt.Errorf("not a number")),
            wantCode: 2,
            wantMsg:  "invalid metric output 'hello': not a number",
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
    metricErr := NewMetricTestError(1, 2, "greater than", "main")
    validErr := NewValidationError("field", "message")
    gitErr := WrapGitError("test", fmt.Errorf("failed"))
    cmdErr := WrapCommandError("test", fmt.Errorf("failed"))
    parseErr := WrapParseError("test", fmt.Errorf("failed"))
    
    // Test each type detection
    assert.True(t, IsMetricTestError(metricErr))
    assert.False(t, IsMetricTestError(validErr))
    
    assert.True(t, IsValidationError(validErr))
    assert.False(t, IsValidationError(metricErr))
    
    // ... etc
}
```

### Help Suppression Testing
```go
func TestShouldSuppressHelp(t *testing.T) {
    // Should suppress help for runtime errors
    assert.True(t, ShouldSuppressHelp(NewMetricTestError(1, 2, ">", "main")))
    assert.True(t, ShouldSuppressHelp(WrapGitError("test", fmt.Errorf("failed"))))
    assert.True(t, ShouldSuppressHelp(WrapCommandError("test", fmt.Errorf("failed"))))
    assert.True(t, ShouldSuppressHelp(WrapParseError("test", fmt.Errorf("failed"))))
    
    // Should NOT suppress help for validation errors
    assert.False(t, ShouldSuppressHelp(NewValidationError("field", "message")))
    assert.False(t, ShouldSuppressHelp(nil))
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
"command 'npm test' failed: command failed with exit code 1"
"git worktree creation: failed to create worktree for main: already checked out"
"validation error: operators: multiple comparison operators specified"

// Bad
"Error!"
"Failed"
"Internal error: panic in goroutine 7"
```

## Concise Error Message Specification

### Overview
Ratchet should provide clear, concise error messages that help users understand what went wrong without overwhelming them with technical details. The error messages should be formatted consistently and avoid redundancy.

### Message Format Requirements

#### General Format
- **Single line**: Error messages should be a single, concise statement
- **No trailing punctuation**: Don't end messages with periods, colons, or other punctuation
- **Lowercase start**: Begin with lowercase unless it's a proper noun
- **Context over detail**: Provide enough context to understand what failed, but avoid technical stack traces

#### Context Priority
Error messages should include context in this priority order:
1. **What failed** (the action being performed)
2. **Where it failed** (branch name, command, etc.)
3. **Why it failed** (specific reason)

#### Redundancy Elimination
- **Avoid repeating "command"**: Don't say "command 'X' failed: command execution failed"
- **Avoid error chains**: Don't show multiple layers of error wrapping
- **Consolidate context**: Include branch/phase information naturally, not as separate clauses

### Error Message Categories

#### 1. Command Execution Errors
**Format**: `{phase} in {branch}: {reason}`

**Examples**:
```
# Current (bad)
"command 'npm test' failed: metric command failed in origin/main: command failed with exit code 1:"

# Improved (good)
"metric command in origin/main: exit code 1"
```

#### 2. Command Not Found
**Format**: `{phase} in {branch}: {command} not found`

**Examples**:
```
# Current (bad)
"command 'npm' failed: pre-command failed in origin/main: command execution failed: exec: \"npm\": executable file not found in $PATH"

# Improved (good)
"pre-command in origin/main: npm not found"
```

#### 3. Cancellation
**Format**: `cancelled`

**Examples**:
```
# Current (bad)
"command 'sleep 3' failed: pre-command failed in origin/main: cancelled"

# Improved (good)
"cancelled"
```

#### 4. Git Errors
**Format**: `{operation}: {reason}`

**Examples**:
```
# Current (bad)
"git branch resolution: branch 'feature-xyz' not found"

# Improved (good)
"branch 'feature-xyz' not found"
```

#### 5. Validation Errors
**Format**: `{field}: {problem}`

**Examples**:
```
# Current (good, keep as is)
"metric-cmd: cannot be empty"
"operators: multiple comparison operators specified"
```

#### 6. Metric Test Failures
**Format**: `HEAD metric ({value}) is not {operator} {branch} ({value})`

**Examples**:
```
# Current (good, keep as is)
"HEAD metric (42) is not greater than main (50)"
```

#### 7. Parse Errors
**Format**: `invalid metric output: {reason}`

**Examples**:
```
# Current (bad)
"invalid metric output 'hello world': no numeric value found"

# Improved (good)
"invalid metric output: no numeric value found"
```

### Implementation Strategy

#### Phase 1: Simplify Command Errors
- Modify `CommandError.Error()` to use context-aware formatting
- Detect the command phase (pre-command, metric, post-command) from the orchestrator
- Include branch information naturally in the message

#### Phase 2: Eliminate Error Wrapping Chains
- Prevent multiple layers of error wrapping from appearing in user messages
- Use the outermost error type to determine the final message format
- Preserve detailed error information for debugging but hide from users

#### Phase 3: Context-Aware Formatting
- Pass context (branch, phase) to error creation functions
- Format messages based on the specific error scenario
- Maintain consistency across different error types

### Error Message Examples

#### Command Execution Scenarios
```bash
# Pre-command failure
./ratchet --gt main 'echo 42' --pre 'false'
# Current: Error: command 'false' failed: pre-command failed in main: command failed with exit code 1:
# Improved: Error: pre-command in main: exit code 1

# Metric command failure  
./ratchet --gt main 'false'
# Current: Error: command 'false' failed: metric command failed in main: command failed with exit code 1:
# Improved: Error: metric command in main: exit code 1

# Command not found
./ratchet --gt main 'echo 42' --pre 'nonexistent'
# Current: Error: command 'nonexistent' failed: pre-command failed in main: command execution failed: exec: "nonexistent": executable file not found in $PATH
# Improved: Error: pre-command in main: nonexistent not found

# Cancellation
./ratchet --gt main 'echo 42' --pre 'sleep 10'  # ^C
# Current: Error: command 'sleep 10' failed: pre-command failed in main: cancelled
# Improved: Error: cancelled
```

#### Other Error Scenarios
```bash
# Git errors
./ratchet --gt nonexistent 'echo 42'
# Current: Error: git branch resolution: branch 'nonexistent' not found
# Improved: Error: branch 'nonexistent' not found

# Validation errors (keep current format)
./ratchet --gt --lt main 'echo 42'
# Current: Error: validation error: operators: multiple comparison operators specified
# Keep as: Error: operators: multiple comparison operators specified

# Parse errors
./ratchet --gt main 'echo hello'
# Current: Error: invalid metric output 'hello': no numeric value found
# Improved: Error: invalid metric output: no numeric value found
```

### Testing Requirements

#### Unit Tests
- Test each error message format with expected inputs
- Verify no trailing punctuation or redundant text
- Test context propagation (branch names, command phases)

#### Integration Tests
- Test end-to-end error scenarios with actual command execution
- Verify error messages match specification
- Test cancellation scenarios

#### Documentation Tests
- Ensure examples in documentation match actual error output
- Update help text and error message examples when formats change

## Common Error Scenarios

### Configuration Errors
```go
// Multiple operators
if count > 1 {
    return errors.NewValidationError("operators", "multiple comparison operators specified")
}

// Missing required field
if c.MetricCmd == "" {
    return errors.NewValidationError("metric-cmd", "metric command is required")
}

// Invalid config file
if err := v.ReadInConfig(); err != nil {
    return errors.NewValidationError("config-file", 
        fmt.Sprintf("failed to read config file %s: %v", cfgFile, err))
}
```

### Git Errors
```go
// Not in a git repository
if err := o.git.IsGitRepository(); err != nil {
    return errors.WrapGitError("validation", fmt.Errorf("not in a git repository"))
}

// Branch doesn't exist
baseBranch, err := o.git.ResolveBranch(cfg.BaseBranch)
if err != nil {
    return errors.WrapGitError("branch resolution",
        fmt.Errorf("branch '%s' not found", cfg.BaseBranch))
}
```

### Command Errors
```go
// Command execution with context
output, err := o.executor.Execute(ctx, dir, cfg.MetricCmd)
if err != nil {
    return errors.WrapCommandError(cfg.MetricCmd,
        fmt.Errorf("metric command failed in %s: %w", branchName, err))
}
```

### Parse Errors
```go
// Invalid metric output
metric, err := o.parser.Parse(output)
if err != nil {
    return err // parser already returns ParseError
}
```

## Future Enhancements

1. **Structured Errors**: Include machine-readable error codes
2. **Error Recovery**: Suggestions for automatic recovery
3. **Error Reporting**: Send errors to monitoring service
4. **Localization**: Support error messages in multiple languages
5. **Error Templates**: Reusable error message templates