package progress

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

// consoleReporter implements the Reporter interface for console output
type consoleReporter struct {
	output  io.Writer
	errOut  io.Writer
	verbose bool
}

// NewConsole creates a new console progress reporter
func NewConsole() Reporter {
	return &consoleReporter{
		output: os.Stdout,
		errOut: os.Stderr,
	}
}

// NewConsoleWithWriters creates a console reporter with custom writers (for testing)
func NewConsoleWithWriters(output, errOut io.Writer) Reporter {
	return &consoleReporter{
		output: output,
		errOut: errOut,
	}
}

// Start begins progress reporting
func (c *consoleReporter) Start(baseRef, headRef string, verbose bool) {
	c.verbose = verbose
	if verbose {
		fmt.Fprintf(c.output, "Comparing %s to %s\n", headRef, baseRef)
	}
}

// UpdateBranch updates progress for a branch and phase
func (c *consoleReporter) UpdateBranch(branch, phase string, completed bool) {
	if !c.verbose {
		return
	}

	status := "[ ]"
	if completed {
		status = "[x]"
	}
	fmt.Fprintf(c.output, "  %s: %s %s\n", branch, phase, status)
}

// Success reports a successful metric test
func (c *consoleReporter) Success(current, base float64, operator, branch string) {
	fmt.Fprintf(c.output, "Success: HEAD metric (%s) is %s %s (%s)\n",
		formatMetric(current), operator, branch, formatMetric(base))
}

// Error reports an error message
func (c *consoleReporter) Error(message string) {
	fmt.Fprintf(c.errOut, "Error: %s\n", message)
}

// NoComparison reports a metric value when no comparison is performed
func (c *consoleReporter) NoComparison(value float64) {
	fmt.Fprintf(c.output, "%s\n", formatMetric(value))
}

// Info reports an informational message
func (c *consoleReporter) Info(message string) {
	fmt.Fprintf(c.errOut, "%s\n", message)
}

// formatMetric formats a metric value for display
func formatMetric(value float64) string {
	// Check if the value is effectively an integer
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}
