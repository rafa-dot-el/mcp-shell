---
title: "Development Guide"
linkTitle: "Development"
weight: 40
description: >
  Set up your development environment and contribute to MCP Shell
---

# Development Guide

This guide will help you set up a development environment for MCP Shell and understand the project structure, workflow, and contribution process.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Nix** (with flakes enabled) - Recommended for reproducible development environment
- **direnv** - Automatic environment loading
- **Git** - Version control
- **GPG** (optional) - For signing commits

### Installing Prerequisites

**Nix with Flakes:**
```bash
# Install Nix (multi-user installation)
sh <(curl -L https://nixos.org/nix/install) --daemon

# Enable flakes (add to ~/.config/nix/nix.conf or /etc/nix/nix.conf)
experimental-features = nix-command flakes
```

**direnv:**
```bash
# Most distributions
nix profile install nixpkgs#direnv

# Or via package manager
# Debian/Ubuntu: apt install direnv
# macOS: brew install direnv

# Add to your shell rc file (~/.bashrc, ~/.zshrc, etc.)
eval "$(direnv hook bash)"  # or zsh, fish, etc.
```

## Setting Up Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/rafa-dot-el/mcp-shell.git
cd mcp-shell
```

### 2. Enter Development Shell

The project uses Nix flakes for reproducible development environments:

```bash
# Allow direnv (automatic environment loading)
direnv allow

# Or manually enter the development shell
nix develop
```

The development shell provides:
- Go toolchain (1.23+)
- Task runner (go-task)
- Hugo (documentation site generator)
- golangci-lint (linting)
- gosec (security analysis)
- govulncheck (vulnerability scanner)
- trivy (container security)
- Development helper functions

### 3. Install Dependencies

```bash
# Download Go modules
task dev-setup
```

### 4. Verify Setup

```bash
# Check available tools
go version
task --version
hugo version

# Build the project
task build

# Run tests
task test
```

## Development Workflow

### Taskfile Commands

The project uses Taskfile for task automation. View all available tasks:

```bash
task --list
```

**Common Tasks:**

```bash
# Build
task build              # Build CLI binary
task build-all          # Build all platforms

# Testing
task test               # Run all tests
task test-unit          # Unit tests only
task test-integration   # Integration tests
task test-e2e           # End-to-end tests
task test-coverage      # Generate coverage report

# Code Quality
task lint               # Run linters
task format             # Format code
task vuln               # Vulnerability scan
task security           # Security analysis

# Documentation
task docs-build         # Build Hugo site
task docs-serve         # Serve docs locally (http://localhost:1313)
task docs-test          # Test documentation

# Development
task clean              # Clean build artifacts
task all                # Build + test + lint + vuln + docs

# Version Management
task version                 # Show current version
task version-bump-patch      # Bump patch version (0.1.0 -> 0.1.1)
task version-bump-minor      # Bump minor version (0.1.0 -> 0.2.0)
task version-bump-major      # Bump major version (0.1.0 -> 1.0.0)
```

### Helper Functions

The development shell automatically loads helper functions from `functions.sh`:

```bash
# Development watchers
build-watch         # Auto-rebuild on file changes
test-watch          # Auto-test on file changes
lint-watch          # Auto-lint on file changes

# Go utilities
go-coverage         # Generate HTML coverage report
go-benchmark        # Run benchmark tests
go-race             # Race condition detection
go-profile          # CPU/memory profiling

# Security
security-scan       # Run gosec + trivy
vuln-check          # Run govulncheck

# Quick checks
quick-check         # lint + unit tests
dev-setup           # Complete environment setup
clean-all           # Clean all artifacts
```

## Project Structure

```
mcp-shell/
├── cmd/mcp-shell/          # Main application entry point
│   ├── main.go            # Bootstrap and dependency injection
│   └── e2e_test.go        # End-to-end tests
├── internal/cmd/          # CLI command implementations (private)
│   ├── root.go            # Root command with global flags
│   ├── version.go         # Version command
│   ├── serve.go           # MCP server command
│   ├── validate.go        # Configuration validation
│   ├── list.go            # List scripts/aliases
│   └── *_test.go          # Unit tests for commands
├── pkg/                   # Public packages
│   ├── config/            # Configuration management
│   ├── mcp/               # MCP protocol implementation
│   ├── script/            # Script management
│   └── job/               # Job execution and scheduling
├── tests/                 # Test utilities
│   ├── integration/       # Integration test suites
│   └── unit/              # Unit test helpers
├── docs/                  # Hugo documentation site
├── examples/              # Example configurations
├── .github/workflows/     # CI/CD pipelines
├── Taskfile.yml           # Task automation
├── flake.nix              # Nix development environment
├── go.mod                 # Go module dependencies
└── Dockerfile             # Container image

