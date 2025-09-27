# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **MCP Shell** - a Go CLI application generated from the `cookie-cutter-golang` template.

**Description**: MCP Shell Server for serving shell AI models

**Key Details**:
- **Binary Name**: `mcp-shell`
- **Author**: Rafael <<>>
- **GitHub**: https://github.com/rafa-dot-el/mcp-shell
- **License**: GPL3
- **Generated**: This project was created using cookiecutter template for professional Go CLI applications

## 🚨 CRITICAL: Boilerplate Cleanup Required

**This is a freshly generated project with boilerplate code that MUST be customized:**

### Immediate Actions Required
1. **Replace placeholder logic** in `internal/cmd/root.go` and other command files
2. **Update configuration schema** in `pkg/config/config.go` for your specific needs
3. **Implement actual functionality** - currently contains only "Hello World" examples
4. **Customize CLI commands** - replace example commands with real functionality
5. **Update documentation** - replace template docs with actual project information
6. **Configure environment variables** - update config for your specific use case

### Template Files to Customize
- `internal/cmd/*.go` - Replace example commands with actual CLI functionality
- `pkg/config/config.go` - Define your application's configuration schema
- `cmd/mcp-shell/main.go` - Main entry point (usually minimal changes needed)
- `README.md` - Update with actual project information and usage
- `docs/content/` - Replace template documentation with actual project docs
## Architecture & Design Decisions

