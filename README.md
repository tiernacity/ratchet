# Ratchet

A [software ratchet](https://qntm.org/ratchet) helps teams make measurable,
long-term improvements to their codebase. If you can define and measure a metric,
a ratchet can help you enforce beneficial changes to it. Perhaps you want to
reduce the incidence of 'TODO' comments in your codebase, for instance.

`ratchet` is a CLI tool and GitHub Action that ensures your metric moves in the right direction. It works by:

1. Running a shell command that outputs a number - your metric
2. Comparing the metric against a base branch (for example, `main`)
3. Failing if the metric hasn't improved

You can deploy it to your CI pipeline to drive incremental improvement.

## A note

This repository is two things:

1. A hopefully working and usable ratchet tool for you to use
2. An experiment in using Claude Code to write everything for me. As such, a learning exercise (and quite the learning exercise it has been!)

Because of the second item, you're going to find some things that could or should be improved. Off the top of my head:

- design documentation does not always reflect the implementation
- repetition and redundancy in the documentation and, most likely, in the codebase
- implementation/patterns/idioms are not going to be optimal in some cases

Over time, I'll be figuring out how to work with the agent to improve such things

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
# Report a count of TODO comments to stdout
ratchet "grep -r TODO . | wc -l"

# TODO must reduce compared to `main` - fail if it has not
ratchet --lt main "grep -r TODO . | wc -l"

# TODO must not increase compared to `main`
ratchet --le main "grep -r TODO . | wc -l"

# Perform setup before running the metric command
ratchet --pre "npm install" "grep -r TODO . | wc -l"

# Perform setup and teardown
ratchet --pre "./setup.sh" --post "./teardown.sh" "grep -r TODO . | wc -l"
```

Ratchet runs your command in two contexts:

1. **Base branch** (perhaps `main`, or the base branch of your PR) - to establishe the baseline metric
2. **Current branch** (actually, the current working copy) - to compare your branch against the base

ratchet compares the two values using the criterion you specify. If the test
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

# override options
$ ratchet --le origin/staging

# specify an explicit config file path
$ ratchet --config-file ./.ratchet

# or... just supply the YAML string!
$ ratchet --config "$(cat ./.ratchet)"
```

## Use a github workflow

Use any of the options allowed in the config file, in your github workflow

```yaml
name: Quality Ratchet
on: [pull_request]

jobs:
  ratchet:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Needed to access the base branch

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
```

### GitHub Actions: Shallow Clone Issues

Add fetch-depth to your checkout:

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0 # or fetch-depth: 2 for faster execution
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
