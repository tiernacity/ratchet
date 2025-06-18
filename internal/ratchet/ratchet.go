package ratchet

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tiernacity/ratchet/internal/executor"
	"github.com/tiernacity/ratchet/internal/git"
	"github.com/tiernacity/ratchet/internal/parser"
)

// ComparisonType represents the type of comparison to perform
type ComparisonType int

const (
	// NoComparison means just report the metric
	NoComparison ComparisonType = iota
	// LessThan means current < base
	LessThan
	// LessEqual means current <= base
	LessEqual
	// Equal means current == base
	Equal
	// GreaterEqual means current >= base
	GreaterEqual
	// GreaterThan means current > base
	GreaterThan
)

// Options contains the configuration for running ratchet
type Options struct {
	Metric         string         // Command to execute that outputs a number
	BaseRef        string         // Base branch/ref to compare against
	ComparisonType ComparisonType // Type of comparison to perform
	Pre            string         // Command to run before metric command
	Post           string         // Command to run after metric command
	Verbose        bool           // Show detailed output
}

func (ct ComparisonType) String() string {
	switch ct {
	case LessThan:
		return "less-than"
	case LessEqual:
		return "less-equal"
	case Equal:
		return "equal-to"
	case GreaterEqual:
		return "greater-equal"
	case GreaterThan:
		return "greater-than"
	default:
		return "unknown"
	}
}

// buildProgressLine creates a progress line with checkboxes
func buildProgressLine(branchName string, baseRef string, preCmd string, hasMetric bool, postCmd string, preComplete bool, metricComplete bool, postComplete bool) string {
	var parts []string

	if preCmd != "" {
		if preComplete {
			parts = append(parts, "pre [x]")
		} else {
			parts = append(parts, "pre [ ]")
		}
	}

	if hasMetric {
		if metricComplete {
			parts = append(parts, "metric [x]")
		} else {
			parts = append(parts, "metric [ ]")
		}
	}

	if postCmd != "" {
		if postComplete {
			parts = append(parts, "post [x]")
		} else {
			parts = append(parts, "post [ ]")
		}
	}

	// Create proper spacing - align based on the longer of baseRef or "HEAD"
	maxBranchLen := len(baseRef)
	if len("HEAD") > maxBranchLen {
		maxBranchLen = len("HEAD")
	}
	spacing := strings.Repeat(" ", maxBranchLen-len(branchName)+1)

	if len(parts) > 0 {
		return fmt.Sprintf("%s:%s%s", branchName, spacing, strings.Join(parts, " ; "))
	}
	return fmt.Sprintf("%s:%smetric [ ]", branchName, spacing)
}

