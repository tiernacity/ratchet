package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CleanupOrphanedWorktrees removes any orphaned ratchet worktrees from temp directories
func CleanupOrphanedWorktrees() error {
	// Get temp directory
	tempDir := os.TempDir()
	if runnerTemp := os.Getenv("RUNNER_TEMP"); runnerTemp != "" {
		tempDir = runnerTemp
	}

	// Find ratchet worktree directories
	pattern := filepath.Join(tempDir, "ratchet-worktree-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to search for orphaned worktrees: %w", err)
	}

	var cleaned []string
	var errors []string

	for _, worktreePath := range matches {
		// Try to remove with git first
		cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
		if err := cmd.Run(); err != nil {
			// Git removal failed, try manual removal
			if err := os.RemoveAll(worktreePath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove %s: %v", worktreePath, err))
				continue
			}
		}
		cleaned = append(cleaned, worktreePath)
	}

	if len(cleaned) > 0 {
		fmt.Printf("Cleaned up %d orphaned worktrees:\n", len(cleaned))
		for _, path := range cleaned {
			fmt.Printf("  - %s\n", path)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
