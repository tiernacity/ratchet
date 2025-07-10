# Progress Reporting Feature

## Overview
The Progress Reporting module provides visual feedback during ratchet execution. It shows the current task status with checkboxes that update in real-time, making it clear what's happening and whether tasks succeed or fail.

## Interface

```go
package progress

type Reporter interface {
    // StartTask begins a new task with the given name
    StartTask(name string)
    
    // UpdateTask updates the status of the current task
    UpdateTask(status TaskStatus)
    
    // FinishTask completes the current task with success/failure
    FinishTask(success bool)
    
    // Info prints an informational message
    Info(message string)
    
    // Error prints an error message
    Error(message string)
}

type TaskStatus int

const (
    TaskPending TaskStatus = iota
    TaskRunning
    TaskComplete
    TaskFailed
)
```

## Implementations

### Console Reporter
The main implementation that provides interactive console output with ANSI escape codes.

```go
type consoleReporter struct {
    writer    io.Writer
    verbose   bool
    useColors bool
    mutex     sync.Mutex
    
    currentTask string
    taskCount   int
}

func NewConsole(writer io.Writer, verbose bool) Reporter {
    return &consoleReporter{
        writer:    writer,
        verbose:   verbose,
        useColors: supportsANSI(),
    }
}
```

### No-Op Reporter
Used for testing or when progress reporting is disabled.

```go
type noopReporter struct{}

func NewNoop() Reporter {
    return &noopReporter{}
}

func (n *noopReporter) StartTask(name string)     {}
func (n *noopReporter) UpdateTask(status TaskStatus) {}
func (n *noopReporter) FinishTask(success bool)    {}
func (n *noopReporter) Info(message string)        {}
func (n *noopReporter) Error(message string)       {}
```

## Visual Output Format

### Standard Mode
Shows checkboxes with task names:
```
□ Checking git repository
☑ Checking git repository
□ Creating worktree for main
☑ Creating worktree for main
□ Executing in main: metric command
☑ Executing in main: metric command
□ Executing in HEAD: metric command
☒ Executing in HEAD: metric command
```

### Verbose Mode
Shows detailed progress with sub-tasks:
```
□ Checking git repository
  ↳ Verifying .git directory exists
  ↳ Checking for uncommitted changes
☑ Checking git repository

□ Creating worktree for main
  ↳ Creating temporary directory: /tmp/ratchet-123
  ↳ Running: git worktree add /tmp/ratchet-123/base main
☑ Creating worktree for main

□ Executing in main
  ↳ Running pre-command: npm install
  ↳ Running metric command: npm test -- --coverage
  ↳ Parsing output: 85.5
  ↳ Running post-command: npm run cleanup
☑ Executing in main: metric = 85.5
```

## Implementation Details

### ANSI Support Detection
```go
func supportsANSI() bool {
    // Check if we're in a terminal
    if !isTerminal() {
        return false
    }
    
    // Check for CI environments that support ANSI
    if os.Getenv("CI") != "" {
        // GitHub Actions, GitLab CI, etc. support ANSI
        return true
    }
    
    // Check for explicit color support
    if os.Getenv("CLICOLOR_FORCE") == "1" {
        return true
    }
    
    if os.Getenv("NO_COLOR") != "" {
        return false
    }
    
    // Windows requires special handling
    if runtime.GOOS == "windows" {
        return windowsSupportsANSI()
    }
    
    return true
}
```

### Task Display
```go
func (c *consoleReporter) StartTask(name string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.currentTask = name
    c.taskCount++
    
    if c.useColors {
        // Clear line and print pending checkbox
        fmt.Fprintf(c.writer, "\r\033[K□ %s", name)
    } else {
        fmt.Fprintf(c.writer, "[ ] %s\n", name)
    }
}

func (c *consoleReporter) FinishTask(success bool) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if c.useColors {
        if success {
            // Green checkmark
            fmt.Fprintf(c.writer, "\r\033[K\033[32m☑\033[0m %s\n", c.currentTask)
        } else {
            // Red X
            fmt.Fprintf(c.writer, "\r\033[K\033[31m☒\033[0m %s\n", c.currentTask)
        }
    } else {
        if success {
            fmt.Fprintf(c.writer, "[✓] %s\n", c.currentTask)
        } else {
            fmt.Fprintf(c.writer, "[✗] %s\n", c.currentTask)
        }
    }
}
```

### Verbose Sub-Tasks
```go
func (c *consoleReporter) Info(message string) {
    if !c.verbose {
        return
    }
    
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if c.useColors {
        // Indent sub-tasks
        fmt.Fprintf(c.writer, "\n  \033[90m↳ %s\033[0m", message)
    } else {
        fmt.Fprintf(c.writer, "  - %s\n", message)
    }
}
```

