---
title: "Configuring Claude Code with MCP Shell"
linkTitle: "Claude Code"
weight: 33
tutorials: ["integration"]
tags: ["claude", "claude-code", "mcp", "integration", "ai", "cli"]
description: "Learn how to integrate MCP Shell with Claude Code for AI-powered shell operations in the terminal"
---

# Configuring Claude Code with MCP Shell

This tutorial shows you how to configure Claude Code (Anthropic's CLI tool) to use MCP Shell as an MCP server, enabling Claude to execute shell scripts and commands directly from your terminal.

## Prerequisites

Before starting, ensure you have:

- Claude Code installed ([Installation guide](https://docs.claude.com/claude-code))
- MCP Shell installed ([Installation guide](/docs/installation/))
- Basic understanding of shell scripts and JSON configuration
- 10-15 minutes of time

## What You'll Learn

By the end of this tutorial, you'll be able to:
- Configure Claude Code to connect to MCP Shell
- Create scripts that Claude can execute from the CLI
- Use Claude's terminal interface to run shell commands safely
- Validate and list available scripts
- Leverage MCP Shell in your command-line workflow

## Architecture Overview

Claude Code connects to MCP Shell via the Model Context Protocol (MCP):

```
Claude Code CLI
    ↓
MCP Protocol (stdio)
    ↓
MCP Shell Server
    ↓
Shell Scripts/Aliases
```

MCP Shell acts as a bridge between Claude's AI capabilities and your system's shell environment, all accessible from your terminal.

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

# CLI-focused aliases
aliases:
  - name: "git-status"
    description: "Show git repository status"
    command: "git status --short --branch"

  - name: "git-log"
    description: "Show recent git commits"
    command: "git log --oneline --graph -10"

  - name: "list-files"
    description: "List files in current directory with details"
    command: "ls -lah"

  - name: "disk-usage"
    description: "Show disk usage summary"
    command: "df -h"

  - name: "process-list"
    description: "List running processes"
    command: "ps aux | head -20"

  - name: "network-connections"
    description: "Show active network connections"
    command: "netstat -tuln 2>/dev/null || ss -tuln"

# CLI-focused scripts with parameters
scripts:
  - name: "search-files"
    description: "Search for files by name pattern"
    path: "/usr/bin/find"
    interpreter: "bash"
    parameters:
      directory:
        description: "Directory to search in"
        required: true
        default: "."
        setter: "{}"
      pattern:
        description: "Filename pattern to match (e.g., '*.go')"
        required: true
        setter: "-name {}"

  - name: "search-content"
    description: "Search for text content in files"
    path: "/usr/bin/grep"
    interpreter: "bash"
    parameters:
      pattern:
        description: "Text pattern to search for"
        required: true
        setter: "-r {}"
      directory:
        description: "Directory to search in"
        required: false
        default: "."
        setter: "{}"
      file_pattern:
        description: "File pattern to search (e.g., '*.go')"
        required: false
        setter: "--include={}"

  - name: "git-find-commits"
    description: "Find commits by message or author"
    path: "/usr/bin/git"
    interpreter: "bash"
    parameters:
      search:
        description: "Search term for commit messages or author"
        required: true
        setter: "log --all --grep={}"

  - name: "check-port"
    description: "Check if a port is in use"
    path: "/usr/bin/lsof"
    interpreter: "bash"
    parameters:
      port:
        description: "Port number to check"
        required: true
        setter: "-i :{}"

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

Before connecting to Claude Code, validate your configuration:

```bash
# Validate configuration
mcp-shell validate

# List available scripts and aliases
mcp-shell list

# Test the server manually
mcp-shell serve
```

Press Ctrl+C to stop the test server.

## Step 4: Configure Claude Code

Claude Code uses a configuration file to define MCP servers.

### Find Configuration File Location

**macOS/Linux:**
```
~/.config/claude/config.json
```

**Windows:**
```
%APPDATA%\Claude\config.json
```

### Create Configuration Directory

```bash
# Create config directory
mkdir -p ~/.config/claude
```

### Edit Configuration

Add MCP Shell to Claude Code's configuration:

```bash
cat > ~/.config/claude/config.json <<'EOF'
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

## Step 5: Start Claude Code

Start a new Claude Code session:

```bash
# Start Claude Code
claude

# Or start in a specific directory
claude --dir /path/to/project
```

Claude Code will automatically connect to configured MCP servers on startup.

## Step 6: Verify Connection

In Claude Code, ask:

```
Can you check what MCP tools are available?
```

Claude should list the tools from your MCP Shell configuration, including:
- Your configured aliases (git-status, git-log, list-files, etc.)
- Your configured scripts (search-files, search-content, check-port, etc.)

## Step 7: Use MCP Shell Through Claude Code

Now you can ask Claude to execute commands from the terminal:

**Example 1: Check git status**
```
User: Can you check the git status of this repository?

Claude: I'll use the git-status tool...
[Claude executes the git-status alias and shows output]
```

**Example 2: List files**
```
User: Show me all files in the current directory

Claude: I'll list the files for you...
[Claude executes the list-files alias]
```

**Example 3: Search for files**
```
User: Find all Go files in the current directory

Claude: I'll search for Go files...
[Claude executes search-files with directory="." and pattern="*.go"]
```

**Example 4: Check if port is in use**
```
User: Is port 8080 in use?

Claude: I'll check that port...
[Claude executes check-port with port=8080]
```

**Example 5: Search code content**
```
User: Find all functions named "Execute" in Go files

Claude: I'll search for that pattern...
[Claude executes search-content with pattern="func Execute" and file_pattern="*.go"]
```

## Advanced Configuration

### Project-Specific Configuration

You can create project-specific MCP configurations:

```bash
# Create project config directory
mkdir -p .claude

# Create project-specific config
cat > .claude/config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "./.mcp-shell-config.yaml"
      }
    }
  }
}
EOF

