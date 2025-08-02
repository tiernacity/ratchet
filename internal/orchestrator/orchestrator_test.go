package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tiernacity/ratchet/internal/config"
	"github.com/tiernacity/ratchet/internal/errors"
)

// Mock implementations
type mockGitOperations struct {
	mock.Mock
}

func (m *mockGitOperations) IsGitRepository() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockGitOperations) CreateWorktree(path, branch string) error {
	args := m.Called(path, branch)
	return args.Error(0)
}

func (m *mockGitOperations) RemoveWorktree(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *mockGitOperations) ResolveBranch(branch string) (string, error) {
	args := m.Called(branch)
	return args.String(0), args.Error(1)
}

type mockCommandExecutor struct {
	mock.Mock
}

func (m *mockCommandExecutor) Execute(ctx context.Context, dir, command string) (string, error) {
	args := m.Called(ctx, dir, command)
	return args.String(0), args.Error(1)
}

type mockMetricParser struct {
	mock.Mock
}

func (m *mockMetricParser) Parse(output string) (float64, error) {
	args := m.Called(output)
	return args.Get(0).(float64), args.Error(1)
}

type mockProgressReporter struct {
	mock.Mock
}

func (m *mockProgressReporter) Start(baseRef, headRef string, phases []string, verbose bool) {
	m.Called(baseRef, headRef, phases, verbose)
}

func (m *mockProgressReporter) UpdateBranch(branch string, phase string, completed bool) {
	m.Called(branch, phase, completed)
}

func (m *mockProgressReporter) Complete() {
	m.Called()
}

func (m *mockProgressReporter) Success(current, base float64, operator, branch string) {
	m.Called(current, base, operator, branch)
}

func (m *mockProgressReporter) Error(message string) {
	m.Called(message)
}

func (m *mockProgressReporter) NoComparison(value float64) {
	m.Called(value)
}

func (m *mockProgressReporter) Info(message string) {
	m.Called(message)
}