```

### Package Organization

**internal/cmd/** - CLI commands (Cobra)
- Each command is a separate file
- Private to this project
- Handles user interaction and orchestration

**pkg/config/** - Configuration management (Viper)
- Configuration loading and validation
- Hierarchical configuration resolution
- Public API for configuration access

**pkg/mcp/** - MCP protocol implementation
- Model Context Protocol server
- Stdio transport
- Tool, prompt, and resource handlers

**pkg/script/** - Script management
- Script discovery and loading
- Parameter validation
- Execution coordination

**pkg/job/** - Job execution
- Job queue management
- Parallel execution control
- Logging and status tracking
- Cron-based scheduling

## Architecture

### Design Principles

**Clean Architecture:**
- Loosely coupled components
- Clear separation of concerns
- Dependency injection
- Testable code

**Configuration Hierarchy:**
```
CLI Flags > Environment Variables > Config Files > Defaults
```

**Command Pattern:**
- Each CLI subcommand is a separate module
- Commands orchestrate package functionality
- Minimal logic in command handlers

### Key Components

**Configuration Management (Viper):**
- Hierarchical configuration loading
- Multiple config file locations
- Environment variable support
- Configuration validation

**Script Manager:**
- Discovers scripts from folders
- Loads explicit script configurations
- Validates script files and parameters
- Manages aliases

**Job Manager:**
- Executes scripts/aliases
- Manages job queue
- Controls parallelism
- Tracks job status and logs

**Scheduler:**
- Cron-based scheduling
- One-time scheduled execution
- Job lifecycle management

**MCP Server:**
- Stdio transport
- Tool execution (scripts/aliases)
- Prompt templates
- Resource access (logs)

## Testing Strategy

### Test Coverage Requirements

- **Overall**: >90% coverage
- **Pure Functions**: 100% coverage goal
- **Integration Points**: Comprehensive integration tests
- **E2E**: Critical user workflows

### Running Tests

```bash
# All tests
task test

# Unit tests only
task test-unit

# Integration tests
task test-integration

# E2E tests
task test-e2e

# With coverage
task test-coverage
go tool cover -html=coverage.out
```

### Writing Tests

**Unit Tests:**
```go
func TestScriptValidation(t *testing.T) {
    // Arrange
    cfg := &config.Config{...}

    // Act
    err := cfg.Validate()

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got: %v", err)
    }
}
```

**Table-Driven Tests:**
```go
func TestParameterValidation(t *testing.T) {
    tests := []struct {
        name    string
        param   Parameter
        value   string
        wantErr bool
    }{
        {"valid required", Parameter{Required: true}, "value", false},
        {"missing required", Parameter{Required: true}, "", true},
        {"valid optional", Parameter{Required: false}, "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.param.Validate(tt.value)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr = %v", err, tt.wantErr)
            }
        })
    }
}
```

**Integration Tests:**
```go
func TestServerIntegration(t *testing.T) {
    // Setup
    server, cfg, tmpDir := setupTestServer(t)
    defer cleanup(server, tmpDir)

    // Test
    info := server.GetServerInfo()
    if info.Name != cfg.MCP.Name {
        t.Errorf("Expected name %s, got %s", cfg.MCP.Name, info.Name)
    }
}
```

## Code Quality

### Linting

The project uses golangci-lint with comprehensive checks:

```bash
# Run all linters
task lint

# Auto-fix issues
task format
```

Configured linters:
- gofmt, goimports (formatting)
- govet (correctness)
- errcheck (error handling)
- staticcheck (bugs and performance)
- gosec (security)
- revive (style)

### Security Scanning

```bash
# Vulnerability scanning
task vuln

# Security analysis
task security

# Container scanning
trivy image mcp-shell:latest
```

### Pre-commit Hooks

The project uses pre-commit hooks for quality gates:

```bash
# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

