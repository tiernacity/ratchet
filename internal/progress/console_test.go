package progress

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleReporter_Success(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	reporter.Success(50.0, 42.0, "greater than", "main")

	result := output.String()
	assert.Contains(t, result, "HEAD metric (50) is greater than main (42)")
	assert.Empty(t, errOut.String())
}

func TestConsoleReporter_NoComparison(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	reporter.NoComparison(42.5)

	result := output.String()
	assert.Equal(t, "42.5\n", result)
	assert.Empty(t, errOut.String())
}

func TestConsoleReporter_Error(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	reporter.Error("Something went wrong")

	result := errOut.String()
	assert.Equal(t, "Error: Something went wrong\n", result)
	assert.Empty(t, output.String())
}

func TestConsoleReporter_Info(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	reporter.Info("This is info")

	result := errOut.String()
	assert.Equal(t, "This is info\n", result)
	assert.Empty(t, output.String())
}

func TestConsoleReporter_VerboseMode(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	// Start in verbose mode with multiple phases
	reporter.Start("main", "HEAD", []string{"pre", "metric", "post"}, true)
	reporter.UpdateBranch("main", "pre", true)
	reporter.UpdateBranch("HEAD", "metric", true)
	reporter.Complete()

	result := output.String()
	assert.Contains(t, result, "main:        pre [x] ; metric [ ] ; post [ ]")
	assert.Contains(t, result, "HEAD:        pre [ ] ; metric [x] ; post [ ]")
}

func TestConsoleReporter_NonVerboseMode(t *testing.T) {
	var output, errOut bytes.Buffer
	reporter := NewConsoleWithWriters(&output, &errOut)

	// Start in non-verbose mode
	reporter.Start("main", "HEAD", []string{"metric"}, false)
	reporter.UpdateBranch("main", "metric", false)
	reporter.UpdateBranch("main", "metric", true)

	result := output.String()
	// Should be empty since verbose is false
	assert.Empty(t, result)
}

func TestFormatMetric(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"integer", 42.0, "42"},
		{"float", 3.14, "3.14"},
		{"negative integer", -5.0, "-5"},
		{"negative float", -2.71, "-2.71"},
		{"zero", 0.0, "0"},
		{"scientific notation", 1e6, "1000000"},
		{"small decimal", 0.001, "0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMetric(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNoopReporter(t *testing.T) {
	// Test that noop reporter doesn't panic and has no side effects
	reporter := NewNoop()

	reporter.Start("main", "HEAD", []string{"metric"}, true)
	reporter.UpdateBranch("main", "metric", false)
	reporter.UpdateBranch("main", "metric", true)
	reporter.Complete()
	reporter.Success(50.0, 42.0, "greater than", "main")
	reporter.Error("test error")
	reporter.NoComparison(42.0)
	reporter.Info("test info")

	// If we get here without panicking, the test passes
	assert.True(t, true)
}
