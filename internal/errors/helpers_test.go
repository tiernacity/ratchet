package errors

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
		{
			name:     "cancelled command error",
			err:      WrapCommandError("sleep 3", fmt.Errorf("cancelled")),
			wantCode: 2,
			wantMsg:  "cancelled",
		},
		{
			name:     "phase error",
			err:      NewPhaseError("pre-command", "main", "exit code 1"),
			wantCode: 2,
			wantMsg:  "pre-command in main: exit code 1",
		},
		{
			name:     "cancelled error",
			err:      NewCancelledError(),
			wantCode: 2,
			wantMsg:  "cancelled",
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
	validErr := NewValidationError("field", "message")
	gitErr := WrapGitError("test", fmt.Errorf("failed"))
	cmdErr := WrapCommandError("test", fmt.Errorf("failed"))
	parseErr := WrapParseError("test", fmt.Errorf("failed"))

	// Test metric error detection
	assert.True(t, IsMetricTestError(metricErr))
	assert.False(t, IsMetricTestError(validErr))

	// Test validation error detection
	assert.True(t, IsValidationError(validErr))
	assert.False(t, IsValidationError(metricErr))

	// Test git error detection
	assert.True(t, IsGitError(gitErr))
	assert.False(t, IsGitError(metricErr))

	// Test command error detection
	assert.True(t, IsCommandError(cmdErr))
	assert.False(t, IsCommandError(metricErr))

	// Test parse error detection
	assert.True(t, IsParseError(parseErr))
	assert.False(t, IsParseError(metricErr))
}

func TestWrapperHelpers(t *testing.T) {

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

func TestExtractConciseReason(t *testing.T) {
	// Create a mock process state for ExitError
	processState := &os.ProcessState{}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "exit error",
			err:  &exec.ExitError{ProcessState: processState},
			want: "exit code " + processState.String(),
		},
		{
			name: "exec error not found",
			err:  &exec.Error{Name: "npm", Err: exec.ErrNotFound},
			want: "npm not found",
		},
		{
			name: "exec error other",
			err:  &exec.Error{Name: "test", Err: fmt.Errorf("permission denied")},
			want: "failed to start test: permission denied",
		},
		{
			name: "syscall ENOENT",
			err:  syscall.ENOENT,
			want: "command not found",
		},
		{
			name: "syscall EACCES",
			err:  syscall.EACCES,
			want: "permission denied",
		},
		{
			name: "unknown error",
			err:  fmt.Errorf("some other error"),
			want: "some other error",
		},
		{
			name: "long error message",
			err:  fmt.Errorf("this is a very long error message that should be truncated because it exceeds the maximum length and goes on and on and on and on and on"),
			want: "this is a very long error message that should be truncated because it exceeds the maximum length ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractConciseReason(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewPhaseErrorFromExecutorError(t *testing.T) {
	// Create mock process state for ExitError
	processState := &os.ProcessState{}

	tests := []struct {
		name     string
		phase    string
		branch   string
		err      error
		wantType string
		wantMsg  string
	}{
		{
			name:     "cancellation returns CancelledError",
			phase:    "pre-command",
			branch:   "main",
			err:      fmt.Errorf("cancelled"),
			wantType: "*errors.CancelledError",
			wantMsg:  "cancelled",
		},
		{
			name:     "exit error returns PhaseError",
			phase:    "metric command",
			branch:   "origin/main",
			err:      &exec.ExitError{ProcessState: processState},
			wantType: "*errors.PhaseError",
			wantMsg:  "metric command in origin/main: exit code " + processState.String(),
		},
		{
			name:     "exec error not found returns PhaseError",
			phase:    "pre-command",
			branch:   "main",
			err:      &exec.Error{Name: "npm", Err: exec.ErrNotFound},
			wantType: "*errors.PhaseError",
			wantMsg:  "pre-command in main: npm not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPhaseErrorFromExecutorError(tt.phase, tt.branch, tt.err)
			assert.Equal(t, tt.wantMsg, got.Error())
			assert.Contains(t, fmt.Sprintf("%T", got), tt.wantType)
		})
	}
}
