# MCP Shell

[![CI/CD Pipeline](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/ci.yml/badge.svg)](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/ci.yml)
[![Container Build](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/container.yml/badge.svg)](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/container.yml)
[![Documentation](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/docs.yml/badge.svg)](https://github.com/rafa-dot-el/mcp-shell/actions/workflows/docs.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rafa-dot-el/mcp-shell)](https://goreportcard.com/report/github.com/rafa-dot-el/mcp-shell)
[![Coverage Status](https://codecov.io/gh/rafa-dot-el/mcp-shell/branch/main/graph/badge.svg)](https://codecov.io/gh/rafa-dot-el/mcp-shell)

[![GitHub release](https://img.shields.io/github/release/rafa-dot-el/mcp-shell.svg)](https://github.com/rafa-dot-el/mcp-shell/releases)
[![GitHub tag](https://img.shields.io/github/tag/rafa-dot-el/mcp-shell.svg)](https://github.com/rafa-dot-el/mcp-shell/tags)
[![Docker Pulls](https://img.shields.io/docker/pulls/rafadotel/mcp-shell)](https://hub.docker.com/r/rafadotel/mcp-shell)
[![Docker Image Size](https://img.shields.io/docker/image-size/rafadotel/mcp-shell/latest)](https://hub.docker.com/r/rafadotel/mcp-shell)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Documentation](https://img.shields.io/badge/docs-github.io-blue)](https://rafa-dot-el.github.io/mcp-shell)

MCP Shell Server for serving shell AI models

## Installation

### Using Nix (Recommended)

```bash
# With flakes enabled
nix run github:rafa-dot-el/mcp-shell

# Or install to profile
nix profile install github:rafa-dot-el/mcp-shell
```

### GitHub Releases

Download the latest binary from the [releases page](https://github.com/rafa-dot-el/mcp-shell/releases):

```bash
# Linux/macOS
curl -L https://github.com/rafa-dot-el/mcp-shell/releases/latest/download/mcp-shell_linux_amd64.tar.gz | tar xz
sudo mv mcp-shell /usr/local/bin/

# Verify installation
mcp-shell version
```

### Container

```bash
# Run with Docker/Podman (Docker Hub)
docker run --rm rafadotel/mcp-shell:latest --help

# Run from GitHub Container Registry
docker run --rm ghcr.io/rafa-dot-el/mcp-shell:latest --help

# Pull specific version
docker pull rafadotel/mcp-shell:v0.1.0
docker pull ghcr.io/rafa-dot-el/mcp-shell:v0.1.0

# Run development version
docker run --rm rafadotel/mcp-shell:dev --help
```




### Build from Source

```bash
# Clone and build
git clone https://github.com/rafa-dot-el/mcp-shell.git
cd mcp-shell
go build -o mcp-shell ./cmd/mcp-shell
```

## Quick Start

```bash
# Basic usage
mcp-shell --help

# Show version
mcp-shell version

# Configuration help
mcp-shell --config /path/to/config.yaml

# Enable verbose output
mcp-shell --verbose
```

## Configuration

MCP Shell supports multiple configuration methods with the following precedence (highest to lowest):

1. **Command line flags** (highest priority)
2. **Environment variables** (`MCP-SHELL_*`)
3. **Configuration files**
4. **Default values** (lowest priority)

### Configuration Files

Configuration files are searched in the following order:

1. `--config` flag path (if specified)
2. `~/.config/mcp-shell/config.yaml`
3. `./mcp-shell.yaml` (current directory)
4. `/etc/mcp-shell/config.yaml`

Example configuration file:

```yaml
# ~/.config/mcp-shell/config.yaml
verbose: false
debug: false
log_level: info
```

### Environment Variables

All configuration options can be set via environment variables:

```bash
export MCP-SHELL_VERBOSE=true
export MCP-SHELL_DEBUG=false
export MCP-SHELL_LOG_LEVEL=debug
```

## Development

### Prerequisites

- [Nix](https://nixos.org/download.html) with flakes enabled
- [direnv](https://direnv.net/) (recommended)

### Setup

```bash
# Clone repository
git clone https://github.com/rafa-dot-el/mcp-shell.git
cd mcp-shell

# Enter development environment (with direnv)
direnv allow

# Or manually
nix develop

# Install dependencies and build
task deps
task build
```

### Available Tasks

Use `task` to see all available development tasks:

```bash
task                    # List all tasks
task build             # Build the binary
task test              # Run all tests
task lint              # Run linters
task vuln              # Security scans
task clean             # Clean artifacts
task dev-setup         # Setup development environment
```

### Development Helpers

The project includes helpful development functions in `functions.sh`:

```bash
# Available after entering nix develop
build-watch            # Watch for changes and rebuild
test-watch             # Watch for changes and retest
go-coverage            # Generate coverage report
security-scan          # Run security scanners
dev-help               # Show all available helpers
```

### Testing

```bash
# Run all tests
task test

# Run specific test types
task test-unit
task test-integration
task test-e2e

# Run with coverage
go-coverage
```

### Release Process

1. Update `VERSION` file
2. Create and push tag:
   ```bash
   task version-bump-minor  # or patch/major
   task version-tag
   git push origin v$(cat VERSION)
   ```
3. GitHub Actions will automatically build and publish the release

## Documentation


Full documentation is available at: https://rafa-dot-el.github.io/mcp-shell

### Building Documentation

```bash
task docs-build         # Build Hugo site
task docs-serve         # Serve locally
task docs-test          # Test documentation
```




## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes following the coding standards
4. Run tests: `task test`
5. Run linting: `task lint`
6. Commit changes: `git commit -am 'Add feature'`
7. Push to branch: `git push origin feature-name`
8. Submit a Pull Request

### Code Standards

- Follow Go conventions and best practices
- All functions must have 100% test coverage
- Use GPL3-compatible dependencies only
- Include proper GPL3 license headers
- Document all public functions and types

## Security

Security issues should be reported privately to <>.

For general vulnerabilities:
- Run `task vuln` to check for known vulnerabilities
- Security scans are run automatically in CI/CD

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📚 [Documentation](https://rafa-dot-el.github.io/mcp-shell)
- 🐛 [Issue Tracker](https://github.com/rafa-dot-el/mcp-shell/issues)
- 💬 [Discussions](https://github.com/rafa-dot-el/mcp-shell/discussions)

## Acknowledgments

Built with:
- [Go](https://golang.org/) - Programming language
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Nix](https://nixos.org/) - Development environment
- [GoReleaser](https://goreleaser.com/) - Release automation

---

Copyright (C) 2025 Rafael. Licensed under GPL-3.0.