func TestOrchestrator_Run_Success(t *testing.T) {
	// Setup mocks
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	// Configure expectations
	git.On("IsGitRepository").Return(nil)
	git.On("ResolveBranch", "main").Return("origin/main", nil)
	git.On("CreateWorktree", "/tmp/ratchet-base", "origin/main").Return(nil)
	git.On("RemoveWorktree", "/tmp/ratchet-base").Return(nil)

	reporter.On("Start", "main", "HEAD", []string{"metric"}, false)
	reporter.On("Complete")

	// Base branch execution
	reporter.On("UpdateBranch", "main", "metric", false)
	executor.On("Execute", mock.Anything, "/tmp/ratchet-base", "echo 42").Return("42", nil)
	parser.On("Parse", "42").Return(42.0, nil)
	reporter.On("UpdateBranch", "main", "metric", true)

	// HEAD branch execution
	reporter.On("UpdateBranch", "HEAD", "metric", false)
	executor.On("Execute", mock.Anything, ".", "echo 42").Return("50", nil)
	parser.On("Parse", "50").Return(50.0, nil)
	reporter.On("UpdateBranch", "HEAD", "metric", true)

	// Success reporting
	reporter.On("Success", 50.0, 42.0, "greater than", "main")

	cfg := &config.Config{
		BaseBranch: "main",
		Metric:     "echo 42",
		Operator:   config.OpGreaterThan,
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.NoError(t, err)
	git.AssertExpectations(t)
	executor.AssertExpectations(t)
	parser.AssertExpectations(t)
	reporter.AssertExpectations(t)
}

func TestOrchestrator_Run_MetricTestFailure(t *testing.T) {
	// Setup mocks
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	// Configure expectations
	git.On("IsGitRepository").Return(nil)
	git.On("ResolveBranch", "main").Return("origin/main", nil)
	git.On("CreateWorktree", "/tmp/ratchet-base", "origin/main").Return(nil)
	git.On("RemoveWorktree", "/tmp/ratchet-base").Return(nil)

	reporter.On("Start", "main", "HEAD", []string{"metric"}, false)
	reporter.On("Complete")

	// Base branch execution
	reporter.On("UpdateBranch", "main", "metric", false)
	executor.On("Execute", mock.Anything, "/tmp/ratchet-base", "echo 42").Return("50", nil)
	parser.On("Parse", "50").Return(50.0, nil)
	reporter.On("UpdateBranch", "main", "metric", true)

	// HEAD branch execution
	reporter.On("UpdateBranch", "HEAD", "metric", false)
	executor.On("Execute", mock.Anything, ".", "echo 42").Return("42", nil)
	parser.On("Parse", "42").Return(42.0, nil)
	reporter.On("UpdateBranch", "HEAD", "metric", true)

	cfg := &config.Config{
		BaseBranch: "main",
		Metric:     "echo 42",
		Operator:   config.OpGreaterThan,
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.Error(t, err)
	assert.True(t, errors.IsMetricTestError(err))
	git.AssertExpectations(t)
	executor.AssertExpectations(t)
	parser.AssertExpectations(t)
	reporter.AssertExpectations(t)
}

func TestOrchestrator_Run_NoComparison(t *testing.T) {
	// Setup mocks
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	// Configure expectations - no worktree operations for no-comparison mode
	git.On("IsGitRepository").Return(nil)

	executor.On("Execute", mock.Anything, ".", "echo 42").Return("42", nil)
	parser.On("Parse", "42").Return(42.0, nil)
	reporter.On("NoComparison", 42.0)

	cfg := &config.Config{
		Metric:   "echo 42",
		Operator: config.OpUnknown, // No comparison
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.NoError(t, err)
	git.AssertExpectations(t)
	executor.AssertExpectations(t)
	parser.AssertExpectations(t)
	reporter.AssertExpectations(t)
}

func TestOrchestrator_Run_WithPreAndPostCommands(t *testing.T) {
	// Setup mocks
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	// Configure expectations
	git.On("IsGitRepository").Return(nil)
	git.On("ResolveBranch", "main").Return("origin/main", nil)
	git.On("CreateWorktree", "/tmp/ratchet-base", "origin/main").Return(nil)
	git.On("RemoveWorktree", "/tmp/ratchet-base").Return(nil)

	reporter.On("Start", "main", "HEAD", []string{"pre", "metric", "post"}, true)
	reporter.On("Complete")

	// Base branch execution with pre/post
	reporter.On("UpdateBranch", "main", "pre", false)
	executor.On("Execute", mock.Anything, "/tmp/ratchet-base", "setup.sh").Return("", nil)
	reporter.On("UpdateBranch", "main", "pre", true)

	reporter.On("UpdateBranch", "main", "metric", false)
	executor.On("Execute", mock.Anything, "/tmp/ratchet-base", "echo 42").Return("42", nil)
	parser.On("Parse", "42").Return(42.0, nil)
	reporter.On("UpdateBranch", "main", "metric", true)

	reporter.On("UpdateBranch", "main", "post", false)
	executor.On("Execute", mock.Anything, "/tmp/ratchet-base", "cleanup.sh").Return("", nil)
	reporter.On("UpdateBranch", "main", "post", true)

	// HEAD branch execution with pre/post
	reporter.On("UpdateBranch", "HEAD", "pre", false)
	executor.On("Execute", mock.Anything, ".", "setup.sh").Return("", nil)
	reporter.On("UpdateBranch", "HEAD", "pre", true)

	reporter.On("UpdateBranch", "HEAD", "metric", false)
	executor.On("Execute", mock.Anything, ".", "echo 42").Return("50", nil)
	parser.On("Parse", "50").Return(50.0, nil)
	reporter.On("UpdateBranch", "HEAD", "metric", true)

	reporter.On("UpdateBranch", "HEAD", "post", false)
	executor.On("Execute", mock.Anything, ".", "cleanup.sh").Return("", nil)
	reporter.On("UpdateBranch", "HEAD", "post", true)

	// Success reporting
	reporter.On("Success", 50.0, 42.0, "greater than", "main")

	cfg := &config.Config{
		BaseBranch: "main",
		Pre:        "setup.sh",
		Metric:     "echo 42",
		Post:       "cleanup.sh",
		Operator:   config.OpGreaterThan,
		Verbose:    true,
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.NoError(t, err)
	git.AssertExpectations(t)
	executor.AssertExpectations(t)
	parser.AssertExpectations(t)
	reporter.AssertExpectations(t)
}

func TestOrchestrator_Run_GitRepositoryError(t *testing.T) {
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	git.On("IsGitRepository").Return(fmt.Errorf("not a git repository"))

	cfg := &config.Config{
		BaseBranch: "main",
		Metric:     "echo 42",
		Operator:   config.OpGreaterThan,
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.Error(t, err)
	assert.True(t, errors.IsGitError(err))
	git.AssertExpectations(t)
}

func TestOrchestrator_Run_BranchNotFound(t *testing.T) {
	git := &mockGitOperations{}
	executor := &mockCommandExecutor{}
	parser := &mockMetricParser{}
	reporter := &mockProgressReporter{}

	git.On("IsGitRepository").Return(nil)
	git.On("ResolveBranch", "nonexistent").Return("", fmt.Errorf("branch not found"))

	cfg := &config.Config{
		BaseBranch: "nonexistent",
		Metric:     "echo 42",
		Operator:   config.OpGreaterThan,
	}

	orchestrator := New(git, executor, parser, reporter, "/tmp")
	err := orchestrator.Run(context.Background(), cfg)

	assert.Error(t, err)
	assert.True(t, errors.IsGitError(err))
	git.AssertExpectations(t)
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
			got := FormatMetric(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}
