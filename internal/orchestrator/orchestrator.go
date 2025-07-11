package orchestrator

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/tiernacity/ratchet/internal/config"
	"github.com/tiernacity/ratchet/internal/errors"
)

// orchestrator implements the Orchestrator interface
type orchestrator struct {
	git      GitOperations
	executor CommandExecutor
	parser   MetricParser
	reporter ProgressReporter
	tempDir  string
}

// New creates a new Orchestrator instance
func New(git GitOperations, executor CommandExecutor, parser MetricParser, reporter ProgressReporter, tempDir string) Orchestrator {
	return &orchestrator{
		git:      git,
		executor: executor,
		parser:   parser,
		reporter: reporter,
		tempDir:  tempDir,
	}
}

// Run executes the complete ratchet workflow
func (o *orchestrator) Run(ctx context.Context, cfg *config.Config) error {
	// Validate Git repository
	if err := o.git.IsGitRepository(); err != nil {
		return errors.WrapGitError("validation", fmt.Errorf("not in a git repository"))
	}

	// If no comparison operator is specified, just run the metric command in current directory
	if cfg.Operator == config.OpUnknown {
		return o.runNoComparison(ctx, cfg)
	}

	// Resolve base branch
	baseCommit, err := o.git.ResolveBranch(cfg.BaseBranch)
	if err != nil {
		return errors.WrapGitError("branch resolution",
			fmt.Errorf("branch '%s' not found", cfg.BaseBranch))
	}

	// Start progress reporting
	o.reporter.Start(cfg.BaseBranch, "HEAD", cfg.Verbose)

	// Create worktree for base branch
	worktreePath := filepath.Join(o.tempDir, "ratchet-base")
	if err := o.git.CreateWorktree(worktreePath, baseCommit); err != nil {
		return errors.WrapGitError("worktree creation",
			fmt.Errorf("failed to create worktree for %s: %w", cfg.BaseBranch, err))
	}

	// Ensure cleanup happens in all cases
	defer func() {
		if cleanErr := o.git.RemoveWorktree(worktreePath); cleanErr != nil {
			o.reporter.Error(fmt.Sprintf("failed to cleanup worktree: %v", cleanErr))
		}
	}()

	// Run metric collection for both branches
	baseMetric, err := o.runMetricSequence(ctx, cfg, worktreePath, cfg.BaseBranch)
	if err != nil {
		return err
	}

	headMetric, err := o.runMetricSequence(ctx, cfg, ".", "HEAD")
	if err != nil {
		return err
	}

	// Compare metrics
	return o.compareMetrics(baseMetric, headMetric, cfg.Operator, cfg.BaseBranch)
}

// runNoComparison runs just the metric command and outputs the result
func (o *orchestrator) runNoComparison(ctx context.Context, cfg *config.Config) error {
	// Run pre-command if specified
	if cfg.PreCmd != "" {
		if _, err := o.executor.Execute(ctx, ".", cfg.PreCmd); err != nil {
			return errors.NewPhaseErrorFromExecutorError("pre-command", "HEAD", err)
		}
	}

	// Run metric command
	output, err := o.executor.Execute(ctx, ".", cfg.MetricCmd)
	if err != nil {
		return errors.NewPhaseErrorFromExecutorError("metric command", "HEAD", err)
	}

	// Parse metric
	metric, err := o.parser.Parse(output)
	if err != nil {
		return err // parser already wraps with ParseError
	}

	// Run post-command if specified
	if cfg.PostCmd != "" {
		if _, err := o.executor.Execute(ctx, ".", cfg.PostCmd); err != nil {
			return errors.NewPhaseErrorFromExecutorError("post-command", "HEAD", err)
		}
	}

	// Report the metric value
	o.reporter.NoComparison(metric)
	return nil
}

// runMetricSequence executes pre/metric/post commands for a single branch
func (o *orchestrator) runMetricSequence(ctx context.Context, cfg *config.Config, dir, branchName string) (float64, error) {
	// Run pre-command if specified
	if cfg.PreCmd != "" {
		o.reporter.UpdateBranch(branchName, "pre", false)
		if _, err := o.executor.Execute(ctx, dir, cfg.PreCmd); err != nil {
			return 0, errors.NewPhaseErrorFromExecutorError("pre-command", branchName, err)
		}
		o.reporter.UpdateBranch(branchName, "pre", true)
	}

	// Run metric command
	o.reporter.UpdateBranch(branchName, "metric", false)
	output, err := o.executor.Execute(ctx, dir, cfg.MetricCmd)
	if err != nil {
		return 0, errors.NewPhaseErrorFromExecutorError("metric command", branchName, err)
	}

	// Parse metric
	metric, err := o.parser.Parse(output)
	if err != nil {
		return 0, err // parser already returns ParseError
	}
	o.reporter.UpdateBranch(branchName, "metric", true)

	// Run post-command if specified
	if cfg.PostCmd != "" {
		o.reporter.UpdateBranch(branchName, "post", false)
		if _, err := o.executor.Execute(ctx, dir, cfg.PostCmd); err != nil {
			return 0, errors.NewPhaseErrorFromExecutorError("post-command", branchName, err)
		}
		o.reporter.UpdateBranch(branchName, "post", true)
	}

	return metric, nil
}

// compareMetrics compares the head and base metrics according to the operator
func (o *orchestrator) compareMetrics(baseMetric, headMetric float64, op config.ComparisonOperator, baseBranch string) error {
	passed := false

	switch op {
	case config.OpGreaterThan:
		passed = headMetric > baseMetric
	case config.OpGreaterThanOrEqual:
		passed = headMetric >= baseMetric
	case config.OpEqual:
		passed = headMetric == baseMetric
	case config.OpLessThanOrEqual:
		passed = headMetric <= baseMetric
	case config.OpLessThan:
		passed = headMetric < baseMetric
	}

	if passed {
		o.reporter.Success(headMetric, baseMetric, op.HumanString(), baseBranch)
		return nil
	}

	return errors.NewMetricTestError(headMetric, baseMetric, op.HumanString(), baseBranch)
}

// FormatMetric formats a metric value for display, showing integers as integers
func FormatMetric(value float64) string {
	// Check if the value is effectively an integer
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}