func Run(opts Options) error {
	// Check if we're in a git repository
	if !git.IsGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	// Get current branch name for display
	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		currentBranch = "HEAD"
	}

	// Set up signal handling for graceful cleanup at the start
	var cleanup func()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		if cleanup != nil {
			cleanup()
		}
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Only create worktree if we need to compare
	var baseValue float64
	var baseProgressShown bool
	if opts.ComparisonType != NoComparison {
		// Ensure base branch exists
		if err := git.EnsureBranchExists(opts.BaseRef); err != nil {
			return fmt.Errorf("base branch '%s' not found", opts.BaseRef)
		}

		// Create temporary worktree for base branch
		worktreePath, cleanupFunc, err := git.CreateWorktree(opts.BaseRef)
		if err != nil {
			return fmt.Errorf("failed to create worktree for branch '%s'", opts.BaseRef)
		}
		cleanup = cleanupFunc
		defer cleanup()

		// Show initial progress line for base branch (only if verbose)
		if opts.Verbose {
			progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
			fmt.Print(progress)
			baseProgressShown = true
		}

		// Run pre command in base branch if specified
		preCompleted := true
		if opts.Pre != "" {
			if _, err := executor.Execute(opts.Pre, worktreePath); err != nil {
				// Show final progress state and exit (only if verbose)
				if opts.Verbose {
					progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
					fmt.Printf("\r%s\n", progress)
					progress = buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
					fmt.Printf("%s\n\n", progress)
				}
				fmt.Fprintf(os.Stderr, "Command '%s' failed in %s\n", opts.Pre, opts.BaseRef)
				fmt.Fprintln(os.Stderr, "Failed")
				return fmt.Errorf("metric test failed")
			}
			// Update progress to show pre command completed (only if verbose)
			if opts.Verbose {
				progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, true, false, false)
				fmt.Printf("\r%s", progress)
			}
		}

		// Execute command in base branch
		baseOutput, err := executor.Execute(opts.Metric, worktreePath)
		if err != nil {
			// Show final progress state and exit (only if verbose)
			if opts.Verbose {
				progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, false, false)
				fmt.Printf("\r%s\n", progress)
				progress = buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
				fmt.Printf("%s\n\n", progress)
			}
			fmt.Fprintf(os.Stderr, "Metric command '%s' failed in %s\n", opts.Metric, opts.BaseRef)
			fmt.Fprintln(os.Stderr, "Failed")
			return fmt.Errorf("metric test failed")
		}
		// Update progress to show metric command completed (only if verbose)
		if opts.Verbose {
			progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, false)
			fmt.Printf("\r%s", progress)
		}

		// Run post command in base branch if specified
		postCompleted := true
		if opts.Post != "" {
			if _, err := executor.Execute(opts.Post, worktreePath); err != nil {
				// Show final progress state and exit (only if verbose)
				if opts.Verbose {
					progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, false)
					fmt.Printf("\r%s\n", progress)
					progress = buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
					fmt.Printf("%s\n\n", progress)
				}
				fmt.Fprintf(os.Stderr, "Command '%s' failed in %s\n", opts.Post, opts.BaseRef)
				fmt.Fprintln(os.Stderr, "Failed")
				return fmt.Errorf("metric test failed")
			}
			// Update progress to show post command state (only if verbose)
			if opts.Verbose {
				progress := buildProgressLine(opts.BaseRef, opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, postCompleted)
				fmt.Printf("\r%s", progress)
			}
		}

		// Complete the base branch line (only if verbose)
		if baseProgressShown {
			fmt.Print("\n")
		}

		baseValue, err = parser.ParseNumber(baseOutput)
		if err != nil {
			return fmt.Errorf("command output from %s is not a number: '%s'", opts.BaseRef, baseOutput)
		}
	}

	// Show progress line for current branch (only if we're doing comparison and verbose)
	var currentProgressShown bool
	if opts.ComparisonType != NoComparison && opts.Verbose {
		progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
		fmt.Print(progress)
		currentProgressShown = true
	}

	// Run pre command in current branch if specified
	preCompleted := true
	if opts.Pre != "" {
		if _, err := executor.Execute(opts.Pre, ""); err != nil {
			if opts.ComparisonType != NoComparison && opts.Verbose {
				// Show final progress state and exit
				progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, false, false, false)
				fmt.Printf("\r%s\n\n", progress)
			}
			fmt.Fprintf(os.Stderr, "Command '%s' failed in %s\n", opts.Pre, currentBranch)
			fmt.Fprintln(os.Stderr, "Failed")
			return fmt.Errorf("metric test failed")
		}
		if opts.ComparisonType != NoComparison && opts.Verbose {
			// Update progress to show pre command completed
			progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, true, false, false)
			fmt.Printf("\r%s", progress)
		}
	}

	// Get current branch value
	currentOutput, err := executor.Execute(opts.Metric, "")
	if err != nil {
		if opts.ComparisonType != NoComparison && opts.Verbose {
			// Show final progress state and exit
			progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, false, false)
			fmt.Printf("\r%s\n\n", progress)
		}
		fmt.Fprintf(os.Stderr, "Metric command '%s' failed in %s\n", opts.Metric, currentBranch)
		fmt.Fprintln(os.Stderr, "Failed")
		return fmt.Errorf("metric test failed")
	}
	if opts.ComparisonType != NoComparison && opts.Verbose {
		// Update progress to show metric command completed
		progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, false)
		fmt.Printf("\r%s", progress)
	}

	// Run post command in current branch if specified
	postCompleted := true
	if opts.Post != "" {
		if _, err := executor.Execute(opts.Post, ""); err != nil {
			if opts.ComparisonType != NoComparison && opts.Verbose {
				// Show final progress state and exit
				progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, false)
				fmt.Printf("\r%s\n\n", progress)
			}
			fmt.Fprintf(os.Stderr, "Command '%s' failed in %s\n", opts.Post, currentBranch)
			fmt.Fprintln(os.Stderr, "Failed")
			return fmt.Errorf("metric test failed")
		}
		if opts.ComparisonType != NoComparison && opts.Verbose {
			// Update progress to show post command state
			progress := buildProgressLine("HEAD", opts.BaseRef, opts.Pre, true, opts.Post, preCompleted, true, postCompleted)
			fmt.Printf("\r%s", progress)
		}
	}

	// Complete the current branch line
	if currentProgressShown {
		fmt.Print("\n")
	}

	currentValue, err := parser.ParseNumber(currentOutput)
	if err != nil {
		return fmt.Errorf("command output from %s is not a number: '%s'", currentBranch, currentOutput)
	}

	// If no comparison, just output the metric
	if opts.ComparisonType == NoComparison {
		fmt.Printf("%g\n", currentValue)
		return nil
	}

	// Perform comparison based on type
	var passed bool
	var comparisonText string
	switch opts.ComparisonType {
	case LessThan:
		passed = currentValue < baseValue
		comparisonText = "less than"
	case LessEqual:
		passed = currentValue <= baseValue
		comparisonText = "less than or equal to"
	case Equal:
		passed = currentValue == baseValue
		comparisonText = "equal to"
	case GreaterEqual:
		passed = currentValue >= baseValue
		comparisonText = "greater than or equal to"
	case GreaterThan:
		passed = currentValue > baseValue
		comparisonText = "greater than"
	}

	if passed {
		// Only show detailed status line if verbose (for passing tests)
		if opts.Verbose {
			// Add blank line before result
			fmt.Println()
			fmt.Printf("%s metric (%g) is %s %s (%g)\n", currentBranch, currentValue, comparisonText, opts.BaseRef, baseValue)
		}
		fmt.Println("Succeeded")
		return nil
	}

	// Test failed - always show status line for failures
	// Add blank line before result if verbose (since progress lines were shown)
	if opts.Verbose {
		fmt.Println()
	}
	fmt.Fprintf(os.Stderr, "%s metric (%g) is NOT %s %s (%g)\n", currentBranch, currentValue, comparisonText, opts.BaseRef, baseValue)
	fmt.Fprintln(os.Stderr, "Failed")
	return fmt.Errorf("metric test failed")
}
