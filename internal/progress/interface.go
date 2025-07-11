package progress

// Reporter defines the interface for reporting progress and results
type Reporter interface {
	// Start begins progress reporting for a comparison between base and head refs
	// phases indicates which commands will be run (e.g. ["pre", "metric", "post"])
	Start(baseRef, headRef string, phases []string, verbose bool)

	// UpdateBranch updates the progress for a specific branch and phase
	UpdateBranch(branch string, phase string, completed bool)

	// Complete finishes progress reporting (adds spacing after progress lines)
	Complete()

	// Success reports a successful metric test
	Success(current, base float64, operator, branch string)

	// Error reports an error message
	Error(message string)

	// NoComparison reports a metric value when no comparison is being performed
	NoComparison(value float64)

	// Info reports an informational message
	Info(message string)
}