## Platform-Specific Implementations

### Windows (progress_windows.go)
```go
//go:build windows

func windowsSupportsANSI() bool {
    // Windows 10 version 1511+ supports ANSI
    return enableVirtualTerminalProcessing()
}

func enableVirtualTerminalProcessing() bool {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    setConsoleMode := kernel32.NewProc("SetConsoleMode")
    
    var mode uint32
    handle := syscall.Handle(os.Stdout.Fd())
    
    // Get current mode
    err := syscall.GetConsoleMode(handle, &mode)
    if err != nil {
        return false
    }
    
    // Enable virtual terminal processing
    mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
    ret, _, _ := setConsoleMode.Call(uintptr(handle), uintptr(mode))
    
    return ret != 0
}
```

## Testing Strategy

### Unit Tests
```go
func TestConsoleReporter_Output(t *testing.T) {
    tests := []struct {
        name     string
        actions  func(r Reporter)
        expected string
    }{
        {
            name: "successful task",
            actions: func(r Reporter) {
                r.StartTask("Test task")
                r.FinishTask(true)
            },
            expected: "□ Test task\r\033[K\033[32m☑\033[0m Test task\n",
        },
        {
            name: "failed task",
            actions: func(r Reporter) {
                r.StartTask("Test task")
                r.FinishTask(false)
            },
            expected: "□ Test task\r\033[K\033[31m☒\033[0m Test task\n",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer
            reporter := NewConsole(&buf, false)
            
            tt.actions(reporter)
            
            assert.Equal(t, tt.expected, buf.String())
        })
    }
}
```

### Integration Tests
```go
func TestConsoleReporter_RealTerminal(t *testing.T) {
    if !isTerminal() {
        t.Skip("Not running in a terminal")
    }
    
    reporter := NewConsole(os.Stdout, true)
    
    // Simulate a workflow
    reporter.StartTask("Checking git repository")
    time.Sleep(500 * time.Millisecond)
    reporter.FinishTask(true)
    
    reporter.StartTask("Running tests")
    reporter.Info("Executing: go test ./...")
    time.Sleep(1 * time.Second)
    reporter.Info("Tests passed: 42/42")
    reporter.FinishTask(true)
}
```

## Usage Examples

### Basic Usage
```go
reporter := progress.NewConsole(os.Stdout, config.Verbose)

reporter.StartTask("Checking git repository")
if err := git.IsGitRepository(); err != nil {
    reporter.FinishTask(false)
    return err
}
reporter.FinishTask(true)
```

### With Sub-Tasks
```go
reporter.StartTask("Executing metric command")
reporter.Info(fmt.Sprintf("Running: %s", config.MetricCmd))

output, _, err := executor.Execute(ctx, ".", config.MetricCmd)
if err != nil {
    reporter.Error(fmt.Sprintf("Command failed: %v", err))
    reporter.FinishTask(false)
    return err
}

reporter.Info(fmt.Sprintf("Output: %s", strings.TrimSpace(output)))
reporter.FinishTask(true)
```

## Configuration Options

### Environment Variables
- `NO_COLOR`: Disable colored output
- `CLICOLOR_FORCE`: Force colored output even in non-terminal
- `RATCHET_PROGRESS`: Set to "none" to disable progress

### Reporter Selection
```go
func createReporter(config *Config) Reporter {
    if os.Getenv("RATCHET_PROGRESS") == "none" {
        return progress.NewNoop()
    }
    
    if config.JSON {
        return progress.NewJSON(os.Stdout)
    }
    
    return progress.NewConsole(os.Stdout, config.Verbose)
}
```

## Future Enhancements

### JSON Reporter
For CI/CD integration:
```go
type jsonReporter struct {
    writer io.Writer
    events []Event
}

type Event struct {
    Type      string    `json:"type"`
    Task      string    `json:"task,omitempty"`
    Status    string    `json:"status,omitempty"`
    Message   string    `json:"message,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Progress Bar
For long-running tasks:
```go
type ProgressBarReporter interface {
    Reporter
    
    // SetProgress updates a progress bar (0-100)
    SetProgress(task string, percent int)
}
```

### Spinner Animation
For indeterminate progress:
```go
func (c *consoleReporter) ShowSpinner(task string) {
    frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    // Animate spinner while task runs
}
```

## Performance Considerations

1. **Thread Safety**: Use mutex for concurrent access
2. **Buffer Flushing**: Ensure output is flushed immediately
3. **ANSI Overhead**: Minimal performance impact
4. **Terminal Detection**: Cache the result to avoid repeated checks

## Accessibility

1. **Screen Readers**: Provide text-only mode
2. **Color Blindness**: Don't rely solely on color
3. **Motion Sensitivity**: Option to disable animations
4. **High Contrast**: Support system theme preferences