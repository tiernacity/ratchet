# Ratchet Design Document

## Overview

Ratchet is designed as a modular, testable CLI application that enforces continuous improvement by comparing metrics between Git branches. This document describes the detailed design, module responsibilities, and interactions.

## Core Design Principles

1. **Thin CLI Layer**: The CLI is only responsible for parsing arguments and delegating to the orchestrator
2. **Interface-Based Architecture**: All external dependencies are abstracted behind interfaces
3. **Dependency Injection**: Components receive their dependencies, enabling easy testing
4. **Single Responsibility**: Each module has one clear purpose
5. **Error Propagation**: Errors bubble up to the CLI layer for consistent exit code handling

## Module Architecture

### Directory Structure
```
ratchet/
├── cmd/
│   └── ratchet/
│       ├── main.go          # Entry point, exit code handling
│       └── root.go          # Cobra command setup, flag parsing
├── internal/
│   ├── ratchet/
│   │   ├── orchestrator.go  # Core orchestration logic
│   │   ├── orchestrator_test.go
│   │   └── interfaces.go    # Interface definitions
│   ├── config/
│   │   ├── config.go        # Configuration types
│   │   ├── validation.go    # Configuration validation
│   │   └── validation_test.go
│   ├── git/
│   │   ├── interface.go     # Git operations interface
│   │   ├── git.go           # Git implementation
│   │   ├── git_test.go
│   │   └── worktree.go      # Worktree management
│   ├── executor/
│   │   ├── interface.go     # Command execution interface
│   │   ├── executor.go      # Command executor implementation
│   │   └── executor_test.go
│   ├── parser/
│   │   ├── parser.go        # Metric parsing logic
│   │   └── parser_test.go
│   ├── progress/
│   │   ├── interface.go     # Progress reporter interface
│   │   ├── console.go       # Console progress implementation
│   │   ├── noop.go          # No-op implementation (for testing)
│   │   └── console_test.go
│   └── errors/
│       ├── errors.go        # Custom error types
│       └── errors_test.go
└── docs/
    └── features/            # Detailed feature documentation
```

## Module Responsibilities

### cmd/ratchet
**Purpose**: CLI entry point and argument parsing

**Responsibilities**:
- Parse command-line arguments using Cobra
- Load configuration file if specified
- Merge CLI flags with config file (flags take precedence)
- Create and configure the orchestrator
- Handle exit codes based on returned errors
- Set up signal handlers for cleanup

**Key Interfaces**: None (uses concrete types from internal packages)

### internal/ratchet
**Purpose**: Core orchestration of the ratchet process

**Responsibilities**:
- Coordinate all components to execute the ratchet workflow
- Manage the lifecycle of Git worktrees
- Execute commands in appropriate directories
- Compare metrics and determine pass/fail
- Report progress through injected reporter

**Key Interfaces**:
```go
type Orchestrator interface {
    Run(ctx context.Context, config *config.Config) error
}

type GitOperations interface {
    IsGitRepository() error
    CreateWorktree(path, branch string) error
    RemoveWorktree(path string) error
    ResolveBranch(branch string) (string, error)
}

type CommandExecutor interface {
    Execute(ctx context.Context, dir, command string) (string, error)
}

type MetricParser interface {
    Parse(output string) (float64, error)
}

type ProgressReporter interface {
    Start(baseRef, headRef string, verbose bool)
    UpdateBranch(branch string, phase string, completed bool)
    Success(current, base float64, operator, branch string)
    Error(message string)
    NoComparison(value float64)
    Info(message string)
}
```

### internal/config
**Purpose**: Configuration types and validation

**Responsibilities**:
- Define configuration structure
- Validate configuration values
- Ensure exactly one comparison operator is specified
- Validate that base branch is provided
- Normalize configuration (e.g., resolve operator aliases)

**Types**:
```go
type Config struct {
    BaseBranch   string
    MetricCmd    string
    PreCmd       string
    PostCmd      string
    Operator     ComparisonOperator
    ConfigFile   string
    Verbose      bool
}

type ComparisonOperator int
const (
    GreaterThan ComparisonOperator = iota
    GreaterThanOrEqual
    Equal
    LessThanOrEqual
    LessThan
)
```

### internal/git
**Purpose**: Git repository operations

**Responsibilities**:
- Check if current directory is a Git repository
- Create and manage Git worktrees
- Clean up worktrees reliably
- Resolve branch references to commits

**Implementation Notes**:
- Uses `git` command-line tool via exec
- Handles platform-specific path separators
- Forces worktree creation to handle existing worktrees
- Generates unique worktree names to avoid conflicts

### internal/executor
**Purpose**: Execute shell commands

**Responsibilities**:
- Execute commands with proper context handling
- Capture stdout and stderr separately
- Support command timeouts
- Handle platform-specific shell differences
- Provide clear error messages on failure

