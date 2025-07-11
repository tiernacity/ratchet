# Error Handling Best Practices Guide

## Critical Rule: Never Use String Matching on Errors

**IMPORTANT**: All error detection must use proper Go error types with `errors.Is()` and `errors.As()`. String matching on error messages is fragile and will break.

## Why String Matching Is Forbidden

### Problems with String-Based Error Detection

```go
// ❌ WRONG - Fragile and unreliable
func badErrorHandling(err error) {
    errStr := err.Error()
    
    if strings.Contains(errStr, "exit status") {
        // PROBLEM: This breaks with different Go versions
    }
    if strings.Contains(errStr, "executable file not found") {
        // PROBLEM: This breaks with different locales
    }
    if strings.Contains(errStr, "killed") {
        // PROBLEM: This breaks with different OS versions
    }
}
```

### Issues with String Matching:

1. **Locale Dependency**: Error messages change with system language
2. **Version Dependency**: Error formats change across Go versions  
3. **OS Dependency**: Different operating systems produce different messages
4. **Brittleness**: Any change in error message format breaks the code
5. **False Positives**: String fragments may match unintended errors

## Correct Approach: Typed Error Detection

### ✅ Standard Go Error Types

```go
import (
    "errors"
    "os/exec"
    "syscall"
    "context"
)

func handleCommandError(err error) string {
    // Check for exec.ExitError (command ran but failed)
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        return fmt.Sprintf("command failed with exit code %d", exitErr.ExitCode())
    }
    
    // Check for exec.Error (command couldn't be started)
    var execErr *exec.Error
    if errors.As(err, &execErr) {
        if execErr.Err == exec.ErrNotFound {
            return fmt.Sprintf("command '%s' not found", execErr.Name)
        }
        return fmt.Sprintf("failed to start '%s': %v", execErr.Name, execErr.Err)
    }
    
    // Check for context cancellation
    if errors.Is(err, context.Canceled) {
        return "operation cancelled"
    }
    
    // Check for context timeout
    if errors.Is(err, context.DeadlineExceeded) {
        return "operation timed out"
    }
    
    // Check for syscall errors
    var errno syscall.Errno
    if errors.As(err, &errno) {
        switch errno {
        case syscall.ENOENT:
            return "file or command not found"
        case syscall.EACCES:
            return "permission denied"
        case syscall.EINTR:
            return "operation interrupted"
        default:
            return fmt.Sprintf("system error: %v", errno)
        }
    }
    
    return err.Error()
}
```

### ✅ Custom Ratchet Error Types

```go
func handleRatchetError(err error) int {
    // Check for metric test failure (expected)
    var metricErr *errors.MetricTestError
    if errors.As(err, &metricErr) {
        fmt.Fprintf(os.Stderr, "HEAD metric (%.4g) is NOT %s %s (%.4g)\n",
            metricErr.Current, metricErr.Operator, metricErr.Branch, metricErr.Base)
        return 1 // Expected failure
    }
    
    // Check for phase execution error
    var phaseErr *errors.PhaseError
    if errors.As(err, &phaseErr) {
        fmt.Fprintf(os.Stderr, "%s in %s: %s\n", 
            phaseErr.Phase, phaseErr.Branch, phaseErr.Reason)
        return 2 // Execution error
    }
    
    // Check for Git operation error
    var gitErr *errors.GitError
    if errors.As(err, &gitErr) {
        fmt.Fprintf(os.Stderr, "Git %s: %v\n", gitErr.Operation, gitErr.Wrapped)
        return 2 // Execution error
    }
    
    // Check for validation error
    var validationErr *errors.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Fprintf(os.Stderr, "%s: %s\n", validationErr.Field, validationErr.Message)
        return 2 // Configuration error
    }
    
    // Check for parse error
    var parseErr *errors.ParseError
    if errors.As(err, &parseErr) {
        fmt.Fprintf(os.Stderr, "Parse error: %v\n", parseErr.Wrapped)
        return 2 // Execution error
    }
    
    // Default handling for unknown errors
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    return 2
}
```

## Testing Error Handling

### ✅ Correct Test Patterns

```go
func TestCommandExecution(t *testing.T) {
    // Use actual error types in tests
    err := someFunction()
    
    // Check for specific error types
    var exitErr *exec.ExitError
    require.True(t, errors.As(err, &exitErr), "Expected exec.ExitError")
    assert.Equal(t, 1, exitErr.ExitCode())
    
    // Or check for error interfaces
    var ratchetErr errors.RatchetError
    if errors.As(err, &ratchetErr) {
        assert.Equal(t, 2, ratchetErr.ExitCode())
    }
}
```

### ❌ Wrong Test Patterns

```go
func TestCommandExecution(t *testing.T) {
    err := someFunction()
    
    // DON'T DO THIS - fragile string matching
    assert.Contains(t, err.Error(), "exit status")
    assert.Contains(t, err.Error(), "command not found")
}
```

## Command Safety Checking

### ✅ Proper Command Parsing

```go
func isDangerousCommand(cmd string) bool {
    if cmd == "" {
        return false
    }
    
    // Parse command properly
    parts := strings.Fields(cmd)
    if len(parts) == 0 {
        return false
    }
    
    // Check the actual executable name
    executable := filepath.Base(parts[0])
    
    dangerousCommands := map[string]bool{
        "rm":     true,
        "rmdir":  true,
        "del":    true,  // Windows
        "format": true,  // Windows
        "dd":     true,
        "sudo":   true,
    }
    
    return dangerousCommands[executable]
}
```

### ❌ Wrong Approach

```go
// DON'T DO THIS - fragile string matching
func isDangerousCommand(cmd string) bool {
    return strings.Contains(cmd, "rm ") ||
           strings.Contains(cmd, "delete") ||
           strings.Contains(cmd, "format")
}
```

## Implementation Rules

### DO Use:
- `errors.As()` for error type checking
- `errors.Is()` for error value checking  
- Proper Go error types (`exec.ExitError`, `exec.Error`, etc.)
- Custom structured error types with fields
- Exit codes from `exec.ExitError.ExitCode()`
- Syscall error constants (`syscall.ENOENT`, etc.)

### DON'T Use:
- `strings.Contains()` on error messages
- `strings.HasPrefix()` on error messages  
- Regular expressions on error text
- Parsing error strings to extract information
- String matching on command output for error detection

## Migration Guide

When updating existing code:

1. **Identify String Matching**: Find all uses of `strings.Contains()`, `strings.HasPrefix()`, etc. on error messages
2. **Determine Error Source**: Identify what type of error is being detected
3. **Use Proper Types**: Replace with appropriate `errors.As()` or `errors.Is()` checks
4. **Update Tests**: Replace string-based assertions with type-based checks
5. **Verify Robustness**: Test with different Go versions, locales, and operating systems

## Examples in Codebase

See these files for correct error handling patterns:
- `internal/errors/helpers.go` - Proper error type detection
- `internal/executor/executor.go` - Raw error passthrough
- `internal/orchestrator/orchestrator.go` - Error type handling
- `internal/config/types.go` - Robust command validation

Remember: Error handling must be robust, type-safe, and immune to changes in error message formatting.