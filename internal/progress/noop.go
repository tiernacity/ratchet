package progress

// noopReporter is a no-operation implementation of Reporter for testing
type noopReporter struct{}

// NewNoop creates a new no-op progress reporter
func NewNoop() Reporter {
	return &noopReporter{}
}

// Start does nothing
func (n *noopReporter) Start(baseRef, headRef string, verbose bool) {}

// UpdateBranch does nothing
func (n *noopReporter) UpdateBranch(branch, phase string, completed bool) {}

// Success does nothing
func (n *noopReporter) Success(current, base float64, operator, branch string) {}

// Error does nothing
func (n *noopReporter) Error(message string) {}

// NoComparison does nothing
func (n *noopReporter) NoComparison(value float64) {}

// Info does nothing
func (n *noopReporter) Info(message string) {}
