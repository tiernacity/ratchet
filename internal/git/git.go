package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// gitImpl implements the Operations interface using the git command-line tool
type gitImpl struct{}

// New creates a new Git operations implementation
func New() Operations {
	return &gitImpl{}
}

// IsGitRepository checks if the current directory is inside a Git repository
func (g *gitImpl) IsGitRepository() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not in a git repository")
	}
	return nil
}

// CreateWorktree creates a new Git worktree for the specified branch
func (g *gitImpl) CreateWorktree(path, branch string) error {
	// Always use --force to handle existing worktrees/directories robustly
	cmd := exec.Command("git", "worktree", "add", "--force", path, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s", output)
	}
	return nil
}

// RemoveWorktree removes a Git worktree and cleans up any locks
func (g *gitImpl) RemoveWorktree(path string) error {
	// First try normal removal
	cmd := exec.Command("git", "worktree", "remove", path)
	if err := cmd.Run(); err != nil {
		// If it fails, try with --force
		cmd = exec.Command("git", "worktree", "remove", "--force", path)
		if err := cmd.Run(); err != nil {
			// Last resort: prune worktrees
			exec.Command("git", "worktree", "prune").Run()
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
	}
	return nil
}

// ResolveBranch resolves a branch name to its commit SHA
func (g *gitImpl) ResolveBranch(branch string) (string, error) {
	// Try to resolve as a branch first
	cmd := exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("refs/heads/%s", branch))
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Try as a remote branch
	cmd = exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("refs/remotes/origin/%s", branch))
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Try as any valid ref
	cmd = exec.Command("git", "rev-parse", "--verify", branch)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve branch '%s'", branch)
	}

	return strings.TrimSpace(string(output)), nil
}
