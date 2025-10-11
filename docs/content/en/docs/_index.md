---
title: "Documentation"
linkTitle: "Docs"
weight: 20
---

# MCP Shell Documentation

Welcome to the MCP Shell documentation! This guide will help you get started with MCP Shell, a cross-platform CLI tool that provides a Model Context Protocol (MCP) server for serving shell AI models.

## Getting Started

New to MCP Shell? Start here:

* [Installation](/docs/installation/): How to install MCP Shell on your system
* [Getting Started](/docs/getting-started/): Basic usage and first steps
* [Configuration](/docs/configuration/): Configure MCP Shell for your needs

## Development

Want to contribute or extend MCP Shell?

* [Development Guide](/docs/development/): Set up development environment
* [Architecture](/docs/development/architecture/): Understanding MCP Shell's design
* [Contributing](/docs/development/contributing/): How to contribute to the project

## Key Features

### Core Functionality
- **Script Management**: Execute shell scripts with parameter validation and type checking
- **Alias System**: Define simple command aliases for common operations
- **Configuration Validation**: Built-in `validate` command checks configuration correctness
- **Discovery & Listing**: `list` command shows all available scripts and aliases
- **Multiple Output Formats**: Support for table, JSON, and simple list formats

### Execution & Management
- **Parallel Execution**: Control concurrent job execution with configurable limits
- **Job Scheduling**: Cron-based scheduling and one-time scheduled execution
- **Job Logging**: Comprehensive logging with log tailing and search capabilities
- **Timeout Management**: Configurable execution timeouts per script
- **Background Jobs**: Optional background execution support

### Configuration & Security
- **Hierarchical Configuration**: CLI flags > Environment vars > Config files > Defaults
- **Security Controls**: Interpreter whitelist, script creation controls
- **Flexible Configuration**: YAML-based configuration with validation
- **Script Discovery**: Auto-discover scripts from folders using glob patterns

### Platform & Distribution
- **Cross-platform**: Works on Linux, macOS, Windows, and FreeBSD
- **Container Ready**: Small, secure container images (scratch-based)
- **Multiple Install Methods**: Nix, binary downloads, containers, source builds
- **Comprehensive Testing**: >90% code coverage with unit, integration, and E2E tests

### Development & Quality
- **Security Scanning**: Built-in vulnerability and security analysis
- **Quality Gates**: Automated linting, testing, and security checks
- **Documentation**: Comprehensive documentation with examples
- **Open Source**: Licensed under GPL-3.0 with full source code available

## Need Help?

- [GitHub Issues](https://github.com/rafa-dot-el/mcp-shell/issues): Report bugs or request features
- [GitHub Discussions](https://github.com/rafa-dot-el/mcp-shell/discussions): Community discussions
- [GitHub Repository](https://github.com/rafa-dot-el/mcp-shell): Source code and releases