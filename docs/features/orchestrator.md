# Orchestrator Feature

## Overview
The Orchestrator is the core component that coordinates all aspects of the ratchet workflow. It manages the lifecycle of Git worktrees, executes commands, compares metrics, and reports progress.

## Responsibilities

1. **Workflow Coordination**: Manage the complete ratchet workflow from start to finish
2. **Resource Management**: Create and clean up temporary directories and Git worktrees
3. **Signal Handling**: Ensure cleanup on interruption (SIGINT, SIGTERM)
4. **Error Propagation**: Convert internal errors to appropriate exit codes
5. **Progress Reporting**: Coordinate progress updates through the workflow

## Interface Design

```go
package ratchet

type Orchestrator interface {
    Run(ctx context.Context, config *config.Config) error
}

type orchestrator struct {
    git      GitOperations
    executor CommandExecutor
    parser   MetricParser
    progress ProgressReporter
}
```

## Workflow Steps

```
1. Validate Prerequisites
   ├── Check Git repository
   └── Check for uncommitted changes (warning only)

2. Setup Environment
   ├── Create temporary directory
   ├── Set up signal handlers
   └── Create Git worktree for base branch

3. Execute Base Branch
   ├── Run pre-command (if specified)
   ├── Run metric command
   ├── Parse metric output
   └── Run post-command (if specified)

4. Execute Current Branch
   ├── Run pre-command (if specified)
   ├── Run metric command
   ├── Parse metric output
   └── Run post-command (if specified)

5. Compare and Report
   ├── Compare metrics based on operator
   ├── Report success or failure
   └── Return appropriate error

6. Cleanup
   ├── Remove Git worktree
   └── Remove temporary directory
```

## Implementation Details

### Constructor
```go
func NewOrchestrator(
    git GitOperations,
    executor CommandExecutor,
    parser MetricParser,
    progress ProgressReporter,
) Orchestrator {
    return &orchestrator{
        git:      git,
        executor: executor,
        parser:   parser,
        progress: progress,
    }
}
```

### Run Method Structure
```go
func (o *orchestrator) Run(ctx context.Context, cfg *config.Config) error {
    // 1. Validate prerequisites
    if err := o.validatePrerequisites(); err != nil {
        return err
    }
    
    // 2. Setup environment with cleanup
    cleanup, err := o.setupEnvironment(ctx, cfg)
    if err != nil {
        return err
    }
    defer cleanup()
    
    // 3. Execute workflow
    baseMetric, err := o.executeInBranch(ctx, cfg, cfg.BaseBranch, tempDir)
    if err != nil {
        return err
    }
    
    currentMetric, err := o.executeInBranch(ctx, cfg, "HEAD", ".")
    if err != nil {
        return err
    }
    
    // 4. Compare metrics
    if !o.compareMetrics(baseMetric, currentMetric, cfg.Operator) {
        return &errors.MetricTestError{
            Message: fmt.Sprintf("metric test failed: %.2f %s %.2f",
                currentMetric, cfg.Operator, baseMetric),
        }
    }
    
    return nil
}
```

### Signal Handling
```go
func (o *orchestrator) setupSignalHandler(cleanup func()) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        o.progress.Error("Received interrupt signal, cleaning up...")
        cleanup()
        os.Exit(2)
    }()
}
```

### Worktree Management
```go
func (o *orchestrator) setupEnvironment(ctx context.Context, cfg *config.Config) (func(), error) {
    // Create temp directory
    tempDir, err := o.createTempDir()
    if err != nil {
        return nil, err
    }
    
    // Create worktree
    worktreePath := filepath.Join(tempDir, "ratchet-base")
    if err := o.git.CreateWorktree(worktreePath, cfg.BaseBranch); err != nil {
        os.RemoveAll(tempDir)
        return nil, err
    }
    
    // Return cleanup function
    cleanup := func() {
        o.git.RemoveWorktree(worktreePath)
        os.RemoveAll(tempDir)
    }
    
    return cleanup, nil
}
```