# Create project-specific MCP Shell config
cat > .mcp-shell-config.yaml <<'EOF'
mcp:
  name: "mcp-shell-project"
  version: "1.0.0"
  transport: "stdio"

aliases:
  - name: "build"
    description: "Build the project"
    command: "task build"

  - name: "test"
    description: "Run all tests"
    command: "task test"

  - name: "lint"
    description: "Run linters"
    command: "task lint"

  - name: "dev"
    description: "Start development server"
    command: "task dev"

execution:
  log_directory: "./.mcp-shell/logs"
  max_parallel_jobs: 5
  default_timeout: "10m"
EOF

# Start Claude Code with project config
claude --config .claude/config.json
```

### Adding Custom Development Scripts

Create custom scripts for your workflow:

```bash
# Create a project analysis script
cat > ~/.local/bin/project-summary.sh <<'EOF'
#!/bin/bash
echo "=== Project Summary ==="
echo "Repository: $(basename $(git rev-parse --show-toplevel 2>/dev/null) || pwd)"
echo "Branch: $(git branch --show-current 2>/dev/null || echo 'Not a git repo')"
echo "Last commit: $(git log -1 --format='%h - %s (%ar)' 2>/dev/null || echo 'No commits')"
echo "Modified files: $(git status --short 2>/dev/null | wc -l || echo 'N/A')"
echo "Lines of code: $(find . -type f \( -name '*.go' -o -name '*.py' -o -name '*.js' \) -exec wc -l {} + 2>/dev/null | tail -1 || echo 'N/A')"
echo "TODO comments: $(grep -r 'TODO\|FIXME' --include='*.go' --include='*.py' --include='*.js' . 2>/dev/null | wc -l || echo '0')"
EOF

chmod +x ~/.local/bin/project-summary.sh
```

Add to your config:

```yaml
scripts:
  - name: "project-summary"
    description: "Display project summary and statistics"
    path: "~/.local/bin/project-summary.sh"
    interpreter: "bash"
```

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
        "MCP_SHELL_EXECUTION_LOG_DIRECTORY": "~/.local/state/mcp-shell/logs",
        "MCP_SHELL_EXECUTION_DEFAULT_TIMEOUT": "10m"
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
    "mcp-shell-ops": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/ops.yaml"
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
ls -l ~/.local/bin/project-summary.sh

# Only use scripts you control
```

### Review Execution Logs

Monitor what Claude executes:

```bash
# View logs
tail -f ~/.local/state/mcp-shell/logs/*.log

# Search for specific operations
grep "executed" ~/.local/state/mcp-shell/logs/*.log

# Watch logs in real-time while using Claude
tail -f ~/.local/state/mcp-shell/logs/*.log | grep --line-buffered "executed"
```

### Limit Permissions

Run scripts with minimal necessary permissions:
- Don't use `sudo` in scripts unless absolutely necessary
- Use unprivileged accounts for execution
- Limit network access where possible
- Avoid destructive operations without user confirmation

## Troubleshooting

### Claude Code Can't Connect to MCP Shell

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

**Verify config file location:**
```bash
cat ~/.config/claude/config.json
```

### Tools Not Appearing in Claude Code

**Verify configuration is loaded:**
```bash
mcp-shell list
```

**Check MCP Shell logs:**
```bash
tail -f ~/.local/state/mcp-shell/logs/*.log
```

**Restart Claude Code:**
```bash
# Exit current session
exit

# Start new session
claude
```

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

