# Output Format Specification

This document defines the exact output formatting requirements for consistent user experience across all ratchet operations.

## Output Streams

### stdout
- Progress indicators (verbose mode only)
- Success/failure status messages
- Metric values (when no comparison is performed)

### stderr
- Error messages
- Help text (for configuration errors only)

## Progress Display (Verbose Mode Only)

### Format
Display one line per branch with checkboxes showing command execution progress:

```
origin/main: metric [ ]
HEAD:        metric [ ]
```

Update checkboxes in-place as commands complete:
```
origin/main: metric [x]
HEAD:        metric [x]
```

### Multiple Commands
When pre/post commands are configured:
```
origin/main: pre [x] ; metric [x] ; post [ ]
HEAD:        pre [x] ; metric [ ] ; post [ ]
```

### Implementation
- Branch names left-aligned, padded to 12 characters
- Use `[ ]` for pending, `[x]` for complete
- Separate commands with ` ; `
- Update lines in-place using ANSI escape sequences

## Success Messages

### Standard Mode
```
HEAD metric (5) is greater than origin/main (4)
```

### No Comparison Mode
When no operator is specified, output only the metric value:
```
42
```

## Failure Messages

### Metric Test Failures
```
HEAD metric (5) is NOT greater than origin/main (6)
```

### Format Rules
- Use "is NOT {operator}" for failed comparisons
- Include actual metric values with appropriate precision
- Display integers as integers, floats with decimals

## Error Message Format

### General Rules
- Single line messages
- No trailing punctuation
- Lowercase start (unless proper noun)
- Prefix with "Error: " for stderr output

### Error Categories

#### Git Errors
```
Error: not in a git repository
Error: branch 'feature-xyz' not found
Error: failed to create worktree for main
```

#### Command Errors
```
Error: command 'npm test' not found
Error: metric command failed with exit code 1
Error: pre-command in main: exit code 1
```

#### Configuration Errors
```
Error: metric command is required
Error: multiple comparison operators specified
Error: config file not found
```

#### Parse Errors
```
Error: invalid metric output 'hello': no numeric value found
```

## ANSI Escape Sequences

Use for in-place progress updates in verbose mode:

```go
const (
    clearLine = "\r\033[K"     // Clear current line
    moveUp    = "\033[1A"      // Move cursor up one line
)
```

## Platform Considerations

### Terminal Detection
- Detect if output is to a terminal vs pipe/file
- Disable ANSI codes for non-interactive output
- Respect `NO_COLOR` environment variable

### CI Environment
- Detect CI environment variables
- Adjust output for non-interactive terminals
- Use appropriate line endings (CRLF on Windows)

## Output Examples

### Successful Run (Standard)
```bash
$ ratchet --lt main "grep -c TODO *.go"
HEAD metric (5) is less than main (8)
```

### Successful Run (Verbose)
```bash
$ ratchet --lt main --verbose "grep -c TODO *.go"
main:        metric [x]
HEAD:        metric [x]

HEAD metric (5) is less than main (8)
```

### Failed Run
```bash
$ ratchet --lt main "grep -c TODO *.go"
HEAD metric (10) is NOT less than main (8)
```

### Command Failure
```bash
$ ratchet --lt main "nonexistent-command"
Error: command 'nonexistent-command' not found
```

### No Comparison
```bash
$ ratchet "grep -c TODO *.go"
5
```

## Testing Output Format

### Capture Patterns
```go
// Capture stdout/stderr for testing
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
    // Implementation
}

// Test output format
func TestOutputFormat(t *testing.T) {
    stdout, stderr := captureOutput(t, func() {
        // Run ratchet command
    })
    
    expected := "HEAD metric (10) is greater than main (5)"
    assert.Contains(t, stdout, expected)
    assert.Empty(t, stderr)
}
```

### Mock Progress Reporter
```go
type mockProgressReporter struct {
    lines []string
}

func (m *mockProgressReporter) UpdateProgress(branch, line string) {
    m.lines = append(m.lines, fmt.Sprintf("%s: %s", branch, line))
}
```

## Consistency Rules

1. **Metric Precision**: Display integers as integers, floats with minimal necessary precision
2. **Branch Names**: Use exactly as provided by user (don't normalize origin/main vs main)
3. **Command Names**: Show actual commands in error messages, not simplified names
4. **Error Context**: Include relevant context (branch, command, phase) but avoid redundancy
5. **Status Indicators**: Use consistent checkbox format `[x]` and `[ ]`

This specification ensures consistent user experience and makes the tool's behavior predictable across different environments and use cases.