### CLI Framework Architecture
**Technology Stack**:
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) for command structure and parsing
- **Configuration**: [Viper](https://github.com/spf13/viper) for hierarchical configuration management
- **Language**: Go (latest version available in nixpkgs 25.05)
- **License**: GPL3 (all dependencies are GPL3-compatible)

**Design Patterns**:
- **Clean Architecture**: Loosely coupled components with clear separation of concerns
- **Command Pattern**: Each CLI subcommand is a separate module in `internal/cmd/`
- **Configuration Hierarchy**: CLI flags > Environment vars > Config files > Defaults
- **Dependency Injection**: Services and dependencies injected through command constructors

### Directory Structure
```
mcp-shell/
├── cmd/mcp-shell/          # Main application entry point
│   ├── main.go                             # Bootstrap and dependency injection
│   └── e2e_test.go                        # End-to-end integration tests
├── internal/cmd/                          # CLI command implementations (private)
│   ├── root.go                            # Root command with global flags
│   ├── version.go                         # Version display command
├── pkg/config/                            # Configuration management (public API)
│   └── config.go                          # Viper-based config with validation
├── tests/                                 # Test fixtures and utilities
│   ├── integration/                       # Integration test suites
│   └── unit/                              # Unit test helpers
├── docs/                                  # Hugo documentation site
│   ├── hugo.toml                          # Hugo configuration
│   ├── content/                           # Documentation content
│   └── static/                            # Static assets
├── roadmap/                               # Project planning and bug tracking
│   ├── OPENBUGS.md                        # Active bug tracking
│   └── stage_*.md                         # Development stage planning
├── .github/workflows/                     # CI/CD pipelines
│   ├── ci.yml                             # Build, test, lint, security
│   └── release.yml                        # GoReleaser-based releases
├── Taskfile.yml                           # Development task automation
├── flake.nix                              # Nix development environment
├── .goreleaser.yml                        # Multi-platform release config
├── functions.sh                           # Development helper functions
└── Dockerfile                             # Multi-stage container build
```

### Configuration Management

**Hierarchical Configuration (highest to lowest priority)**:
1. **Command Line Flags**: `--log-level debug`
2. **Environment Variables**: `MCP-SHELL_LOG_LEVEL=debug`
3. **Configuration Files**: YAML format in multiple locations
4. **Default Values**: Hardcoded fallbacks

**Configuration File Search Order**:
1. `--config /path/to/config.yaml` (explicit path)
2. `~/.config/mcp-shell/config.yaml` (user config)
3. `./mcp-shell.yaml` (project-local config)
4. `/etc/mcp-shell/config.yaml` (system config)

**Environment Variable Convention**: `MCP-SHELL_SETTING_NAME`

Examples:
- `MCP-SHELL_LOG_LEVEL=debug`
- `MCP-SHELL_CONFIG_FILE=/custom/path/config.yaml`
- `MCP-SHELL_OUTPUT_FORMAT=json`

## Development Environment

### Nix Development Shell
**Automatic Environment**: This project uses direnv with `.envrc` for automatic environment activation.

```bash
# Enter development environment (automatic with direnv)
direnv allow

# Manual activation if needed
nix develop

# View available development tools
go version && task --version && hugo version
```

**Available Tools in Development Shell**:
- **Go Toolchain**: go, gofmt, goimports
- **Build Tools**: task (Taskfile runner), goreleaser
- **Quality Tools**: golangci-lint, gosec, govulncheck
- **Security**: trivy (vulnerability scanner)
- **Documentation**: hugo (static site generator)
### Development Helper Functions
**Loaded automatically via `functions.sh`**:

```bash
# Development watchers
build-watch           # Auto-rebuild on Go file changes
test-watch           # Auto-test on Go file changes
lint-watch           # Auto-lint on Go file changes

# Go-specific helpers
go-coverage          # Generate HTML coverage report
go-benchmark         # Run benchmark tests
go-race             # Race condition detection
go-profile          # CPU/memory profiling

# Security scanning
security-scan        # Run gosec + trivy
vuln-check          # Run govulncheck

# Quick operations
quick-check         # lint + unit tests
dev-setup          # Complete development setup
clean-all          # Clean all artifacts including git-ignored files
```

## Development Workflow

### Core Development Tasks
**Standard development commands via Taskfile**:

```bash
# Environment and dependencies
task                    # List all available tasks
task dev-setup         # Initial development setup

# Build and test
task build             # Build the CLI binary
task test              # Run all tests (unit + integration + e2e)
task test-unit         # Unit tests only
task test-integration  # Integration tests only
task test-e2e          # End-to-end tests only

# Code quality
task lint              # Code linting and formatting
task format            # Code formatting only
task vuln              # Vulnerability scanning
task security          # Security analysis (gosec + trivy)

# Documentation
task docs-build        # Build Hugo documentation site
task docs-serve        # Serve docs locally for development
task docs-test         # Test documentation (build + validate links)
# Complete pipeline
task all               # Build + test + lint + vuln + docs + validate
task clean             # Clean all build artifacts

# Version management
task version                 # Show current version
task version-bump-patch      # Bug fixes (0.1.0 -> 0.1.1)
task version-bump-minor      # New features (0.1.0 -> 0.2.0)
task version-bump-major      # Breaking changes (0.1.0 -> 1.0.0)
task version-tag             # Create git tag
task version-release         # Tag and trigger release
```

### Quality Gates
**A feature is ready when**:
- All tests pass (unit, integration, e2e)
- Code coverage > 90%
- Linting passes without errors
- Security scans show no critical vulnerabilities
- Documentation is updated
- All artifacts build successfully

**Pre-commit Requirements**:
- `task all` must pass completely
- No secrets or credentials in code
- All public functions documented
- GPL3 license headers on new files

## Testing Strategy

### Test Structure
```bash
# Unit tests - fast, isolated
go test ./internal/... ./pkg/...

# Integration tests - test component interaction
go test -tags=integration ./tests/integration/...

# End-to-end tests - full CLI testing
go test -tags=e2e ./cmd/mcp-shell/...

# Coverage reporting
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Testing Best Practices
- **100% Unit Test Coverage**: Every function must have comprehensive tests
- **Integration Tests**: Test configuration loading, command interaction
- **E2E Tests**: Test actual CLI binary with real inputs/outputs
- **Table-driven Tests**: Use test tables for multiple input scenarios
- **Test Fixtures**: Use `tests/` directory for test data and utilities

## Documentation

### Hugo Documentation Site
**Location**: `docs/` directory contains Hugo static site

**Local Development**:
```bash
task docs-serve    # Serve at http://localhost:1313
task docs-build    # Build static site to docs/public/
task docs-test     # Validate build and links
```

**Structure**:
- `docs/content/` - Markdown content files
- `docs/hugo.toml` - Hugo configuration
- `docs/static/` - Images, downloads, assets
- `docs/themes/` - Hugo theme (if custom)

**Deployment**: Automatic via GitHub Actions to GitHub Pages

### Documentation Requirements
**Must be documented**:
- Installation methods (Nix, containers, packages, binaries)
- Configuration options and file formats
- All CLI commands and flags
- Environment variables
- Uninstallation procedures
- Troubleshooting guides

## Release & Distribution

### Release Process
**Automated via GoReleaser and GitHub Actions**:

```bash
# Create release
task version-bump-minor     # or patch/major
task version-tag           # Create git tag
git push origin v1.2.3     # Triggers release pipeline
```

**Release Artifacts** (generated automatically):
- **Binaries**: Linux, macOS, Windows, FreeBSD (multiple architectures)
- **Container Images**: Tagged with semantic versions
- **Packages**: Homebrew (macOS/Linux), Scoop (Windows)
- **Archives**: tar.gz and zip files for all platforms
- **Checksums**: SHA256 checksums for all artifacts

### Distribution Channels
- **GitHub Releases**: Primary distribution with all artifacts
- **Container Registries**: Docker Hub, GitHub Container Registry
- **Package Managers**: Homebrew, Scoop
- **Direct Download**: Binary downloads from GitHub releases

### Version Management
**Semantic Versioning**: This project follows [SemVer](https://semver.org/)
- **MAJOR**: Breaking changes to CLI interface or configuration
- **MINOR**: New features, commands, or options (backward compatible)
- **PATCH**: Bug fixes and documentation updates

**Version Source**: `VERSION` file is the single source of truth
- Updated by `task version-bump-*` commands
- Propagated to all build artifacts and releases
- Git tags match version exactly (v1.2.3)

## Security & Compliance

### Security Standards
- **GPL3 License**: All code and dependencies must be GPL3-compatible
- **Vulnerability Scanning**: Automated via govulncheck and trivy
- **Dependency Auditing**: Regular security updates for Go modules
- **Secret Scanning**: No credentials or keys in source code
- **Static Analysis**: gosec for Go security best practices

### Security Commands
```bash
task vuln              # Run govulncheck
task security          # Full security scan (gosec + trivy)
go mod audit           # Check for known vulnerabilities in dependencies
gosec ./...           # Static security analysis
```

## Configuration Examples

### Sample Configuration File
**`~/.config/mcp-shell/config.yaml`**:
```yaml
# Application configuration
log_level: "info"        # debug, info, warn, error
output_format: "text"    # text, json, yaml

# Add your application-specific configuration here
# Replace this template configuration with your actual settings

# Example sections (customize for your CLI):
# database:
#   host: "localhost"
#   port: 5432
#   name: "myapp"
#
# api:
#   endpoint: "https://api.example.com"
#   timeout: "30s"
#   retries: 3
```

### Environment Variables
```bash
# Configuration
export MCP-SHELL_LOG_LEVEL="debug"
export MCP-SHELL_CONFIG_FILE="/path/to/config.yaml"
export MCP-SHELL_OUTPUT_FORMAT="json"

# Add your application-specific environment variables here
# export MCP-SHELL_DATABASE_URL="postgresql://..."
# export MCP-SHELL_API_KEY="your-api-key"
```

## Troubleshooting

### Common Issues
1. **Build Failures**: Ensure you're in the Nix development shell (`nix develop`)
2. **Test Failures**: Run `task clean` then `task build` to clear stale artifacts
3. **Linting Errors**: Run `task format` to auto-fix formatting issues
4. **Missing Dependencies**: Run `task dev-setup` to install/update dependencies

### Debug Commands
```bash
# Environment debugging
go env                 # Check Go environment
task --version        # Verify task runner
nix --version         # Verify Nix installation

# Build debugging
go build -x ./cmd/mcp-shell  # Verbose build output
go mod tidy           # Clean up module dependencies
go mod verify         # Verify module integrity

# Test debugging
go test -v ./...      # Verbose test output
go test -race ./...   # Race condition detection
go test -bench=. ./...  # Benchmark tests
```

### Getting Help
- **Documentation**: Generated Hugo site in `docs/`
- **CLI Help**: `mcp-shell --help` and `mcp-shell [command] --help`
- **Issues**: Report bugs and feature requests on GitHub

## Next Steps for Development

### Immediate Tasks (Post-Generation)
1. **Implement Core Logic**: Replace placeholder code in `internal/cmd/` with actual functionality
2. **Define Configuration Schema**: Update `pkg/config/config.go` for your specific needs
3. **Write Tests**: Create comprehensive tests for your specific functionality
4. **Update Documentation**: Replace template docs with actual project information
5. **Customize CLI Interface**: Define the commands and flags your tool actually needs

### Development Stages
1. **Stage 1**: Core functionality implementation
2. **Stage 2**: Configuration and CLI interface refinement
3. **Stage 3**: Testing and quality assurance
4. **Stage 4**: Documentation and user experience
5. **Stage 5**: Release preparation and distribution

**See `ROADMAP.md` and `roadmap/` directory for detailed development planning.**

This project provides enterprise-grade infrastructure - now customize it with your specific application logic and requirements.
- the project located at /home/rafael/src/codeberg/cookiecutter-golang
is the template that created this one, when you change code here to work, make sure you also change in the template