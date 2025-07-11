# CLI Interface Feature

## Overview
The CLI interface provides the command-line argument parsing and configuration management for Ratchet. It uses Cobra for command parsing and Viper for configuration file support.

## Command Structure

### Basic Usage
```bash
ratchet --base-branch main --metric-cmd "grep -c TODO *.go"
```

### Full Command Syntax
```bash
ratchet [flags]
```

## Flags

### Required Flags
One of the following comparison operators must be specified:
- `--gt, --greater-than`: Current metric must be greater than base
- `--ge, --greater-than-or-equal`: Current metric must be greater than or equal to base
- `--eq, --equal`: Current metric must be equal to base
- `--le, --less-than-or-equal`: Current metric must be less than or equal to base
- `--lt, --less-than`: Current metric must be less than base

### Optional Flags
- `--base-branch string`: Base branch to compare against (default: "main")
- `--metric-cmd string`: Command that outputs a single metric (required unless in config)
- `--pre-cmd string`: Command to run before metric command
- `--post-cmd string`: Command to run after metric command
- `--config string`: Path to config file (YAML or JSON)
- `-v, --verbose`: Enable verbose output
- `-h, --help`: Show help message
- `--version`: Show version information

## Configuration File Support

### File Formats
Supports both YAML and JSON configuration files.

### YAML Example
```yaml
base-branch: main
metric-cmd: "go test -cover | grep coverage | awk '{print $2}' | sed 's/%//'"
pre-cmd: "go mod download"
post-cmd: "go clean -testcache"
greater-than: true
verbose: true
```

### JSON Example
```json
{
  "base-branch": "main",
  "metric-cmd": "eslint . --format json | jq length",
  "less-than": true,
  "verbose": false
}
```

### Configuration Precedence
1. Command-line flags (highest priority)
2. Configuration file
3. Default values (lowest priority)

## Implementation Details

### File Structure
```
cmd/ratchet/
├── main.go    # Entry point, exit code handling
└── root.go    # Cobra command setup, flag definitions
```

### main.go Responsibilities
- Initialize Cobra command
- Execute command
- Handle exit codes based on error types:
  - 0: Success
  - 1: Metric test failed (expected)
  - 2: Execution error

### root.go Responsibilities
- Define root command
- Register all flags with proper types
- Set up Viper for config file support
- Validate configuration
- Create and run orchestrator

### Key Implementation Points

1. **Thin Implementation**: The CLI layer only parses arguments and delegates to the orchestrator
2. **Flag Aliases**: Support both long and short forms (e.g., --gt and --greater-than)
3. **Validation**: Ensure exactly one comparison operator is specified
4. **Config Loading**: Use Viper's automatic config file detection
5. **Error Messages**: Provide clear, actionable error messages

### Error Handling
```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        switch err.(type) {
        case *errors.MetricTestError:
            os.Exit(1) // Test failed (expected)
        default:
            os.Exit(2) // Execution error
        }
    }
}
```

### Configuration Validation
```go
import (
    "github.com/tiernacity/ratchet/internal/errors"
)

func validateConfig(cfg *config.Config) error {
    // Ensure metric command is provided
    if cfg.MetricCmd == "" {
        return errors.NewValidationError("metric-cmd", "metric command is required")
    }
    
    // Ensure exactly one operator is specified
    operatorCount := 0
    if cfg.GreaterThan { operatorCount++ }
    if cfg.GreaterThanOrEqual { operatorCount++ }
    if cfg.Equal { operatorCount++ }
    if cfg.LessThanOrEqual { operatorCount++ }
    if cfg.LessThan { operatorCount++ }
    
    if operatorCount != 1 {
        return errors.NewValidationError("operator", "exactly one comparison operator must be specified")
    }
    
    return nil
}
```

## Testing Strategy

### Unit Tests
- Test flag parsing with various combinations
- Test config file loading
- Test configuration validation
- Test precedence rules

### Integration Tests
- Test full command execution with different flags
- Test config file with CLI override
- Test error conditions and exit codes

## User Experience

### Help Output
```
Ratchet ensures your metrics always improve

Usage:
  ratchet [flags]

Comparison Operators (exactly one required):
  --gt, --greater-than              Current branch metric > base branch metric
  --ge, --greater-than-or-equal     Current branch metric >= base branch metric
  --eq, --equal                     Current branch metric == base branch metric
  --le, --less-than-or-equal        Current branch metric <= base branch metric
  --lt, --less-than                 Current branch metric < base branch metric

Flags:
      --base-branch string   Base branch to compare against (default "main")
      --metric-cmd string    Command that outputs a single metric
      --pre-cmd string       Command to run before metric command
      --post-cmd string      Command to run after metric command
      --config string        Config file path (YAML or JSON)
  -v, --verbose             Enable verbose output
  -h, --help               Help for ratchet
      --version            Version for ratchet
```

### Example Workflows

1. **Reducing TODOs**:
   ```bash
   ratchet --base-branch main --metric-cmd "grep -c TODO *.go" --lt
   ```

2. **Increasing Test Coverage**:
   ```bash
   ratchet --base-branch develop \
     --metric-cmd "go test -cover | grep coverage | awk '{print \$2}' | sed 's/%//'" \
     --gt
   ```

3. **Using Config File**:
   ```bash
   ratchet --config .ratchet.yml --verbose
   ```