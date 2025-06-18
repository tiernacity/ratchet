# Ratchet CLI Specification

## Overview
Ratchet is a "software ratchet" tool that ensures metrics improve over time by
comparing command output between the current branch and a base branch. It helps
teams enforce continuous improvement by failing builds when quality metrics
don't improve.

## Core Functionality
- Accept a command (a "metric command") that outputs a single number (the "metric")
- If requested, compare the metric (a "metric test") between the working copy and the supplied branch
- Succeed if the metric test passes, fail otherwise
- Work both locally and in GitHub Actions
- Automatically handle git setup and branch fetching

README.md contains a description of ratchet behaviour and the accepted command
line options, config file options and github actions integration.

## Implementation Requirements

### Git Worktree Implementation
- Use git worktrees for branch comparison to avoid affecting current working directory
- Create temporary worktree in system temp directory
- Use `$RUNNER_TEMP` in GitHub Actions, fall back to `os.TempDir()` locally
- Clean up temporary worktrees after execution, even on errors
- Handle cases where base branch needs to be fetched from remote

### Branch Management
- Automatically and fetch base branch if not available locally
- In GitHub Actions, we expect users to use the `GITHUB_BASE_REF` environment variable or some other reference
- Support both local branch names (`main`) and remote references (`origin/main`)
- Provide clear error messages when branches cannot be found or fetched

### Command Execution
- Execute user-provided command in both base branch and current branch
- Parse numeric output from command stdout
- Commands that generate non-zero exit codes cause failure (error: 2)
- Validate that command output is a valid number (int or float)

### Cross-Platform Compatibility
- Work on Linux, macOS, and Windows
- Use Go's standard library for file operations and command execution
- Handle path separators and temp directories appropriately

### stdout/stderr
Ensure that:
- an error messages appear on stderr (not stdout)
- for a problem with the supplied command line arguments, display the usage/help text
- for **any other error**, do not display usage, just the error

In the examples below, there are checkboxes for each branch being tested. The checkboxes should become filled
like this [x] **as each step completes**. Ensure only one progress line is displayed for each branch.
- for passing metric tests, display something like the following example on stdout:
	```
	$ ratchet --gt origin/main './my-metric-test.sh'

	HEAD metric (5) is greater than origin/main (4)
	Succeeded
	```
- for failed metric tests, display something like the following example. Progress is reported on stdout, the status and failure on stderr:
	```
	$ ratchet --gt origin/main './my-metric-test.sh'
	origin/main: metric [ ]
	HEAD:        metric [ ]

	HEAD metric (5) is NOT greater than origin/main (6)
	Failed
	```
- when --verbose is given, and the ratchet fails, display something like the following on stdout and stderr:
	```
	$ ratchet --gt origin/main --pre './my-pre.sh' './my-metric-test.sh'
	origin/main: pre [ ] ; metric [ ]
	HEAD:        pre [ ] ; metric [ ]

	HEAD metric (5) is NOT greater than origin/main (6)
	Failed
	```
- when --verbose is given, and the ratchet succeeds, display something like the following on stdout:
	```
	$ ratchet --gt origin/main --pre './my-pre.sh' --post './my-post.sh' './my-metric-test.sh'
	origin/main: pre [ ] ; metric [ ] ; post [ ]
	HEAD:        pre [ ] ; metric [ ] ; post [ ]

	HEAD metric (5) is greater than origin/main (4)
	Succeeded
	```

- when and command fails, display something like the following on stdout and stderr:
	```
	$ ratchet --gt origin/main --pre './my-pre.sh' --post ''./my-post.sh './my-metric-test.sh'
	origin/main: pre [x] ; metric [x] ; post [x]
	HEAD:        pre [x] ; metric [ ] ; post [ ]

        Command './my-metric-test.sh' failed in HEAD
	Failed
	```

## Technical Implementation Details

### Directory Structure
```
/
├── main.go              # CLI entry point
├── cmd/                 # Cobra CLI commands
│   └── root.go
├── internal/
│   ├── git/             # Git operations (worktree, fetch, etc.)
│   ├── executor/        # Command execution
│   ├── parser/          # Output parsing and validation
│   └── ratchet/         # Core ratchet logic
├── action.yml           # GitHub Action definition
├── SPEC.md              # This file
└── README.md            # User documentation
```

### Core Algorithm
1. Parse and validate CLI arguments
2. Ensure base branch exists (fetch if necessary)
3. Create temporary worktree for base branch
4. Execute command in base branch worktree, capture output
5. Parse and validate numeric output from base branch
6. ALWAYS clean up base branch worktree
7. Execute command in current branch, capture output
8. Parse and validate numeric output from current branch
9. Compare values: succeed if metric test passes, fail otherwise
10. Exit with appropriate exit code
11. ONLY output on stdout if a) no metric test is supplied, in which case output the metric b) --verbose is supplied or c) the metric test failed or an error occurred

### Error Handling
- Git errors (branch not found, fetch failures)
- Command execution errors (command not found, permission issues)
- Command errors (valid commands that do not succeed)
- Output parsing errors (non-numeric output, empty output)
- File system errors (temp directory creation, cleanup)
- Provide helpful error messages with suggestions for resolution

### GitHub Actions Integration
- Detect GitHub Actions environment via environment variables
- Handle shallow clones by fetching necessary history
- Integrate with GitHub's workflow annotations for error reporting

### Dependencies
- Use minimal external dependencies
- Prefer Go standard library where possible
- Consider: cobra (CLI framework), viper, go-git (if needed for complex git operations)

## Testing Requirements
- Write only integrations tests, not unit tests
- Integration tests with real git repositories
- Tests for GitHub Actions environment simulation
- Tests for various command outputs and edge cases
- Cross-platform testing (GitHub Actions matrix)

## Documentation
- **DO NOT EDIT THE README** without asking first
