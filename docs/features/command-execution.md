# Command Execution Feature

## Overview
The Command Execution module provides a platform-agnostic way to execute shell commands with proper context handling, output capture, and error reporting.

## Interface

```go
package executor

type Executor interface {
    // Execute runs a command in the specified directory and returns stdout, stderr, and any error
    Execute(ctx context.Context, dir, command string) (stdout, stderr string, err error)
}

type Options struct {
    Timeout time.Duration // Optional timeout for command execution
    Env     []string      // Additional environment variables
}
```

## Implementation Details

### Structure
```go
type executorImpl struct {
    shell string // Shell to use (bash, sh, cmd.exe)
}

func New() Executor {
    return &executorImpl{
        shell: detectShell(),
    }
}
```

### Shell Detection
```go
func detectShell() string {
    if runtime.GOOS == "windows" {
        return "cmd.exe"
    }
    
    // Prefer bash if available
    if _, err := exec.LookPath("bash"); err == nil {
        return "bash"
    }
    
    return "sh"
}
```

### Execute Method
```go
func (e *executorImpl) Execute(ctx context.Context, dir, command string) (string, string, error) {
    var cmd *exec.Cmd
    
    switch e.shell {
    case "cmd.exe":
        cmd = exec.CommandContext(ctx, "cmd.exe", "/C", command)
    case "bash":
        cmd = exec.CommandContext(ctx, "bash", "-c", command)
    default:
        cmd = exec.CommandContext(ctx, "sh", "-c", command)
    }
    
    cmd.Dir = dir
    
    // Capture stdout and stderr separately
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // Run the command
    err := cmd.Run()
    
    return stdout.String(), stderr.String(), wrapError(err, command, stderr.String())
}
```

### Error Handling
```go
func wrapError(err error, command, stderr string) error {
    if err == nil {
        return nil
    }
    
    // Check if context was cancelled
    if errors.Is(err, context.Canceled) {
        return fmt.Errorf("command cancelled: %s", command)
    }
    
    // Check if context deadline exceeded
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("command timed out: %s", command)
    }
    
    // Include stderr in error message for debugging
    if stderr != "" {
        return fmt.Errorf("command failed: %s\nstderr: %s", command, stderr)
    }
    
    return fmt.Errorf("command failed: %s: %w", command, err)
}
```

## Platform-Specific Implementations

### Windows (executor_windows.go)
```go
//go:build windows

func (e *executorImpl) Execute(ctx context.Context, dir, command string) (string, string, error) {
    // Windows-specific handling
    cmd := exec.CommandContext(ctx, "cmd.exe", "/C", command)
    cmd.Dir = dir
    
    // Set up process group for proper cleanup
    cmd.SysProcAttr = &syscall.SysProcAttr{
        HideWindow:     true,
        CreationFlags:  syscall.CREATE_NEW_PROCESS_GROUP,
    }
    
    return e.runCommand(cmd)
}
```

### Unix (executor_unix.go)
```go
//go:build !windows

func (e *executorImpl) Execute(ctx context.Context, dir, command string) (string, string, error) {
    // Unix-specific handling
    shell := e.shell
    cmd := exec.CommandContext(ctx, shell, "-c", command)
    cmd.Dir = dir
    
    // Set up process group for proper cleanup
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }
    
    return e.runCommand(cmd)
}
```

## Features

### Timeout Support
```go
func (e *executorImpl) ExecuteWithTimeout(
    ctx context.Context,
    dir, command string,
    timeout time.Duration,
) (string, string, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    return e.Execute(ctx, dir, command)
}
```

### Environment Variables
```go
func (e *executorImpl) ExecuteWithEnv(
    ctx context.Context,
    dir, command string,
    env []string,
) (string, string, error) {
    cmd := e.createCommand(ctx, command)
    cmd.Dir = dir
    cmd.Env = append(os.Environ(), env...)
    
    return e.runCommand(cmd)
}
```

### Output Streaming (for large outputs)
```go
type StreamingExecutor interface {
    ExecuteStreaming(
        ctx context.Context,
        dir, command string,
        stdoutWriter, stderrWriter io.Writer,
    ) error
}
```

## Error Types

### Common Errors
1. **Command Not Found**: Executable doesn't exist
2. **Permission Denied**: Insufficient permissions
3. **Timeout**: Command exceeded time limit
4. **Cancelled**: Context was cancelled
5. **Non-Zero Exit**: Command returned error code

### Error Messages
Provide context-rich error messages:
```go
// Bad: "exit status 1"
// Good: "command failed: npm test\nstderr: 5 tests failed"

// Bad: "signal: killed"
// Good: "command timed out after 30s: long-running-test.sh"
```

## Testing Strategy

### Unit Tests
```go
func TestExecutor_Execute(t *testing.T) {
    tests := []struct {
        name    string
        command string
        dir     string
        setup   func() // Set up test environment
        want    struct {
            stdout string
            stderr string
            err    bool
        }
    }{
        {
            name:    "successful echo command",
            command: "echo hello",
            dir:     ".",
            want: struct {
                stdout string
                stderr string
                err    bool
            }{
                stdout: "hello\n",
                stderr: "",
                err:    false,
            },
        },
        {
            name:    "command with error",
            command: "exit 1",
            dir:     ".",
            want: struct {
                stdout string
                stderr string
                err    bool
            }{
                stdout: "",
                stderr: "",
                err:    true,
            },
        },
    }
}
```

### Mock Implementation
```go
type MockExecutor struct {
    responses map[string]struct {
        stdout string
        stderr string
        err    error
    }
}

func (m *MockExecutor) Execute(ctx context.Context, dir, command string) (string, string, error) {
    key := fmt.Sprintf("%s:%s", dir, command)
    if resp, ok := m.responses[key]; ok {
        return resp.stdout, resp.stderr, resp.err
    }
    return "", "", fmt.Errorf("unexpected command: %s", command)
}
```

## Usage Examples

### Basic Usage
```go
executor := executor.New()
stdout, stderr, err := executor.Execute(context.Background(), ".", "ls -la")
if err != nil {
    log.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
}
fmt.Println("Output:", stdout)
```

### With Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stdout, stderr, err := executor.Execute(ctx, "/path/to/project", "npm test")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Fatal("Tests timed out after 30 seconds")
    }
    log.Fatalf("Tests failed: %v", err)
}
```

### With Working Directory
```go
// Execute in specific directory
stdout, _, err := executor.Execute(ctx, "/tmp/build", "make all")
```

## Performance Considerations

1. **Output Buffering**: For large outputs, consider streaming instead of buffering
2. **Process Cleanup**: Ensure child processes are properly terminated
3. **Resource Limits**: Set appropriate limits for memory and CPU
4. **Parallel Execution**: Commands can be executed concurrently with proper context

## Security Considerations

1. **Command Injection**: Never interpolate user input directly into commands
2. **Path Security**: Validate directory paths before execution
3. **Environment Variables**: Be careful with sensitive environment variables
4. **Process Isolation**: Consider using separate process groups

## Platform Compatibility

### Windows
- Handle path separators correctly
- Support both CMD and PowerShell
- Handle different line endings (CRLF vs LF)

### macOS
- Handle zsh as default shell
- Support Homebrew-installed tools
- Handle case-insensitive filesystem

### Linux
- Support different shell variants
- Handle different distributions
- Respect system resource limits

## Future Enhancements

1. **Shell Selection**: Allow users to specify preferred shell
2. **Resource Limits**: Set memory/CPU limits on commands
3. **Progress Reporting**: Report progress for long-running commands
4. **Retry Logic**: Automatic retry for transient failures
5. **Command Chaining**: Support for piped commands