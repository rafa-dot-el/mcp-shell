---
title: "Installation"
linkTitle: "Installation"
weight: 10
description: >
  How to install MCP Shell on your system
---

# Installation

MCP Shell provides multiple installation methods to suit different environments and preferences.

## Nix (Recommended)

The easiest way to install MCP Shell is through Nix:

### Run Directly
```bash
nix run github:rafa-dot-el/mcp-shell
```

### Install to Profile
```bash
nix profile install github:rafa-dot-el/mcp-shell
```

### Development Environment
```bash
nix develop github:rafa-dot-el/mcp-shell
```

## Binary Downloads

Download pre-built binaries from our GitHub releases:

### Linux
```bash
curl -L https://github.com/rafa-dot-el/mcp-shell/releases/latest/download/mcp-shell_linux_amd64.tar.gz | tar xz
sudo mv mcp-shell /usr/local/bin/
```

### macOS
```bash
curl -L https://github.com/rafa-dot-el/mcp-shell/releases/latest/download/mcp-shell_darwin_amd64.tar.gz | tar xz
sudo mv mcp-shell /usr/local/bin/
```

### Windows
Download the Windows binary from the [releases page](https://github.com/rafa-dot-el/mcp-shell/releases/latest) and add it to your PATH.

## Container Image

Run MCP Shell in a container:

```bash
# Docker
docker run --rm -it ghcr.io/rafa-dot-el/mcp-shell:latest

# Podman
podman run --rm -it ghcr.io/rafa-dot-el/mcp-shell:latest
```

## From Source

Build from source using Go:

```bash
git clone https://github.com/rafa-dot-el/mcp-shell.git
cd mcp-shell
go build -o mcp-shell ./cmd/mcp-shell
```

Or with our development environment:

```bash
git clone https://github.com/rafa-dot-el/mcp-shell.git
cd mcp-shell
nix develop
task build
```

## Verification

Verify your installation:

```bash
mcp-shell version
```

## Next Steps

- [Getting Started](/docs/getting-started/): Learn basic usage
- [Configuration](/docs/configuration/): Configure MCP Shell for your needs