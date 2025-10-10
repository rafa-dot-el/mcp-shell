---
title: "Basic Usage Tutorial"
linkTitle: "Basic Usage"
weight: 10
tutorials: ["beginner"]
tags: ["getting-started", "basics"]
description: "Learn the fundamentals of MCP Shell through practical examples"
---

# Basic Usage Tutorial

This tutorial walks you through the essential operations of MCP Shell, from installation to running your first commands.

## Prerequisites

Before starting this tutorial, you should have:

- Basic command line knowledge
- A supported operating system (Linux, macOS, Windows, FreeBSD)
- 5-10 minutes of time

## Step 1: Installation

First, install MCP Shell using your preferred method:

```bash
# Using Nix (recommended)
nix profile install github:rafa-dot-el/mcp-shell

# Or download binary from releases
wget https://github.com/rafa-dot-el/mcp-shell/releases/latest/download/mcp-shell-linux-amd64
chmod +x mcp-shell-linux-amd64
sudo mv mcp-shell-linux-amd64 /usr/local/bin/mcp-shell
```

## Step 2: Verify Installation

Check that MCP Shell is installed correctly:

```bash
mcp-shell version
```

You should see version information displayed.

## Step 3: View Help

Explore available commands:

```bash
mcp-shell --help
```

This shows all available commands and options.

## Step 4: Basic Configuration

Create a basic configuration file:

```bash
mkdir -p ~/.config/mcp-shell
cat > ~/.config/mcp-shell/config.yaml <<EOF
log_level: info
output_format: text
EOF
```

## Step 5: Run Your First Command

Try running a basic command:

```bash
mcp-shell version
```

## Next Steps

Now that you understand the basics:

- Explore [Advanced Configuration](/tutorials/advanced-configuration/)
- Read the [Configuration Documentation](/docs/configuration/)
- Check out [Development Guide](/docs/development/)

## Troubleshooting

**Command not found?**
- Ensure the binary is in your PATH
- Try using the full path to the binary

**Permission denied?**
- Make sure the binary is executable: `chmod +x mcp-shell`
- On Linux/macOS, you may need sudo for system-wide installation
