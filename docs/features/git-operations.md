# Git Operations Feature

## Overview
The Git Operations module provides an abstraction over Git commands needed by Ratchet. It handles repository validation, worktree management, and branch operations.

## Interface

```go
package git

type Operations interface {
    // IsGitRepository checks if the current directory is inside a Git repository
    IsGitRepository() error
    
    // CreateWorktree creates a new worktree at the specified path for the given branch
    CreateWorktree(path, branch string) error
    
    // RemoveWorktree removes the worktree at the specified path
    RemoveWorktree(path string) error
    
    // ResolveBranch resolves a branch reference to a commit SHA
    ResolveBranch(branch string) (string, error)
    
    // HasUncommittedChanges checks if there are uncommitted changes
    HasUncommittedChanges() (bool, error)
}
```

## Implementation Details

### Structure
```go
type gitImpl struct {
    // Could add configuration like git binary path if needed
}

func New() Operations {
    return &gitImpl{}
}
```

### IsGitRepository
Checks if the current directory is inside a Git repository.

```go
func (g *gitImpl) IsGitRepository() error {
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("not in a git repository")
    }
    return nil
}
```

### CreateWorktree
Creates a new Git worktree for the specified branch.

```go
func (g *gitImpl) CreateWorktree(path, branch string) error {
    // First, try to create the worktree
    cmd := exec.Command("git", "worktree", "add", path, branch)
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        // If it fails because branch is already checked out, force it
        if bytes.Contains(output, []byte("already checked out")) {
            cmd = exec.Command("git", "worktree", "add", "--force", path, branch)
            if output, err = cmd.CombinedOutput(); err != nil {
                return fmt.Errorf("failed to create worktree: %s", output)
            }
        } else {
            return fmt.Errorf("failed to create worktree: %s", output)
        }
    }
    
    return nil
}
```

### RemoveWorktree
Removes a Git worktree and cleans up any locks.

```go
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
```

### ResolveBranch
Resolves a branch name to its commit SHA.

```go
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
```

### HasUncommittedChanges
Checks for uncommitted changes in the repository.

```go
func (g *gitImpl) HasUncommittedChanges() (bool, error) {
    // Check for staged changes
    cmd := exec.Command("git", "diff", "--cached", "--quiet")
    if err := cmd.Run(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
            return true, nil // Has staged changes
        }
        return false, fmt.Errorf("failed to check staged changes: %w", err)
    }
    
    // Check for unstaged changes
    cmd = exec.Command("git", "diff", "--quiet")
    if err := cmd.Run(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
            return true, nil // Has unstaged changes
        }
        return false, fmt.Errorf("failed to check unstaged changes: %w", err)
    }
    
    return false, nil
}
```

## Error Handling

### Common Errors
1. **Not a Git Repository**: Current directory is not in a Git repo
2. **Branch Not Found**: Specified branch doesn't exist
3. **Worktree Already Exists**: Handle by forcing creation
4. **Worktree Locked**: Handle by pruning and retrying
5. **Permission Denied**: Insufficient permissions for Git operations

### Error Messages
Provide clear, actionable error messages:
- "not in a git repository" → "Please run ratchet from within a Git repository"
- "branch 'feature' not found" → "Branch 'feature' does not exist. Please check the branch name"
- "worktree locked" → "Git worktree is locked. Attempting to clean up..."

## Testing Strategy

### Unit Tests
Mock the exec.Command to test various scenarios:

```go
func TestGitOperations_CreateWorktree(t *testing.T) {
    tests := []struct {
        name       string
        path       string
        branch     string
        mockOutput func(cmd string, args []string) ([]byte, error)
        wantErr    bool
    }{
        {
            name:   "successful creation",
            path:   "/tmp/ratchet-123",
            branch: "main",
            mockOutput: func(cmd string, args []string) ([]byte, error) {
                return []byte(""), nil
            },
            wantErr: false,
        },
        {
            name:   "force creation when already checked out",
            path:   "/tmp/ratchet-123",
            branch: "main",
            mockOutput: func(cmd string, args []string) ([]byte, error) {
                if contains(args, "--force") {
                    return []byte(""), nil
                }
                return []byte("already checked out"), fmt.Errorf("exit 1")
            },
            wantErr: false,
        },
    }
}
```

### Integration Tests
Test with actual Git operations in a test repository:

```go
func TestGitOperations_Integration(t *testing.T) {
    // Create test repository
    testRepo := createTestRepo(t)
    defer os.RemoveAll(testRepo)
    
    // Change to test repo
    oldWd, _ := os.Getwd()
    os.Chdir(testRepo)
    defer os.Chdir(oldWd)
    
    git := New()
    
    // Test IsGitRepository
    assert.NoError(t, git.IsGitRepository())
    
    // Test CreateWorktree
    worktreePath := filepath.Join(t.TempDir(), "test-worktree")
    assert.NoError(t, git.CreateWorktree(worktreePath, "main"))
    
    // Test RemoveWorktree
    assert.NoError(t, git.RemoveWorktree(worktreePath))
}
```

## Platform Considerations

### Windows
- Use proper path separators
- Handle different Git installations (Git Bash, Git for Windows)
- Account for case-insensitive filesystems

### Unix (Linux/macOS)
- Handle symlinks properly
- Respect file permissions
- Use forward slashes for paths

## Performance Considerations

1. **Command Execution**: Git commands are generally fast for local operations
2. **Output Parsing**: Parse only what's needed, avoid large output
3. **Error Checking**: Use appropriate Git flags to minimize output
4. **Cleanup**: Always clean up worktrees to prevent accumulation

## Security Considerations

1. **Command Injection**: Never interpolate user input into Git commands
2. **Path Traversal**: Validate paths before passing to Git
3. **Permissions**: Ensure proper file permissions on created worktrees
4. **Sensitive Data**: Don't log potentially sensitive Git output

## Future Enhancements

1. **Shallow Clones**: Support shallow worktrees for large repositories
2. **Progress Reporting**: Report progress for long operations
3. **Batch Operations**: Combine multiple Git commands for efficiency
4. **Custom Git Binary**: Allow specifying alternative Git binary path