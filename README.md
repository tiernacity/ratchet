# Ratchet

A [software ratchet](https://qntm.org/ratchet) helps teams make measurable,
long-term improvements to their codebase. Want to reduce `TODO` comments or
increase test coverage? If you can measure it, a ratchet can help you enforce
it.

`ratchet` is a CLI tool and GitHub Action that ensures your metric moves in the right direction. It works by:
1. Running your metric command (any shell command that outputs a number)
2. Comparing the result against a base branch (e.g., `main`)
3. Failing if the metric hasn't improved

Good for use in your CI pipeline to drive incremental improvement.

## Quick Start

### Installation

#### Download Binary
Download the latest release from [GitHub Releases](https://github.com/tiernacity/ratchet/releases).

#### Install with Go
```bash
go install github.com/tiernacity/ratchet/cmd/ratchet@latest
```

### Basic Usage

```bash
# Report a count of TODO comments 
ratchet "grep -r TODO . | wc -l"

# TODO must reduce compared to `main`
ratchet --lt main "grep -r TODO . | wc -l"

# TODO must not increase compared to `main`
ratchet --le main "grep -r TODO . | wc -l"

# Perform setup before running test command
ratchet --pre "npm install" "grep -r TODO . | wc -l"

# Perform setup and teardown
ratchet --pre "./setup.sh" --post "./teardown.sh" "grep -r TODO . | wc -l"
```

Ratchet runs your command in two contexts:
1. **Base branch** (usually `main`) - establishes the baseline metric
2. **Current branch** - your changes being tested

ratchet compares the two values using the test you specified. If the test
passes, ratchet succeeds. If not, it returns an error code.

The "base" for comparison can be any git commit-ish object. The other test
is run using the current git working copy. Only one test must be specified:
- --gt: working copy metric is **greater than** the supplied branch
- --ge: metric is **greater than or equal**
- --eq: metrics are **equal**
- --le: metric is **less than or equal**
- --lt: metric is **less than**

## Some Use Cases

```bash
# Reduce linting errors
ratchet --lt main "eslint . --format=compact | wc -l"

# Do not increase complexity violations
ratchet --le develop "gocyclo -over 10 . | wc -l"

# Increase coverage 
ratchet --gt trunk --pre "./run-tests.sh --coverage=true" "cat ./coverage.txt"
```

## Use a Config File
```bash
$ cat .ratchet
metric: npm test | grep skip | wc -l
pre: npm install ; npm run db-setup
post: npm run db-teardown
lt: origin/main

# Defaults to using ./.ratchet, if present
$ ratchet

# or specify an explicit config file path
$ ratchet --config-file ./.ratchet

# or... just supply the YAML string!
$ ratchet --config "$(cat ./.ratchet)"
```

## GitHub Actions Integration

Use any of the options allowed in the config file
```yaml
name: Quality Ratchet
on: [pull_request]

jobs:
  ratchet:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Needed for branch comparison
          
      - uses: tiernacity/ratchet@v1
        with:
          metric: "grep -r TODO . | wc -l"
          lt: ${{ github.base_ref }}
```

## CLI Options

```
Usage:
  ratchet [flags] <metric command>

Comparison operators (choose one):
      --less-than, --lt <base>       test that HEAD metric < base branch metric
      --less-equal, --le <base>      test that HEAD metric <= base branch metric
      --equal-to, --eq <base>        test that HEAD metric == base branch metric
      --greater-equal, --ge <base>   test that HEAD metric >= base branch metric
      --greater-than, --gt <base>    test that HEAD metric > base branch metric

Other flags:
  -h, --help                   help for ratchet
      --pre <command>          Command to run before metric command
      --post <command>         Command to run after metric command
      --config-file string     Path to config file (YAML or JSON)
      --config string          Config string (YAML or JSON)
  -v, --verbose                Show detailed output including both values
      --version                Show version information
```

## Exit Codes

- `0`: Success - metric test succeeded
- `1`: Failure - metric test failed
- `2`: Error - invalid usage, git errors, or command failures

## Troubleshooting

### "Base branch not found"
```bash
# Fetch the base branch manually
git fetch origin main

# Or specify a remote branch
ratchet --lt origin/main "your-command"
```

### "Invalid metric"
Ensure your command outputs only a number:
```bash
# ❌ Bad - includes text
eslint . 

# ✅ Good - only number
eslint . --format=compact | wc -l

# ✅ Good - extract number from output
npm audit --parseable | wc -l
```

### GitHub Actions: Shallow Clone Issues
Add fetch-depth to your checkout:
```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0  # or fetch-depth: 2 for faster execution
```

### Command Fails But Has Valid Output
Some commands might fail but still produce countable output:
```bash
# This might exit with code 1 but still count errors
ratchet "eslint . 2>/dev/null | wc -l"
```

## Contributing

I welcome contributions! There is not yet a [Contributing Guide](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE.md) file for details.
