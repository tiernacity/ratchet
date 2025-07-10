# Metric Parsing Feature

## Overview
The Metric Parsing module is responsible for extracting numeric values from command output. It validates that the output contains exactly one numeric value and converts it to a format suitable for comparison.

## Interface

```go
package parser

type Parser interface {
    // Parse extracts a numeric value from the given output string
    Parse(output string) (float64, error)
}
```

## Implementation Details

### Structure
```go
type parserImpl struct{}

func New() Parser {
    return &parserImpl{}
}
```

### Parse Method
```go
func (p *parserImpl) Parse(output string) (float64, error) {
    // Trim whitespace
    output = strings.TrimSpace(output)
    
    if output == "" {
        return 0, fmt.Errorf("metric command produced no output")
    }
    
    // First try to parse as integer
    if intVal, err := strconv.ParseInt(output, 10, 64); err == nil {
        return float64(intVal), nil
    }
    
    // Then try to parse as float
    if floatVal, err := strconv.ParseFloat(output, 64); err == nil {
        return floatVal, nil
    }
    
    // If neither works, provide helpful error
    return 0, fmt.Errorf("metric command output '%s' is not a valid number", output)
}
```

## Validation Rules

### Valid Outputs
1. **Integers**: `"42"`, `"0"`, `"-17"`
2. **Floats**: `"3.14"`, `"0.0"`, `"-2.5"`
3. **Scientific Notation**: `"1.23e4"`, `"5e-3"`
4. **With Whitespace**: `"  42  \n"` → `42`

### Invalid Outputs
1. **Empty Output**: `""` → Error
2. **Non-Numeric**: `"hello"` → Error
3. **Multiple Numbers**: `"1 2 3"` → Error
4. **Mixed Content**: `"42 tests"` → Error
5. **Special Values**: `"NaN"`, `"Inf"` → Error

## Advanced Parsing

### Extracting Numbers from Text (Future Enhancement)
```go
type AdvancedParser interface {
    // ParseWithPattern extracts a number using a regex pattern
    ParseWithPattern(output, pattern string) (float64, error)
    
    // ParseLast extracts the last number in the output
    ParseLast(output string) (float64, error)
    
    // ParseNth extracts the nth number in the output
    ParseNth(output string, n int) (float64, error)
}
```

### Example Advanced Implementation
```go
func (p *parserImpl) ParseWithPattern(output, pattern string) (float64, error) {
    re, err := regexp.Compile(pattern)
    if err != nil {
        return 0, fmt.Errorf("invalid pattern: %w", err)
    }
    
    matches := re.FindStringSubmatch(output)
    if len(matches) < 2 {
        return 0, fmt.Errorf("pattern '%s' did not match any number", pattern)
    }
    
    return p.Parse(matches[1])
}
```

## Error Handling

### Error Messages
Provide clear, actionable error messages:

```go
func (p *parserImpl) Parse(output string) (float64, error) {
    output = strings.TrimSpace(output)
    
    if output == "" {
        return 0, fmt.Errorf("metric command produced no output. " +
            "Ensure your command outputs a single number")
    }
    
    // Check for common issues
    if strings.Contains(output, " ") {
        numbers := extractNumbers(output)
        if len(numbers) > 1 {
            return 0, fmt.Errorf("metric command output contains multiple numbers: %v. " +
                "The command must output exactly one number", numbers)
        }
    }
    
    // Try parsing
    val, err := strconv.ParseFloat(output, 64)
    if err != nil {
        // Provide context-specific error
        if strings.Contains(output, "%") {
            return 0, fmt.Errorf("output '%s' contains '%'. " +
                "Remove percentage signs from the output", output)
        }
        
        return 0, fmt.Errorf("output '%s' is not a valid number. " +
            "Ensure the command outputs only a numeric value", output)
    }
    
    return val, nil
}
```

## Testing Strategy

