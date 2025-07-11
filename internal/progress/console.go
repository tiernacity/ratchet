package progress

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ProgressLine tracks the status of commands for a single branch
type ProgressLine struct {
	branch string
	phases map[string]bool // phase name -> completed
	order  []string        // ordered list of phases
}

// consoleReporter implements the Reporter interface for console output
type consoleReporter struct {
	output      io.Writer
	errOut      io.Writer
	verbose     bool
	lines       map[string]*ProgressLine // branch name -> progress line
	branchOrder []string                 // ordered list of branches
	linesShown  bool                     // whether we've shown the initial lines
}

// NewConsole creates a new console progress reporter
func NewConsole() Reporter {
	return &consoleReporter{
		output: os.Stdout,
		errOut: os.Stderr,
		lines:  make(map[string]*ProgressLine),
	}
}

// NewConsoleWithWriters creates a console reporter with custom writers (for testing)
func NewConsoleWithWriters(output, errOut io.Writer) Reporter {
	return &consoleReporter{
		output: output,
		errOut: errOut,
		lines:  make(map[string]*ProgressLine),
	}
}

// Start begins progress reporting
func (c *consoleReporter) Start(baseRef, headRef string, phases []string, verbose bool) {
	c.verbose = verbose
	if !verbose {
		return
	}

	// Initialize progress lines for both branches
	c.branchOrder = []string{baseRef, headRef}
	for _, branch := range c.branchOrder {
		c.lines[branch] = &ProgressLine{
			branch: branch,
			phases: make(map[string]bool),
			order:  phases,
		}
		// Initialize all phases as incomplete
		for _, phase := range phases {
			c.lines[branch].phases[phase] = false
		}
	}

	c.linesShown = false
}

// UpdateBranch updates progress for a branch and phase
func (c *consoleReporter) UpdateBranch(branch, phase string, completed bool) {
	if !c.verbose {
		return
	}

	line, exists := c.lines[branch]
	if !exists {
		return
	}

	// Update the phase status
	line.phases[phase] = completed

	// Show initial lines if not yet shown
	if !c.linesShown {
		for _, b := range c.branchOrder {
			_, _ = fmt.Fprintf(c.output, "%s\n", c.formatProgressLine(c.lines[b]))
		}
		c.linesShown = true
		return
	}

	// Find which line to update (1-based index from current cursor position)
	lineIndex := 0
	for i, b := range c.branchOrder {
		if b == branch {
			lineIndex = len(c.branchOrder) - i
			break
		}
	}

	if lineIndex > 0 {
		// Move cursor up to the correct line, clear it, and rewrite
		_, _ = fmt.Fprintf(c.output, "\033[%dA\r\033[K%s\033[%dB", lineIndex, c.formatProgressLine(line), lineIndex)
	}
}

// Complete finishes progress reporting by adding a blank line
func (c *consoleReporter) Complete() {
	if c.verbose && len(c.lines) > 0 {
		_, _ = fmt.Fprintf(c.output, "\n")
	}
}

// formatProgressLine formats a progress line for display
func (c *consoleReporter) formatProgressLine(line *ProgressLine) string {
	parts := make([]string, 0, len(line.order))
	for _, phase := range line.order {
		checkbox := "[ ]"
		if line.phases[phase] {
			checkbox = "[x]"
		}
		parts = append(parts, fmt.Sprintf("%s %s", phase, checkbox))
	}
	return fmt.Sprintf("%-12s %s", line.branch+":", strings.Join(parts, " ; "))
}

// Success reports a successful metric test
func (c *consoleReporter) Success(current, base float64, operator, branch string) {
	_, _ = fmt.Fprintf(c.output, "Success: HEAD metric (%s) is %s %s (%s)\n",
		formatMetric(current), operator, branch, formatMetric(base))
}

// Error reports an error message
func (c *consoleReporter) Error(message string) {
	_, _ = fmt.Fprintf(c.errOut, "Error: %s\n", message)
}

// NoComparison reports a metric value when no comparison is performed
func (c *consoleReporter) NoComparison(value float64) {
	_, _ = fmt.Fprintf(c.output, "%s\n", formatMetric(value))
}

// Info reports an informational message
func (c *consoleReporter) Info(message string) {
	_, _ = fmt.Fprintf(c.errOut, "%s\n", message)
}

// formatMetric formats a metric value for display
func formatMetric(value float64) string {
	// Check if the value is effectively an integer
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}