### Command Execution
```go
func (o *orchestrator) executeInBranch(
    ctx context.Context,
    cfg *config.Config,
    branch string,
    dir string,
) (float64, error) {
    o.progress.StartTask(fmt.Sprintf("Executing in %s", branch))
    
    // Pre-command
    if cfg.PreCmd != "" {
        o.progress.UpdateTask("pre-command", TaskRunning)
        if _, _, err := o.executor.Execute(ctx, dir, cfg.PreCmd); err != nil {
            o.progress.FinishTask("pre-command", false)
            return 0, fmt.Errorf("pre-command failed: %w", err)
        }
    }
    
    // Metric command
    o.progress.UpdateTask("metric-command", TaskRunning)
    output, _, err := o.executor.Execute(ctx, dir, cfg.MetricCmd)
    if err != nil {
        o.progress.FinishTask("metric-command", false)
        return 0, fmt.Errorf("metric command failed: %w", err)
    }
    
    // Parse metric
    metric, err := o.parser.Parse(output)
    if err != nil {
        return 0, fmt.Errorf("failed to parse metric: %w", err)
    }
    
    // Post-command
    if cfg.PostCmd != "" {
        o.progress.UpdateTask("post-command", TaskRunning)
        if _, _, err := o.executor.Execute(ctx, dir, cfg.PostCmd); err != nil {
            o.progress.FinishTask("post-command", false)
            return 0, fmt.Errorf("post-command failed: %w", err)
        }
    }
    
    o.progress.FinishTask(fmt.Sprintf("Executing in %s", branch), true)
    return metric, nil
}
```

### Metric Comparison
```go
func (o *orchestrator) compareMetrics(
    current, base float64,
    operator config.ComparisonOperator,
) bool {
    switch operator {
    case config.GreaterThan:
        return current > base
    case config.GreaterThanOrEqual:
        return current >= base
    case config.Equal:
        return current == base
    case config.LessThanOrEqual:
        return current <= base
    case config.LessThan:
        return current < base
    default:
        return false
    }
}
```

## Error Handling

### Error Types
1. **Validation Errors**: Git repository not found, invalid configuration
2. **Setup Errors**: Failed to create temp directory or worktree
3. **Execution Errors**: Command execution failures
4. **Parse Errors**: Invalid metric output
5. **Test Failures**: Metric comparison failed (expected)

### Error Propagation
- All errors except metric test failures are wrapped as `ExecutionError`
- Metric test failures return `MetricTestError`
- Errors include context about what failed and why

## Testing Strategy

### Unit Tests
```go
func TestOrchestrator_Run(t *testing.T) {
    tests := []struct {
        name        string
        config      *config.Config
        gitMock     func(*mockGit)
        execMock    func(*mockExecutor)
        wantErr     error
    }{
        {
            name: "successful run with metric improvement",
            config: &config.Config{
                BaseBranch: "main",
                MetricCmd:  "count-todos",
                Operator:   config.LessThan,
            },
            gitMock: func(m *mockGit) {
                m.On("IsGitRepository").Return(nil)
                m.On("CreateWorktree", mock.Anything, "main").Return(nil)
                m.On("RemoveWorktree", mock.Anything).Return(nil)
            },
            execMock: func(m *mockExecutor) {
                m.On("Execute", mock.Anything, mock.Anything, "count-todos").
                    Return("10", "", nil).Once() // base branch
                m.On("Execute", mock.Anything, ".", "count-todos").
                    Return("5", "", nil).Once() // current branch
            },
            wantErr: nil,
        },
    }
}
```

### Integration Tests
- Test full workflow with real Git operations
- Test cleanup in various error scenarios
- Test signal handling and cleanup
- Test with different command combinations

## Platform Considerations

### Temporary Directory
```go
func (o *orchestrator) createTempDir() (string, error) {
    // Respect RUNNER_TEMP in GitHub Actions
    tempBase := os.TempDir()
    if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
        tempBase = runnerTemp
    }
    
    return os.MkdirTemp(tempBase, "ratchet-*")
}
```

### Signal Handling
- Handle SIGINT and SIGTERM on Unix systems
- Ensure cleanup runs even on Windows (using appropriate signals)

## Performance Considerations

1. **Worktree Creation**: Fast for local operations
2. **Command Execution**: Run with reasonable timeouts
3. **Memory Usage**: Stream command output rather than buffering
4. **Cleanup**: Ensure resources are freed promptly

## Future Enhancements

1. **Parallel Execution**: Could run base and current branch in parallel
2. **Caching**: Cache base branch metrics for repeated runs
3. **Multiple Metrics**: Support comparing multiple metrics in one run
4. **Custom Reporters**: Support different output formats (JSON, XML)