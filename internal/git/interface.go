package git

// Operations defines the interface for Git operations required by Ratchet
type Operations interface {
	// IsGitRepository checks if the current directory is inside a Git repository
	IsGitRepository() error

	// CreateWorktree creates a new worktree at the specified path for the given branch
	CreateWorktree(path, branch string) error

	// RemoveWorktree removes the worktree at the specified path
	RemoveWorktree(path string) error

	// ResolveBranch resolves a branch reference to a commit SHA
	ResolveBranch(branch string) (string, error)
}
