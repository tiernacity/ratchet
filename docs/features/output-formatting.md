# Output Formatting Feature

## Overview
This document specifies the exact output formatting requirements for Ratchet, including progress display, success/failure messages, and error reporting. All output must follow these specifications to ensure consistent user experience.

## Output Streams

### stdout
- Progress indicators (checkboxes)
- Success messages
- Metric values (when no comparison is performed)
- Verbose output (when --verbose flag is used)

### stderr
- Error messages
- Failure status messages
- Usage help (only for invalid arguments)

## Progress Display

### Standard Mode
Display one line per branch with updating checkboxes:

```
origin/main: metric [ ]
HEAD:        metric [ ]
```

As tasks complete, update checkboxes in-place:
```
origin/main: metric [x]
HEAD:        metric [x]
```

### Verbose Mode
Include all commands in the progress line:

```
origin/main: pre [ ] ; metric [ ] ; post [ ]
HEAD:        pre [ ] ; metric [ ] ; post [ ]
```

Update each checkbox as the corresponding command completes:
```
origin/main: pre [x] ; metric [x] ; post [x]
HEAD:        pre [x] ; metric [ ] ; post [ ]
```

## Success Output

### Standard Success
```
$ ratchet --gt origin/main './my-metric-test.sh'

HEAD metric (5) is greater than origin/main (4)
Succeeded
```

### Verbose Success
```
$ ratchet --gt origin/main --pre './my-pre.sh' --post './my-post.sh' './my-metric-test.sh'
origin/main: pre [x] ; metric [x] ; post [x]
HEAD:        pre [x] ; metric [x] ; post [x]

HEAD metric (5) is greater than origin/main (4)
Succeeded
```

## Failure Output

### Metric Test Failure
Progress on stdout, status and failure on stderr:

```
$ ratchet --gt origin/main './my-metric-test.sh'
origin/main: metric [x]
HEAD:        metric [x]

HEAD metric (5) is NOT greater than origin/main (6)
Failed
```

### Command Execution Failure
```
$ ratchet --gt origin/main --pre './my-pre.sh' --post './my-post.sh' './my-metric-test.sh'
origin/main: pre [x] ; metric [x] ; post [x]
HEAD:        pre [x] ; metric [ ] ; post [ ]

Command './my-metric-test.sh' failed in HEAD
Failed
```

## No Comparison Mode
When no comparison operator is specified, output only the metric value:

```
$ ratchet './count-todos.sh'
42
```

## Implementation Details

### Progress Line Builder
```go
type ProgressLine struct {
    branch   string
    commands []CommandStatus
}

type CommandStatus struct {
    name     string
    complete bool
}

func (p *ProgressLine) String() string {
    parts := []string{}
    for i, cmd := range p.commands {
        checkbox := "[ ]"
        if cmd.complete {
            checkbox = "[x]"
        }
        parts = append(parts, fmt.Sprintf("%s %s", cmd.name, checkbox))
    }
    
    return fmt.Sprintf("%-12s %s", p.branch+":", strings.Join(parts, " ; "))
}
```

### ANSI Escape Sequences
Use ANSI codes to update lines in-place:

```go
const (
    clearLine = "\r\033[K"
    moveUp    = "\033[1A"
)

func updateProgress(line string) {
    fmt.Printf("%s%s", clearLine, line)
}
```

### Output Formatting Rules

1. **Branch Names**: Left-aligned, padded to 12 characters
2. **Checkboxes**: Use `[ ]` for pending, `[x]` for complete
3. **Separators**: Use ` ; ` between commands in verbose mode
4. **Metric Values**: Display with appropriate precision (integers as integers, floats with decimals)
5. **Status Messages**: "Succeeded" or "Failed" on separate line

## Comparison Messages

### Success Messages
Format: `HEAD metric ({value}) is {operator} {branch} ({value})`

Examples:
- `HEAD metric (10) is greater than origin/main (5)`
- `HEAD metric (3.14) is equal to develop (3.14)`
- `HEAD metric (42) is less than or equal to main (50)`

### Failure Messages
Format: `HEAD metric ({value}) is NOT {operator} {branch} ({value})`

Examples:
- `HEAD metric (5) is NOT greater than origin/main (10)`
- `HEAD metric (3.14) is NOT equal to develop (2.71)`
- `HEAD metric (100) is NOT less than main (50)`

## Error Messages

### Git Errors
```
Error: not in a git repository
Error: branch 'feature-xyz' not found
Error: failed to create worktree: exit status 128
```

### Command Errors
```
Error: command not found: './missing-script.sh'
Error: pre-command failed: exit status 1
Error: metric command produced no output
```

### Parse Errors
```
Error: invalid metric output 'hello': not a number
Error: metric command output contains multiple numbers
```

## Special Cases

### Windows Compatibility
- Use appropriate line endings (CRLF on Windows)
- Handle paths with backslashes
- Account for cmd.exe output differences

### CI Environment Detection
- Detect CI environment variables
- Adjust output for non-interactive terminals
- Disable ANSI codes if not supported

### Terminal Width
- Assume minimum 80 character width
- Truncate long branch names if necessary
- Wrap long error messages appropriately

## Testing Output

### Mock Progress Reporter
```go
type mockProgressReporter struct {
    lines []string
}

func (m *mockProgressReporter) UpdateLine(branch, line string) {
    m.lines = append(m.lines, fmt.Sprintf("%s: %s", branch, line))
}
```

### Output Assertions
```go
func TestOutputFormat(t *testing.T) {
    expected := `origin/main: metric [x]
HEAD:        metric [x]

HEAD metric (10) is greater than origin/main (5)
Succeeded`
    
    assert.Equal(t, expected, capturedOutput)
}
```

## Future Enhancements

1. **Colored Output**: Green for success, red for failure
2. **Progress Bars**: For long-running commands
3. **JSON Output**: Machine-readable format for CI
4. **Quiet Mode**: Suppress all but essential output
5. **Interactive Mode**: Real-time updates with spinners