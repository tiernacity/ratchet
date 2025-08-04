# Implementation Notes

This document captures key implementation decisions and technical constraints for ongoing development.

## CLI Interface Design

### Flag Structure

- Comparison operators: `--lt`, `--le`, `--eq`, `--ge`, `--gt` with required base branch argument
- Commands: `--metric` (alternative to positional), `--pre`, `--post`
- Configuration: `--config-file`, `--config` (inline YAML/JSON)
- Output: `--verbose`, `--version`, `--help`

### Operator Precedence

1. Command-line flags override config file values
2. Config file values override defaults
3. Exactly one comparison operator must be specified

### Configuration File Support

- YAML and JSON formats supported
- Field mapping: CLI kebab-case → config snake_case → Go PascalCase
- Config file discovery: `.ratchet` in current directory by default

## Git Operations

### Worktree Management Strategy

- Create temporary worktrees for base branch comparison
- Use unique temporary directory names: `ratchet-*`
- Always use `--force` flag for robust operations
- Clean up worktrees in ALL error conditions, or when interrupted by a signal, using defer

### Branch Resolution

- Support any git-ish reference: branches, tags, commits, remote refs
- Handle both local and remote branch references
- Fail early if base branch cannot be resolved

### Repository Validation

- Verify current directory is in a git repository
- Warn about uncommitted changes (but don't fail)
- Support repositories with and without remotes

## Command Execution

### Executor Interface

```go
type CommandExecutor interface {
    Execute(ctx context.Context, dir, command string) (stdout, stderr string, err error)
}
```

### Execution Context

- Commands run in specific directories (worktree for base, current dir for HEAD)
- Support context cancellation for timeouts
- Capture both stdout and stderr separately
- Pass through command exit codes as part of wrapped errors

### Platform Considerations

- Use appropriate shell detection (cmd.exe on Windows, /bin/sh elsewhere)
- Handle environment variable inheritance
- Support commands with arguments and shell operators

## Configuration Structure

### Current Implementation

```go
type Config struct {
    Metric       string  // The metric command
    Pre          string  // Pre-command (optional)
    Post         string  // Post-command (optional)
    BaseBranch   string  // Base branch for comparison
    Operator     ComparisonOperator
    Verbose      bool
    ConfigFile   string  // Path to config file
}
```

### Validation Rules

1. Exactly one comparison operator required
2. Metric command cannot be empty
3. Base branch defaults to "main" if not specified
4. Config file must be valid YAML/JSON if specified

## Metric Parsing

### Parser Interface

```go
type MetricParser interface {
    Parse(output string) (float64, error)
}
```

### Parsing Rules

1. Extract the first numeric value from command output
2. Support both integer and floating-point numbers
3. Ignore leading/trailing whitespace and text
4. Fail if no numeric value is found
5. Fail if output is empty

### Example Valid Outputs

```
42
42.5
Coverage: 85.2%
Found 123 issues
  42
```

### Invalid Outputs

```
hello world
(empty output)
no numbers here
Found 123 issues in 3 files
```

## Progress Reporting

### Interface Design

```go
type ProgressReporter interface {
    StartBranch(branch string, commands []string)
    UpdateCommand(branch, command string, completed bool)
    FinishBranch(branch string, success bool)
    Success(current, base float64, operator, branch string)
    Failure(current, base float64, operator, branch string)
}
```

### Implementation Strategy

- Console implementation for interactive terminals
- No-op implementation for non-interactive output
- ANSI escape sequences for in-place updates
- Graceful fallback when terminal doesn't support ANSI

## Dependency Management

### External Dependencies

- `github.com/spf13/cobra` - CLI parsing
- `github.com/spf13/viper` - Configuration management
- `github.com/stretchr/testify` - Testing framework
- Standard library only for everything else

### Dependency Injection

All major components use interfaces:

- GitOperations for Git commands
- CommandExecutor for shell commands
- MetricParser for output parsing
- ProgressReporter for user feedback

## Testing Strategy

### Unit Test Coverage

- Configuration validation (internal/config)
- Error type creation and detection (internal/errors)
- Metric parsing edge cases (internal/parser)
- Progress reporting output (internal/progress)

### Integration Test Coverage

- Full CLI workflow with mocked dependencies
- Git operations with test repositories
- Command execution with various exit codes
- Configuration file loading and merging

### Test Data Management

- Use testdata directories for config file examples
- Create temporary Git repositories for integration tests
- Mock external dependencies consistently

## Platform-Specific Considerations

### Temporary Directory Management

```go
func createTempDir() (string, error) {
    // Respect GitHub Actions RUNNER_TEMP
    tempBase := os.TempDir()
    if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
        tempBase = runnerTemp
    }
    return os.MkdirTemp(tempBase, "ratchet-*")
}
```

### Signal Handling

- Handle SIGINT and SIGTERM for cleanup
- Ensure worktrees are removed on interruption
- Exit with appropriate code (2) on signal

### Cross-Platform Compatibility

- Use filepath.Join for path construction
- Handle different line endings appropriately
- Account for different shell behaviors

## Performance Considerations

### Git Operations

- Worktree creation is fast for local operations
- Remote branch resolution may require network
- Use shallow clones where possible in CI

### Command Execution

- Set reasonable default timeouts
- Stream output rather than buffering large results
- Cancel operations on context timeout

### Memory Usage

- Parse metric output immediately, don't store large strings
- Clean up temporary files promptly
- Use defer for resource cleanup

## Security Considerations

### Command Execution

- Execute commands through system shell (no injection protection needed as user controls input)
- Inherit environment variables from parent process
- Run commands in appropriate working directories

### Temporary Files

- Create temporary directories with appropriate permissions
- Clean up all temporary files on exit
- Use cryptographically secure random names

## Future Extensibility

### Potential Enhancements

1. Parallel execution (run base and current branch simultaneously)
2. Custom output formats (JSON, XML for CI integration)
3. Plugin system (custom metric parsers or reporters)

### Backward Compatibility

- Maintain CLI flag compatibility
- Support old config file formats
- Preserve exit code behavior
- Keep error message formats stable

## Development Workflow

### Code Organization

- Thin CLI layer in cmd/ratchet
- Business logic in internal/ packages
- Interfaces defined in separate files from implementations
- Tests alongside implementation files

### Documentation Requirements

- Update DESIGN_PATTERNS.md when error handling changes
- Update OUTPUT_SPECIFICATION.md when message formats change
- Update this file when interfaces or major patterns change
- Keep README.md examples current with actual behavior

### Release Process

- All tests must pass before merging
- Update version in internal/version
- Tag releases with semantic versioning
- GitHub Actions handles cross-platform builds