**Implementation Notes**:
- Uses appropriate shell based on platform (bash/sh on Unix, cmd on Windows)
- Respects context cancellation
- Returns both stdout and stderr for debugging

### internal/parser
**Purpose**: Parse and validate metric output

**Responsibilities**:
- Parse command output as numeric value
- Support both integer and floating-point numbers
- Validate that output contains exactly one number
- Provide clear error messages for invalid output

**Implementation Notes**:
- Trims whitespace from output
- Uses strconv for parsing
- Returns float64 to support both int and float comparisons

### internal/progress
**Purpose**: Report execution progress

**Responsibilities**:
- Display progress of each task
- Update progress in-place (when supported)
- Handle both verbose and quiet modes
- Provide different implementations (console, no-op for testing)

**Implementation Notes**:
- Console implementation uses ANSI escape codes
- Falls back to simple printing if terminal doesn't support ANSI
- No-op implementation for unit tests

### internal/errors
**Purpose**: Custom error types for proper exit codes and help display control

**Responsibilities**:
- Define error types that map to exit codes
- Distinguish between test failures and execution errors
- Provide error wrapping utilities
- Control when help text is displayed to users

**Types**:
```go
type MetricTestError struct {
    Current  float64
    Base     float64
    Operator string
    Branch   string
}

type ValidationError struct {
    Field   string
    Message string
}

type GitError struct {
    Operation string
    Wrapped   error
}

type CommandError struct {
    Command string
    Wrapped error
}

type ParseError struct {
    Output  string
    Wrapped error
}
```

**Key Functions**:
- `GetExitCode(err error) int` - Returns appropriate exit code for error type
- `ShouldSuppressHelp(err error) bool` - Determines if help should be shown

## Execution Flow

1. **CLI Parsing** (cmd/ratchet)
   - Parse command-line arguments
   - Load and merge config file if specified
   - Validate configuration
   - Create orchestrator with dependencies

2. **Orchestration** (internal/ratchet)
   - Validate Git repository
   - Set up signal handlers for cleanup
   - Create temporary directory and worktree
   - Execute workflow with progress reporting
   - Clean up resources
   - Return appropriate error

3. **Workflow Steps**:
   ```
   ┌─────────────────────┐
   │ Check Git Repo      │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │ Create Worktree     │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │ Run Base Branch     │
   │ - Pre-command       │
   │ - Metric command    │
   │ - Post-command      │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │ Run Current Branch  │
   │ - Pre-command       │
   │ - Metric command    │
   │ - Post-command      │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │ Compare Metrics     │
   └──────────┬──────────┘
              │
   ┌──────────▼──────────┐
   │ Cleanup & Return    │
   └─────────────────────┘
   ```

## Error Handling Strategy

1. **Validation Errors**: Return ValidationError, show help text (exit code 2)
2. **Git Errors**: Wrap with GitError, suppress help (exit code 2)
3. **Command Failures**: Wrap with CommandError, suppress help (exit code 2)
4. **Parse Errors**: Wrap with ParseError, suppress help (exit code 2)
5. **Metric Test Failures**: Return MetricTestError, suppress help (exit code 1)

### Help Display Logic
- **Show help**: ValidationError, CLI parsing errors (missing args, unknown flags)
- **Suppress help**: All runtime errors (GitError, CommandError, ParseError, MetricTestError)
- Implemented via `cmd.SilenceUsage = true` when `errors.ShouldSuppressHelp()` returns true

## Testing Strategy

### Unit Tests
- Mock all interfaces for isolated testing
- Table-driven tests for validation logic
- Test error conditions thoroughly
- Achieve >80% code coverage

### Integration Tests
- Test full workflow with real Git operations
- Test various command scenarios
- Verify cleanup in error conditions
- Test signal handling

### Test Utilities
```go
// Mock implementations in test files
type mockGit struct {
    mock.Mock
}

type mockExecutor struct {
    outputs map[string]string
    errors  map[string]error
}

type mockProgress struct {
    tasks []string
}
```

## Security Considerations

1. **Command Injection**: Commands are executed as-is, users must trust their metric commands
2. **File System**: Temporary directories are created with secure permissions
3. **Git Operations**: Only local operations are performed, no network access
4. **Resource Cleanup**: Worktrees are always cleaned up to prevent accumulation

## Performance Considerations

1. **Git Operations**: Worktree creation is fast for local branches
2. **Command Execution**: Commands run with timeouts to prevent hanging
3. **Memory Usage**: Minimal memory footprint, output is streamed
4. **Parallel Execution**: Commands run sequentially to ensure consistent results

## Future Extensibility

1. **Multiple Metrics**: Architecture supports comparing multiple metrics
2. **Remote Execution**: Executor interface could support remote execution
3. **Different VCS**: Git interface could be generalized to support other VCS
4. **Reporting**: Progress interface could support different output formats