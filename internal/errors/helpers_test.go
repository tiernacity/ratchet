package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCreationHelpers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{
			name:     "metric test error",
			err:      NewMetricTestError(5.0, 10.0, "greater than", "main"),
			wantCode: 1,
			wantMsg:  "HEAD metric (5) is NOT greater than main (10)",
		},
		{
			name:     "execution error",
			err:      WrapExecutionError("setup", fmt.Errorf("command failed")),
			wantCode: 2,
			wantMsg:  "setup: command failed",
		},
		{
			name:     "validation error",
			err:      NewValidationError("metric-cmd", "cannot be empty"),
			wantCode: 2,
			wantMsg:  "validation error: metric-cmd: cannot be empty",
		},
		{
			name:     "git error",
			err:      WrapGitError("checkout", fmt.Errorf("branch not found")),
			wantCode: 2,
			wantMsg:  "git checkout: branch not found",
		},
		{
			name:     "command error",
			err:      WrapCommandError("echo test", fmt.Errorf("exit status 1")),
			wantCode: 2,
			wantMsg:  "command 'echo test' failed: exit status 1",
		},
		{
			name:     "parse error",
			err:      WrapParseError("hello", fmt.Errorf("not a number")),
			wantCode: 2,
			wantMsg:  "invalid metric output 'hello': not a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
			assert.Equal(t, tt.wantCode, GetExitCode(tt.err))
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	metricErr := NewMetricTestError(1, 2, "greater than", "main")
	execErr := WrapExecutionError("test", fmt.Errorf("failed"))
	validErr := NewValidationError("field", "message")
	gitErr := WrapGitError("test", fmt.Errorf("failed"))
	cmdErr := WrapCommandError("test", fmt.Errorf("failed"))
	parseErr := WrapParseError("test", fmt.Errorf("failed"))

	// Test metric error detection
	assert.True(t, IsMetricTestError(metricErr))
	assert.False(t, IsMetricTestError(execErr))
	assert.False(t, IsMetricTestError(validErr))

	// Test execution error detection
	assert.True(t, IsExecutionError(execErr))
	assert.False(t, IsExecutionError(metricErr))
	assert.False(t, IsExecutionError(validErr))

	// Test validation error detection
	assert.True(t, IsValidationError(validErr))
	assert.False(t, IsValidationError(metricErr))
	assert.False(t, IsValidationError(execErr))

	// Test git error detection
	assert.True(t, IsGitError(gitErr))
	assert.False(t, IsGitError(metricErr))
	assert.False(t, IsGitError(execErr))

	// Test command error detection
	assert.True(t, IsCommandError(cmdErr))
	assert.False(t, IsCommandError(metricErr))
	assert.False(t, IsCommandError(execErr))

	// Test parse error detection
	assert.True(t, IsParseError(parseErr))
	assert.False(t, IsParseError(metricErr))
	assert.False(t, IsParseError(execErr))
}

func TestWrapperHelpers(t *testing.T) {
	t.Run("wrap execution error with nil", func(t *testing.T) {
		err := WrapExecutionError("test", nil)
		assert.Nil(t, err)
	})

	t.Run("wrap git error with nil", func(t *testing.T) {
		err := WrapGitError("test", nil)
		assert.Nil(t, err)
	})

	t.Run("wrap command error with nil", func(t *testing.T) {
		err := WrapCommandError("test", nil)
		assert.Nil(t, err)
	})

	t.Run("wrap parse error with nil", func(t *testing.T) {
		err := WrapParseError("test", nil)
		assert.Nil(t, err)
	})
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil error",
			err:  nil,
			want: 0,
		},
		{
			name: "metric test error",
			err:  NewMetricTestError(1, 2, ">", "main"),
			want: 1,
		},
		{
			name: "execution error",
			err:  WrapExecutionError("test", fmt.Errorf("failed")),
			want: 2,
		},
		{
			name: "standard error",
			err:  fmt.Errorf("generic error"),
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetExitCode(tt.err))
		})
	}
}
