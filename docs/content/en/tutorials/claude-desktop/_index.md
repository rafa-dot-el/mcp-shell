---
title: "Configuring Claude Desktop with MCP Shell"
linkTitle: "Claude Desktop"
weight: 30
tutorials: ["integration"]
tags: ["claude", "mcp", "integration", "ai"]
description: "Learn how to integrate MCP Shell with Claude Desktop for AI-powered shell operations"
---

# Configuring Claude Desktop with MCP Shell

This tutorial shows you how to configure Claude Desktop to use MCP Shell as an MCP server, enabling Claude to execute shell scripts and commands on your behalf.

## Prerequisites

Before starting, ensure you have:

- Claude Desktop installed ([Download here](https://claude.ai/download))
- MCP Shell installed ([Installation guide](/docs/installation/))
- Basic understanding of shell scripts and JSON configuration
- 10-15 minutes of time

## What You'll Learn

By the end of this tutorial, you'll be able to:
- Configure Claude Desktop to connect to MCP Shell
- Create scripts that Claude can execute
- Use Claude to run shell commands safely
- Validate and list available scripts

## Architecture Overview

Claude Desktop connects to MCP Shell via the Model Context Protocol (MCP):

```
Claude Desktop App
       ↓
   MCP Protocol (stdio)
       ↓
   MCP Shell Server
       ↓
   Shell Scripts/Aliases
```

MCP Shell acts as a bridge between Claude's AI capabilities and your system's shell environment.

## Step 1: Install MCP Shell

If you haven't already, install MCP Shell:

```bash
# Using Nix (recommended)
nix profile install github:rafa-dot-el/mcp-shell

# Verify installation
mcp-shell version
```

See the [Installation guide](/docs/installation/) for other installation methods.

## Step 2: Create MCP Shell Configuration

Create a configuration file for your scripts and aliases:

```bash
# Create config directory
mkdir -p ~/.config/mcp-shell

# Create configuration file
cat > ~/.config/mcp-shell/config.yaml <<'EOF'
mcp:
  name: "mcp-shell"
  version: "1.0.0"
  transport: "stdio"

# Example aliases for common operations
aliases:
  - name: "git-status"
    description: "Show git repository status"
    command: "git status --short --branch"

  - name: "list-files"
    description: "List files in current directory"
    command: "ls -lah"

  - name: "disk-usage"
    description: "Show disk usage"
    command: "df -h"

# Example scripts with parameters
scripts:
  - name: "search-files"
    description: "Search for files by name"
    path: "/usr/bin/find"
    interpreter: "bash"
    parameters:
      directory:
        description: "Directory to search in"
        required: true
        setter: "{}"
      pattern:
        description: "Filename pattern to match"
        required: true
        setter: "-name {}"

execution:
  log_directory: "~/.local/state/mcp-shell/logs"
  max_parallel_jobs: 5
  default_timeout: "5m"

security:
  allowed_interpreters:
    - "bash"
    - "python3"
    - "node"
  allow_script_creation: false

logging:
  level: "info"
  format: "text"
EOF
```

## Step 3: Validate Configuration

Before connecting to Claude, validate your configuration:

```bash
# Validate configuration
mcp-shell validate

# List available scripts and aliases
mcp-shell list

# Test the server manually
mcp-shell serve
```

Press Ctrl+C to stop the test server.

## Step 4: Configure Claude Desktop

Claude Desktop uses a configuration file to define MCP servers.

### Find Configuration File Location

**macOS:**
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Linux:**
```
~/.config/Claude/claude_desktop_config.json
```

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

### Edit Configuration

Add MCP Shell to Claude's configuration:

```bash
# macOS/Linux
cat > ~/Library/Application\ Support/Claude/claude_desktop_config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/config.yaml"
      }
    }
  }
}
EOF
```

**Key Configuration Options:**

- `command`: Path to mcp-shell binary (use full path if not in PATH)
- `args`: Arguments passed to mcp-shell (always `["serve"]`)
- `env`: Environment variables for MCP Shell
  - `MCP_SHELL_CONFIG`: Custom config file path
  - `MCP_SHELL_LOGGING_LEVEL`: Override log level

**Alternative with full path:**

If `mcp-shell` is not in your PATH:

```json
{
  "mcpServers": {
    "mcp-shell": {
      "command": "/usr/local/bin/mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "/home/user/.config/mcp-shell/config.yaml"
      }
    }
  }
}
```

## Step 5: Restart Claude Desktop

After saving the configuration:

1. **Quit Claude Desktop** completely (not just close the window)
2. **Reopen Claude Desktop**
3. Wait a few seconds for MCP servers to initialize

## Step 6: Verify Connection

In Claude Desktop, ask Claude:

```
Can you check what MCP tools are available?
```

Claude should list the tools from your MCP Shell configuration, including:
- Your configured aliases (git-status, list-files, disk-usage)
- Your configured scripts (search-files)

## Step 7: Use MCP Shell Through Claude

Now you can ask Claude to execute commands:

**Example 1: Check git status**
```
User: Can you check the git status of the current directory?

Claude: I'll use the git-status tool...
[Claude executes the git-status alias]
```

**Example 2: List files**
```
User: Show me all files in the current directory

Claude: I'll list the files for you...
[Claude executes the list-files alias]
```

**Example 3: Search for files**
```
User: Find all Python files in the /home/user/projects directory

Claude: I'll search for Python files...
[Claude executes search-files with parameters]
```

## Advanced Configuration

### Adding Custom Scripts

Create a custom script and add it to your configuration:

```bash
# Create a script
cat > ~/.local/bin/system-info.sh <<'EOF'
#!/bin/bash
echo "=== System Information ==="
echo "Hostname: $(hostname)"
echo "OS: $(uname -s)"
echo "Kernel: $(uname -r)"
echo "Uptime: $(uptime -p)"
echo "Memory: $(free -h | grep Mem | awk '{print $3 "/" $2}')"
EOF

chmod +x ~/.local/bin/system-info.sh
```

Add to your config:

```yaml
scripts:
  - name: "system-info"
    description: "Display system information"
    path: "~/.local/bin/system-info.sh"
    interpreter: "bash"
```

Restart Claude Desktop to pick up the changes.

### Using Environment Variables

Configure MCP Shell behavior via environment variables:

```json
{
  "mcpServers": {
    "mcp-shell": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/config.yaml",
        "MCP_SHELL_LOGGING_LEVEL": "debug",
        "MCP_SHELL_EXECUTION_LOG_DIRECTORY": "~/.local/state/mcp-shell/logs"
      }
    }
  }
}
```

### Multiple MCP Shell Instances

You can run multiple MCP Shell instances with different configurations:

```json
{
  "mcpServers": {
    "mcp-shell-dev": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/dev.yaml"
      }
    },
    "mcp-shell-prod": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/prod.yaml"
      }
    }
  }
}
```

## Security Considerations

### Script Whitelisting

Only allow trusted scripts:

```yaml
security:
  allow_script_creation: false
  allowed_interpreters:
    - "bash"
    # Don't add untrusted interpreters
```

### Restrict Script Locations

Use absolute paths and verify script ownership:

```bash
# Check script ownership
ls -l ~/.local/bin/system-info.sh

# Only use scripts you control
```

### Review Execution Logs

Monitor what Claude executes:

```bash
# View logs
tail -f ~/.local/state/mcp-shell/logs/*.log

# Search for specific operations
grep "executed" ~/.local/state/mcp-shell/logs/*.log
```

### Limit Permissions

Run scripts with minimal necessary permissions:
- Don't use `sudo` in scripts unless absolutely necessary
- Use unprivileged accounts for execution
- Limit network access where possible

## Troubleshooting

### Claude Can't Connect to MCP Shell

**Check if mcp-shell is in PATH:**
```bash
which mcp-shell
```

**Use full path in config if needed:**
```json
"command": "/usr/local/bin/mcp-shell"
```

**Check configuration syntax:**
```bash
mcp-shell validate
```

### Tools Not Appearing in Claude

**Verify configuration is loaded:**
```bash
mcp-shell list
```

**Check Claude's logs (macOS):**
```bash
tail -f ~/Library/Logs/Claude/mcp*.log
```

**Restart Claude Desktop:**
- Fully quit (not just close window)
- Reopen application

### Scripts Failing to Execute

**Check script permissions:**
```bash
ls -l /path/to/script.sh
chmod +x /path/to/script.sh
```

**Verify interpreter is allowed:**
```bash
mcp-shell validate --verbose
```

**Check execution logs:**
```bash
tail -f ~/.local/state/mcp-shell/logs/*.log
```

### Permission Denied Errors

**Check log directory permissions:**
```bash
mkdir -p ~/.local/state/mcp-shell/logs
chmod 755 ~/.local/state/mcp-shell/logs
```

**Verify script ownership:**
```bash
# Scripts should be owned by your user
chown $USER:$USER /path/to/script.sh
```

## Example Workflows

### Development Workflow

Ask Claude to help with development tasks:

```
User: Can you show me the git status and list recent files?

Claude: [Executes git-status and list-files]
```

### System Administration

```
User: Check disk usage and system information

Claude: [Executes disk-usage and system-info scripts]
```

### File Management

```
User: Find all markdown files in my documents folder

Claude: [Executes search-files with appropriate parameters]
```

## Next Steps

- [Configuration Guide](/docs/configuration/): Learn about advanced configuration options
- [Advanced Configuration Tutorial](/tutorials/advanced-configuration/): Master complex setups
- [Security Best Practices](/docs/security/): Secure your MCP Shell deployment

## Resources

- [Claude Desktop Documentation](https://claude.ai/docs)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [MCP Shell GitHub](https://github.com/rafa-dot-el/mcp-shell)
- [Example Configurations](https://github.com/rafa-dot-el/mcp-shell/tree/main/examples)
