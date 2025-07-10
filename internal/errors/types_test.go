package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricTestError(t *testing.T) {
	err := &MetricTestError{
		Current:  5.0,
		Base:     10.0,
		Operator: "greater than",
		Branch:   "main",
	}

	assert.Equal(t, "HEAD metric (5) is NOT greater than main (10)", err.Error())
	assert.Equal(t, 1, err.ExitCode())
}

func TestExecutionError(t *testing.T) {
	underlying := fmt.Errorf("command not found")
	err := &ExecutionError{
		Phase:   "setup",
		Wrapped: underlying,
	}

	assert.Equal(t, "setup: command not found", err.Error())
	assert.Equal(t, 2, err.ExitCode())
	assert.Equal(t, underlying, err.Unwrap())
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "metric-cmd",
		Message: "cannot be empty",
	}

	assert.Equal(t, "validation error: metric-cmd: cannot be empty", err.Error())
	assert.Equal(t, 2, err.ExitCode())
}

func TestGitError(t *testing.T) {
	underlying := fmt.Errorf("branch not found")
	err := &GitError{
		Operation: "checkout",
		Wrapped:   underlying,
	}

	assert.Equal(t, "git checkout: branch not found", err.Error())
	assert.Equal(t, 2, err.ExitCode())
	assert.Equal(t, underlying, err.Unwrap())
}

func TestCommandError(t *testing.T) {
	underlying := fmt.Errorf("exit status 1")
	err := &CommandError{
		Command: "echo test",
		Wrapped: underlying,
	}

	assert.Equal(t, "command 'echo test' failed: exit status 1", err.Error())
	assert.Equal(t, 2, err.ExitCode())
	assert.Equal(t, underlying, err.Unwrap())
}

func TestParseError(t *testing.T) {
	underlying := fmt.Errorf("not a number")
	err := &ParseError{
		Output:  "hello world",
		Wrapped: underlying,
	}

	assert.Equal(t, "invalid metric output 'hello world': not a number", err.Error())
	assert.Equal(t, 2, err.ExitCode())
	assert.Equal(t, underlying, err.Unwrap())
}

func TestFloatFormatting(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		base     float64
		expected string
	}{
		{
			name:     "integers display as integers",
			current:  5.0,
			base:     10.0,
			expected: "HEAD metric (5) is NOT greater than main (10)",
		},
		{
			name:     "floats display with decimals",
			current:  3.14159,
			base:     2.71828,
			expected: "HEAD metric (3.142) is NOT greater than main (2.718)",
		},
		{
			name:     "small numbers",
			current:  0.001,
			base:     0.002,
			expected: "HEAD metric (0.001) is NOT greater than main (0.002)",
		},
		{
			name:     "large numbers",
			current:  1000000.0,
			base:     999999.0,
			expected: "HEAD metric (1e+06) is NOT greater than main (1e+06)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &MetricTestError{
				Current:  tt.current,
				Base:     tt.base,
				Operator: "greater than",
				Branch:   "main",
			}
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}
