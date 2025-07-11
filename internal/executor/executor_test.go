package executor

import (
	"context"
	"errors"
	osexec "os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_Execute_Success(t *testing.T) {
	exec := New()

	// Use a simple command that works on all platforms
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	} else {
		cmd = "echo hello"
	}

	output, err := exec.Execute(context.Background(), ".", cmd)

	assert.NoError(t, err)
	assert.Contains(t, output, "hello")
}

func TestExecutor_Execute_WithContext(t *testing.T) {
	exec := New()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use a command that should timeout
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "timeout /t 1"
	} else {
		cmd = "sleep 1"
	}

	_, err := exec.Execute(ctx, ".", cmd)

	assert.Error(t, err)
	// Should be a cancellation error
	assert.True(t,
		strings.Contains(err.Error(), "context") ||
			strings.Contains(err.Error(), "cancel") ||
			strings.Contains(err.Error(), "timeout"),
		"Expected context cancellation error, got: %s", err.Error())
}

func TestExecutor_Execute_InvalidCommand(t *testing.T) {
	exec := New()

	// Use a command that definitely doesn't exist
	output, err := exec.Execute(context.Background(), ".", "this-command-does-not-exist-12345")

	assert.Error(t, err)
	assert.Empty(t, output)
	
	// Check that it's an exec.ExitError with exit code 127 (command not found)
	var exitErr *osexec.ExitError
	assert.True(t, 
		errors.As(err, &exitErr) && exitErr.ExitCode() == 127, 
		"Expected exec.ExitError with exit code 127 for command not found")
}

func TestExecutor_Execute_InvalidDirectory(t *testing.T) {
	exec := New()

	// Try to execute in a directory that doesn't exist
	_, err := exec.Execute(context.Background(), "/this/directory/does/not/exist", "echo test")

	assert.Error(t, err)
}

func TestExecutor_Execute_CommandFailure(t *testing.T) {
	exec := New()

	// Use a command that will fail
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "exit 1"
	} else {
		cmd = "false"
	}

	_, err := exec.Execute(context.Background(), ".", cmd)

	assert.Error(t, err)
	
	// Check that it's an exec.ExitError with exit code 1
	var exitErr *osexec.ExitError
	assert.True(t, 
		errors.As(err, &exitErr) && exitErr.ExitCode() == 1, 
		"Expected exec.ExitError with exit code 1 for failed command")
}