**Test script manually:**
```bash
/path/to/script.sh
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

### Configuration Not Loading

**Check config file format:**
```bash
# Validate JSON syntax
cat ~/.config/claude/config.json | jq .
```

**Check MCP Shell config:**
```bash
mcp-shell validate --config ~/.config/mcp-shell/config.yaml
```

## Example Workflows

### Development Workflow

```bash
# Start Claude Code in your project
cd ~/projects/my-app
claude

# Ask Claude to help
User: Can you check git status and run the tests?
Claude: [Executes git-status and test aliases]

User: Find all TODO comments in the codebase
Claude: [Executes search-content with pattern="TODO"]
```

### System Administration

```bash
claude

User: Check disk usage and list running processes
Claude: [Executes disk-usage and process-list]

User: Show me active network connections
Claude: [Executes network-connections]
```

### Code Investigation

```bash
claude

User: Find all files that import "fmt" in Go files
Claude: [Executes search-content with pattern="import.*fmt" and file_pattern="*.go"]

User: Show me recent commits that mention "bug"
Claude: [Executes git-find-commits with search="bug"]
```

### Project Analysis

```bash
claude

User: Give me a summary of this project
Claude: [Executes project-summary script and analyzes output]

User: Check if port 3000 is in use
Claude: [Executes check-port with port=3000]
```

## Integration with Claude Code Features

### Inline Execution

Claude Code can execute MCP Shell commands inline with your conversation:

```
User: I'm working on a Go project. Can you check the git status and find all TODO comments?

Claude: I'll check both for you...
[Executes git-status and search-content sequentially]
[Provides analysis of both outputs]
```

### Multi-Step Workflows

Claude can chain multiple commands:

```
User: Build the project, run tests, and if everything passes, show me the git status

Claude: I'll execute those steps in order...
[Executes build, then test, then git-status]
[Reports on each step]
```

### Context-Aware Execution

Claude understands your context and suggests appropriate commands:

```
User: I need to debug why my server won't start

Claude: Let me check if the port is already in use. What port is your server using?

User: Port 8080

Claude: [Executes check-port with port=8080]
[Analyzes output and suggests next steps]
```

## Docker-Based MCP Server

For containerized environments, you can run MCP Shell in Docker:

```bash
# Pull the image
docker pull ghcr.io/rafa-dot-el/mcp-shell:latest

# Configure Claude Code to use Docker
cat > ~/.config/claude/config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v", "${HOME}/.config/mcp-shell:/config:ro",
        "-v", "${HOME}/.local/state/mcp-shell:/logs",
        "-v", "${PWD}:/workspace",
        "-w", "/workspace",
        "ghcr.io/rafa-dot-el/mcp-shell:latest",
        "serve"
      ],
      "env": {
        "MCP_SHELL_CONFIG": "/config/config.yaml"
      }
    }
  }
}
EOF
```

**Benefits:**
- Consistent environment across machines
- Isolation from host system
- Easy version management
- Reproducible configurations
- No need to install MCP Shell locally

**Using with Docker:**
```bash
# Start Claude Code
claude

User: Can you check the git status?
Claude: [Uses Docker container to execute git-status]
```

## Tips and Best Practices

### Efficient Commands

Keep aliases short and focused:
```yaml
aliases:
  - name: "gs"  # Short alias for git status
    description: "Git status"
    command: "git status --short"
```

### Descriptive Names

Use clear, descriptive names for scripts:
```yaml
scripts:
  - name: "find-large-files"  # Clear purpose
    description: "Find files larger than 10MB"
    # ...
```

### Safe Defaults

Provide safe default values:
```yaml
parameters:
  directory:
    required: false
    default: "."  # Safe default
```

### Logging

Enable detailed logging for debugging:
```yaml
logging:
  level: "debug"  # For troubleshooting
  format: "json"  # For log analysis
```

### Testing

Test your scripts before adding to MCP Shell:
```bash
# Test script manually first
~/.local/bin/my-script.sh

# Validate config
mcp-shell validate

# Test with MCP server
mcp-shell serve
```

## Next Steps

- [Configuration Guide](/docs/configuration/): Learn about advanced configuration options
- [Advanced Configuration Tutorial](/tutorials/advanced-configuration/): Master complex setups
- [Security Best Practices](/docs/security/): Secure your MCP Shell deployment
- [Claude Desktop Tutorial](/tutorials/claude-desktop/): Use MCP Shell with Claude Desktop

## Resources

- [Claude Code Documentation](https://docs.claude.com/claude-code)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [MCP Shell GitHub](https://github.com/rafa-dot-el/mcp-shell)
- [Example Configurations](https://github.com/rafa-dot-el/mcp-shell/tree/main/examples)
- [Claude API Documentation](https://docs.anthropic.com)