### Unit Tests
```go
func TestParser_Parse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    float64
        wantErr string
    }{
        {
            name:  "integer",
            input: "42",
            want:  42.0,
        },
        {
            name:  "float",
            input: "3.14",
            want:  3.14,
        },
        {
            name:  "negative",
            input: "-17",
            want:  -17.0,
        },
        {
            name:  "with whitespace",
            input: "  100  \n",
            want:  100.0,
        },
        {
            name:  "scientific notation",
            input: "1.23e4",
            want:  12300.0,
        },
        {
            name:    "empty output",
            input:   "",
            wantErr: "produced no output",
        },
        {
            name:    "non-numeric",
            input:   "hello",
            wantErr: "not a valid number",
        },
        {
            name:    "multiple numbers",
            input:   "1 2 3",
            wantErr: "multiple numbers",
        },
    }
    
    parser := New()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.Parse(tt.input)
            
            if tt.wantErr != "" {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.wantErr)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Property-Based Tests
```go
func TestParser_ParseProperty(t *testing.T) {
    parser := New()
    
    // Test that all valid integers can be parsed
    quick.Check(func(n int64) bool {
        input := strconv.FormatInt(n, 10)
        result, err := parser.Parse(input)
        return err == nil && result == float64(n)
    }, nil)
    
    // Test that all valid floats can be parsed
    quick.Check(func(f float64) bool {
        if math.IsNaN(f) || math.IsInf(f, 0) {
            return true // Skip special values
        }
        input := strconv.FormatFloat(f, 'f', -1, 64)
        result, err := parser.Parse(input)
        return err == nil && math.Abs(result-f) < 0.0001
    }, nil)
}
```

## Common Use Cases

### Test Coverage
```bash
# Command output: "coverage: 85.5%"
# Need to extract: 85.5
metric-cmd: "go test -cover | grep coverage | awk '{print $2}' | sed 's/%//'"
```

### Line Count
```bash
# Command output: "1234"
# Direct numeric output
metric-cmd: "wc -l *.go | tail -1 | awk '{print $1}'"
```

### TODO Count
```bash
# Command output: "7"
# Direct numeric output
metric-cmd: "grep -c TODO *.go"
```

### Build Time
```bash
# Command output: "45.3"
# Time in seconds
metric-cmd: "time -p make build 2>&1 | grep real | awk '{print $2}'"
```

## Performance Considerations

1. **Parsing Speed**: Use strconv functions for optimal performance
2. **Memory Usage**: Avoid creating unnecessary string copies
3. **Regex Usage**: Compile regex patterns once if used repeatedly
4. **Input Size**: Handle large outputs gracefully

## Internationalization

### Number Formats
Currently supports standard numeric formats. Future considerations:
- Locale-specific decimal separators (`,` vs `.`)
- Thousand separators
- Different numeric systems

```go
// Future enhancement
type LocalizedParser interface {
    ParseWithLocale(output string, locale string) (float64, error)
}
```

## Integration with Commands

### Best Practices for Metric Commands
1. **Output only the number**: Avoid additional text
2. **Use consistent format**: Always use same decimal places
3. **Handle errors gracefully**: Output 0 or exit with error
4. **Avoid percentages**: Strip % signs in the command
5. **Single line output**: Ensure only one line is output

### Command Examples
```bash
# Good: Clean numeric output
echo "42"

# Bad: Multiple numbers
echo "Test results: 10 passed, 2 failed"

# Good: Processed to single number
echo "Test results: 10 passed, 2 failed" | awk '{print $3}'

# Bad: Percentage sign included
echo "85.5%"

# Good: Percentage sign removed
echo "85.5%" | sed 's/%//'
```

## Future Enhancements

1. **Pattern Matching**: Support extracting numbers with patterns
2. **Unit Support**: Parse and convert units (e.g., "5MB" → 5242880)
3. **Expression Evaluation**: Support simple math expressions
4. **Multiple Metrics**: Parse multiple metrics from single output
5. **Custom Parsers**: Allow users to provide custom parsing logic