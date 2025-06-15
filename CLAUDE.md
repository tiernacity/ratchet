# Ratchet - Claude Development Guide

## Project Overview
Ratchet is a CLI application written in Go that [describe what ratchet does - you'll need to fill this in based on your specific use case].

## Architecture
- **CLI Application**: Located in `cmd/ratchet/`, built with Go
- **GitHub Action**: Defined in `action.yml`, wraps the CLI binary
- **Configuration**: Supports both command-line flags and YAML configuration files
- **Cross-platform**: Builds for Linux, macOS, and Windows

## Development Guidelines

### Code Structure
- Use standard Go project layout
- CLI framework: Consider using `cobra` for command-line parsing
- Configuration: Use `viper` for handling config files and environment variables
- Follow Go best practices and idiomatic code

### Key Files
- `cmd/ratchet/main.go`: Main CLI entry point
- `action.yml`: GitHub Action definition
- `config/`: Configuration file examples and schemas
- `.github/workflows/`: CI/CD pipelines

### GitHub Workflows
1. **Testing** (`.github/workflows/test.yml`):
   - Run on PR and push to main
   - Test on multiple Go versions
   - Run linting and formatting checks
   - Execute unit and integration tests

2. **Release** (`.github/workflows/release.yml`):
   - Triggered on version tags (v*.*.*)
   - Build cross-platform binaries
   - Create GitHub release with binaries
   - Update GitHub Action marketplace (if applicable)

### Configuration Support
- Command-line flags take precedence over config file
- Support common config formats (YAML preferred)
- Provide clear examples and documentation
- Validate configuration and provide helpful error messages

### GitHub Action Integration
- Action should download appropriate binary for runner OS
- Support input parameters that map to CLI flags
- Provide clear examples in action documentation
- Consider caching binaries for performance

### Release Strategy
- Use semantic versioning (semver)
- Automated releases via GitHub Actions on tag push
- Generate changelog automatically
- Cross-platform binary releases (Linux, macOS, Windows, ARM variants)

### Testing
- Unit tests for core functionality
- Integration tests for CLI commands
- Test GitHub Action functionality
- Consider using testify for assertions

### Documentation
- Clear README with installation and usage instructions
- Command-line help documentation
- GitHub Action usage examples
- Configuration file documentation

## Development Workflow
1. Create feature branches from `main`
2. Write tests for new functionality
3. Ensure all tests pass and code is formatted
4. Create PR with clear description
5. Merge after review and CI passes
6. Tag releases following semver

## Common Commands
```bash
# Run tests
go test ./...

# Build locally
go build -o bin/ratchet cmd/ratchet/main.go

# Install locally
go install ./cmd/ratchet

# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

## Dependencies
- Minimal external dependencies preferred
- Use standard library when possible
- Consider: cobra, viper, testify for testing
- Avoid dependencies with security issues or poor maintenance
