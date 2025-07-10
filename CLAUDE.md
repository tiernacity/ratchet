# Ratchet - Claude Development Guide

## Project Overview
Ratchet is a CLI application written in Go that enforces continuous improvement by comparing command output metrics between Git branches. It ensures metrics move in the desired direction (e.g., reducing TODO comments, increasing test coverage). It is also available as a GitHub Action for use in CI/CD workflows.

## High-Level Architecture

### Design Principles
1. **Testability First**: Thin CLI layer with all business logic in testable modules
2. **Interface-Driven**: Use interfaces to enable mocking and testing
3. **Separation of Concerns**: Each module has a single, well-defined responsibility
4. **Dependency Injection**: Components receive dependencies rather than creating them
5. **Error Bubbling**: Errors flow up to main for consistent exit code handling

### Module Structure
```
/
├── cmd/ratchet/         # Thin CLI layer (argument parsing only)
├── internal/
│   ├── ratchet/         # Core orchestrator module
│   ├── git/             # Git operations interface and implementation
│   ├── executor/        # Command execution interface and implementation
│   ├── parser/          # Metric parsing and validation
│   ├── config/          # Configuration types and validation
│   ├── progress/        # Progress reporting interface and implementation
│   └── errors/          # Custom error types for proper exit codes
└── docs/
    └── features/        # Detailed feature documentation
```

## Development Guidelines

### Code Structure
- **CLI Layer**: Must be extremely thin - only parse arguments and pass to orchestrator
- **Interfaces**: Define interfaces for all external dependencies (Git, command execution, output)
- **Testing**: Every module must have comprehensive unit tests
- **No Direct I/O**: All printing/file operations must be injected as dependencies
- **Platform Independence**: Use build tags for platform-specific code

### Testing Requirements
1. **Unit Tests Required For**:
   - Core orchestrator logic (`internal/ratchet`)
   - Configuration validation (`internal/config`)
   - Metric parsing (`internal/parser`)
   - Error handling (`internal/errors`)
   - Progress reporting (`internal/progress`)

2. **Integration Tests For**:
   - Full CLI flow with mocked Git and executor
   - Git worktree operations (using test repositories)
   - Command execution with various scenarios

3. **Test Patterns**:
   - Use table-driven tests for validation logic
   - Mock interfaces for external dependencies
   - Test error conditions as thoroughly as success paths
   - Use testify for assertions and mocking

### Error Handling
- Define custom error types in `internal/errors`
- Use error wrapping with `fmt.Errorf("%w", err)` for context
- Exit codes determined by error type in main():
  - 0: Success
  - 1: Metric test failed (expected failure)
  - 2: Unexpected error or execution failure

### Git Worktree Management
- **Critical**: Must reliably clean up worktrees in ALL error conditions
- Use defer for cleanup immediately after creation
- Handle interrupt signals (SIGINT, SIGTERM) for cleanup
- Force worktree creation to handle existing worktrees
- Use unique names to avoid conflicts

### Configuration
- Command-line flags take precedence over config file
- Support YAML and JSON config formats via viper
- Validate all configuration at the boundary
- Use strong types (not strings) for operators

### Key Implementation Notes
1. **Temporary Directory**: Use platform-appropriate temp dirs (respect RUNNER_TEMP in GitHub Actions)
2. **Progress Display**: Inject output writer, update lines in-place with ANSI codes
3. **Command Execution**: Support timeouts, capture both stdout and stderr
4. **Metric Comparison**: Support both integer and floating-point comparisons

### Documentation Maintenance
**CRITICAL**: As the implementation evolves, the following documentation must be kept synchronized:
- `DESIGN.md` - Update module interfaces, architecture decisions, and execution flow
- `docs/features/*.md` - Update feature specifications when implementation changes
- If interfaces change, update the corresponding feature documentation immediately
- If new modules are added, create new feature documentation files
- If error handling changes, update both DESIGN.md and docs/features/error-handling.md
- This documentation is not just reference material - it's the authoritative specification

## Common Commands
```bash
# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Build locally
go build -o bin/ratchet cmd/ratchet/main.go

# Install locally
go install ./cmd/ratchet

# Format code
go fmt ./...

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Run tests with race detector
go test -race ./...
```

## Dependencies
- **CLI**: `github.com/spf13/cobra` - command-line parsing
- **Config**: `github.com/spf13/viper` - configuration management
- **Testing**: `github.com/stretchr/testify` - assertions and mocking
- **No other external dependencies** - use standard library for everything else

## Code Style Guidelines
1. **Naming**:
   - Interfaces end with `-er` suffix (e.g., `Executor`, `Parser`)
   - Exported functions/types have descriptive names
   - Unexported functions use camelCase

2. **Error Messages**:
   - Start with lowercase
   - Include context about what failed
   - Don't include punctuation at the end

3. **Comments**:
   - Package comments describe the package purpose
   - Exported types/functions have godoc comments
   - Implementation comments explain "why", not "what"

4. **File Organization**:
   - One primary type per file
   - Tests in `*_test.go` files
   - Platform-specific code uses build tags

## Development Workflow
1. Create feature branches from `main`
2. Write tests FIRST (TDD approach)
3. Implement until tests pass
4. Ensure all tests pass and coverage is maintained
5. Run linter and fix any issues
6. Create PR with clear description
7. Merge after review and CI passes

## Release Process
1. Ensure all tests pass on main branch
2. Update version in relevant files
3. Create and push semantic version tag (e.g., v1.0.0)
4. GitHub Actions will automatically:
   - Build cross-platform binaries
   - Create GitHub release
   - Attach binaries to release
   - Update GitHub Action marketplace

## Debugging Tips
1. Use `-v/--verbose` flag to see detailed execution steps
2. Check Git worktree list if cleanup fails: `git worktree list`
3. Set `RATCHET_DEBUG=1` environment variable for debug logs
4. Use `go test -v` to see test output
5. Use debugger with VS Code or Delve for complex issues