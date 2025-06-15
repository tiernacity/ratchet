# Ratchet

[![Test](https://github.com/tiernacity/ratchet/actions/workflows/test.yml/badge.svg)](https://github.com/tiernacity/ratchet/actions/workflows/test.yml)
[![Release](https://github.com/tiernacity/ratchet/actions/workflows/release.yml/badge.svg)](https://github.com/tiernacity/ratchet/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tiernacity/ratchet)](https://goreportcard.com/report/github.com/tiernacity/ratchet)
[![License](https://img.shields.io/github/license/tiernacity/ratchet)](LICENSE)

Ratchet is a CLI tool that [describe what ratchet does - you'll need to fill this in based on your specific use case].

## Features

- Cross-platform support (Linux, macOS, Windows)
- Configuration file support (YAML)
- Available as both CLI and GitHub Action
- Verbose mode for debugging
- [Add your specific features here]

## Installation

### Using Go

```bash
go install github.com/tiernacity/ratchet/cmd/ratchet@latest
```

### Download Binary

Download the latest binary for your platform from the [releases page](https://github.com/tiernacity/ratchet/releases).

#### macOS/Linux

```bash
# Download the binary (replace OS and ARCH with your values)
curl -LO https://github.com/tiernacity/ratchet/releases/latest/download/ratchet-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')

# Make it executable
chmod +x ratchet-*

# Move to PATH (optional)
sudo mv ratchet-* /usr/local/bin/ratchet
```

#### Windows

Download `ratchet-windows-amd64.exe` from the [releases page](https://github.com/tiernacity/ratchet/releases).

### Using GitHub Action

```yaml
name: Run Ratchet
on: [push]

jobs:
  ratchet:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Ratchet
        uses: tiernacity/ratchet@v1
        with:
          config: .ratchet.yaml
          verbose: true
```

## Usage

### CLI

```bash
# Display help
ratchet --help

# Run with default settings
ratchet run

# Run with config file
ratchet run --config .ratchet.yaml

# Run in verbose mode
ratchet run --verbose

# Display version
ratchet version
```

### Configuration

Ratchet supports configuration via:
1. Command-line flags (highest priority)
2. Configuration file (YAML)
3. Environment variables

#### Configuration File

Create a `.ratchet.yaml` file in your project root or home directory:

```yaml
# See config/.ratchet.example.yaml for full example
verbose: false

# Add your configuration here
```

Copy `config/.ratchet.example.yaml` for a complete example with all available options.

#### Environment Variables

All configuration options can be set via environment variables with the `RATCHET_` prefix:

```bash
export RATCHET_VERBOSE=true
ratchet run
```

### GitHub Action

The Ratchet GitHub Action supports the following inputs:

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `version` | Version of ratchet to use | No | `latest` |
| `config` | Path to configuration file | No | - |
| `verbose` | Enable verbose output | No | `false` |
| `command` | Ratchet command to run | No | `run` |
| `args` | Additional arguments | No | - |

#### Example Workflows

**Basic usage:**
```yaml
- uses: tiernacity/ratchet@v1
```

**With configuration:**
```yaml
- uses: tiernacity/ratchet@v1
  with:
    config: .github/ratchet.yaml
    verbose: true
```

**Custom command:**
```yaml
- uses: tiernacity/ratchet@v1
  with:
    command: custom-command
    args: --flag value
```

## Development

### Prerequisites

- Go 1.22 or higher
- Make (optional)

### Building

```bash
# Build for current platform
go build -o bin/ratchet ./cmd/ratchet

# Build for all platforms
make build-all

# Install locally
go install ./cmd/ratchet
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.txt ./...

# Run linter (requires golangci-lint)
golangci-lint run
```

### Project Structure

```
ratchet/
├── cmd/
│   └── ratchet/      # CLI application entry point
├── config/           # Configuration examples
├── .github/
│   └── workflows/    # GitHub Actions workflows
├── action.yml        # GitHub Action definition
├── go.mod           # Go module definition
└── README.md        # This file
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass
- Code is formatted (`go fmt ./...`)
- Linting passes (`golangci-lint run`)

### Release Process

Releases are automated via GitHub Actions:

1. Update version in code if needed
2. Commit changes
3. Create and push a tag: `git tag v1.0.0 && git push --tags`
4. GitHub Actions will build and create a release

## License

[Specify your license here]

## Acknowledgments

[Add any acknowledgments here]