package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitImpl_IsGitRepository(t *testing.T) {
	git := New()

	// This test assumes we're running in a git repository
	// In a real test environment, we might want to create a temporary git repo
	err := git.IsGitRepository()

	// If we're in a git repo, this should pass
	// If not, we'll get an error which is also valid behavior
	if err != nil {
		assert.Contains(t, err.Error(), "not in a git repository")
	}
}

func TestGitImpl_ResolveBranch_InvalidBranch(t *testing.T) {
	git := New()

	// Test with a branch that definitely doesn't exist
	_, err := git.ResolveBranch("this-branch-definitely-does-not-exist-12345")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve branch")
}

func TestGitImpl_CreateWorktree_InvalidPath(t *testing.T) {
	git := New()

	// Test with invalid path and invalid branch
	err := git.CreateWorktree("/this/path/does/not/exist", "invalid-branch")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create worktree")
}

func TestGitImpl_RemoveWorktree_NonexistentPath(t *testing.T) {
	git := New()

	// Test removing a worktree that doesn't exist
	err := git.RemoveWorktree("/this/path/does/not/exist")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree")
}