Hooks run:
- Format checking
- Linting
- Tests
- Security scans

## Contributing

### Git Workflow

**Branch Strategy:**
- `main` - Stable releases
- `dev` - Development integration
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches
- `chore/*` - Maintenance branches

**Workflow:**
```bash
# Create feature branch
git checkout -b feature/my-feature dev

# Make changes, commit often
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/my-feature

# Create pull request to dev branch
# After review and merge, PR from dev to main for release
```

### Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding/updating tests
- `refactor`: Code restructuring
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples:**
```
feat(cli): add validate command for config validation

Implemented validation command that checks:
- Configuration syntax
- Script file permissions
- Parameter definitions

Closes #123
```

```
fix(job): resolve race condition in parallel execution

Added mutex protection for concurrent job status updates
```

### Pull Request Process

1. **Create Feature Branch** from `dev`
2. **Implement Changes** with tests
3. **Run Quality Checks**:
   ```bash
   task all  # Must pass
   ```
4. **Commit Changes** with conventional commits
5. **Push Branch** to remote
6. **Create Pull Request** to `dev` branch
7. **Address Review Feedback**
8. **Squash Commits** before merge
9. **Merge to Dev** after approval
10. **Release to Main** via PR from dev

### Quality Gates

Before submitting a PR, ensure:

- [ ] All tests pass (`task test`)
- [ ] Code coverage >90% (`task test-coverage`)
- [ ] Linting passes (`task lint`)
- [ ] Security scans pass (`task security`)
- [ ] Documentation updated
- [ ] Example configurations updated if needed
- [ ] CHANGELOG.md updated (for releases)

## Building and Releasing

### Local Builds

```bash
# Build for current platform
task build

# Build for all platforms
task build-all

# Test build
./bin/mcp-shell version
```

### Container Images

```bash
# Build container image
podman build -t mcp-shell:local .

# Test container
podman run --rm mcp-shell:local version
```

### Release Process

Releases are automated via GoReleaser:

1. **Update Version**: `task version-bump-minor`
2. **Commit Version**: `git commit -am "chore: bump version to X.Y.Z"`
3. **Create Tag**: `git tag vX.Y.Z`
4. **Push Tag**: `git push origin vX.Y.Z`
5. **GitHub Actions** builds and publishes:
   - Binaries for all platforms
   - Container images (Docker Hub, GHCR)
   - GitHub Release with artifacts
   - Documentation site

## Documentation

### Building Documentation

```bash
# Serve locally
task docs-serve

# Visit http://localhost:1313

# Build static site
task docs-build
```

### Writing Documentation

Documentation uses Hugo with Docsy theme:

```
docs/content/en/
├── _index.html           # Homepage
├── docs/
│   ├── _index.md        # Docs landing
│   ├── installation/    # Installation guide
│   ├── getting-started/ # Getting started
│   ├── configuration/   # Configuration
│   └── development/     # This guide
└── tutorials/           # Tutorials
```

**Front Matter:**
```yaml
---
title: "Page Title"
linkTitle: "Short Title"
weight: 10
description: >
  Brief description
---
```

## Troubleshooting

### Common Issues

**Nix shell not loading:**
```bash
# Ensure flakes are enabled
cat ~/.config/nix/nix.conf

# Reload direnv
direnv reload
```

**Tests failing:**
```bash
# Clean and rebuild
task clean
task build
task test
```

**Linting errors:**
```bash
# Auto-fix formatting
task format

# Review remaining issues
task lint
```

### Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/rafa-dot-el/mcp-shell/issues)
- **GitHub Discussions**: [Community discussions](https://github.com/rafa-dot-el/mcp-shell/discussions)
- **Documentation**: [Full documentation site](https://rafa-dot-el.github.io/mcp-shell)

## License

MCP Shell is licensed under GPL-3.0. All contributions must be compatible with this license.

When contributing:
- Ensure your code is original or properly attributed
- Do not include proprietary code
- All dependencies must be GPL-3.0 compatible
- Include appropriate license headers in new files

## Next Steps

- [Configuration](/docs/configuration/): Learn about configuration options
- [Getting Started](/docs/getting-started/): Basic usage guide
- [GitHub Repository](https://github.com/rafa-dot-el/mcp-shell): Source code and issues
