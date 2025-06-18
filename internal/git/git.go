package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// IsGitRepository checks if the current directory is a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetCurrentBranch returns the name of the current git branch
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// EnsureBranchExists checks if a branch exists locally, and fetches it if not
func EnsureBranchExists(branch string) error {
	// Check if branch exists locally
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// If running in GitHub Actions, use GITHUB_BASE_REF if available
	if githubRef := os.Getenv("GITHUB_BASE_REF"); githubRef != "" && branch == "main" {
		branch = githubRef
	}

	// Try to fetch the branch from origin
	remoteBranch := branch
	if !strings.HasPrefix(branch, "origin/") {
		remoteBranch = "origin/" + branch
	}

	// Fetch from remote
	cmd = exec.Command("git", "fetch", "origin", strings.TrimPrefix(remoteBranch, "origin/"))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch branch %s: %w\nOutput: %s", branch, err, output)
	}

	// Verify the remote branch exists
	cmd = exec.Command("git", "rev-parse", "--verify", remoteBranch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("branch %s not found locally or on remote", branch)
	}

	return nil
}

// CreateWorktree creates a temporary git worktree for the specified branch
func CreateWorktree(branch string) (string, func(), error) {
	// Determine temp directory
	tempDir := os.TempDir()
	if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
		tempDir = runnerTemp
	}

	// Create unique worktree directory with timestamp to avoid conflicts
	worktreeDir := filepath.Join(tempDir, fmt.Sprintf("ratchet-worktree-%d-%d", os.Getpid(), time.Now().UnixNano()))

	// Resolve branch reference
	branchRef := branch
	if !strings.HasPrefix(branch, "origin/") {
		// Check if local branch exists
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if err := cmd.Run(); err != nil {
			// Use remote branch
			branchRef = "origin/" + branch
		}
	}

	// Create worktree
	// First try without --force
	cmd := exec.Command("git", "worktree", "add", worktreeDir, branchRef)
	if output, err := cmd.CombinedOutput(); err != nil {
		// If it fails because branch is already checked out elsewhere, try with --detach
		if strings.Contains(string(output), "is already used by worktree") ||
			strings.Contains(string(output), "is already checked out") {
			// Use --detach to create a detached worktree at the same commit
			cmd = exec.Command("git", "worktree", "add", "--detach", worktreeDir, branchRef)
			if output2, err2 := cmd.CombinedOutput(); err2 != nil {
				return "", nil, fmt.Errorf("failed to create worktree: %w\nOutput: %s", err2, output2)
			}
		} else {
			return "", nil, fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, output)
		}
	}

	// Cleanup function
	cleanup := func() {
		// Remove worktree
		cmd := exec.Command("git", "worktree", "remove", worktreeDir, "--force")
		if err := cmd.Run(); err != nil {
			// Log but don't fail - we'll try manual removal
			fmt.Fprintf(os.Stderr, "Warning: failed to remove git worktree %s: %v\n", worktreeDir, err)
		}
		// Also try to remove directory in case git worktree remove failed
		if err := os.RemoveAll(worktreeDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove worktree directory %s: %v\n", worktreeDir, err)
		}
	}

	return worktreeDir, cleanup, nil
}
