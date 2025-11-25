---
title: "Configuring Cursor IDE with MCP Shell"
linkTitle: "Cursor"
weight: 32
tutorials: ["integration"]
tags: ["cursor", "mcp", "integration", "ai", "ide"]
description: "Learn how to integrate MCP Shell with Cursor IDE for AI-powered shell operations"
---

# Configuring Cursor IDE with MCP Shell

This tutorial shows you how to configure Cursor IDE to use MCP Shell as an MCP server, enabling Cursor's AI to execute shell scripts and commands on your behalf.

## Prerequisites

Before starting, ensure you have:

- Cursor IDE installed ([Download here](https://cursor.sh/))
- MCP Shell installed ([Installation guide](/docs/installation/))
- Basic understanding of shell scripts and JSON configuration
- 10-15 minutes of time

## What You'll Learn

By the end of this tutorial, you'll be able to:
- Configure Cursor IDE to connect to MCP Shell
- Create scripts that Cursor can execute
- Use Cursor's AI to run shell commands safely
- Validate and list available scripts
- Leverage MCP Shell in your development workflow

## Architecture Overview

Cursor IDE connects to MCP Shell via the Model Context Protocol (MCP):

```
Cursor IDE
    ↓
MCP Protocol (stdio)
    ↓
MCP Shell Server
    ↓
Shell Scripts/Aliases
```

MCP Shell acts as a bridge between Cursor's AI capabilities and your system's shell environment.

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

# Development-focused aliases
aliases:
  - name: "git-status"
    description: "Show git repository status"
    command: "git status --short --branch"

  - name: "git-log"
    description: "Show recent git commits"
    command: "git log --oneline -10"

  - name: "list-files"
    description: "List files in current directory"
    command: "ls -lah"

  - name: "find-todos"
    description: "Find TODO comments in code"
    command: "grep -rn 'TODO\\|FIXME\\|XXX' --include='*.go' --include='*.js' --include='*.py' ."

  - name: "run-tests"
    description: "Run project tests"
    command: "go test ./... -v"

# Development scripts with parameters
scripts:
  - name: "search-code"
    description: "Search for code patterns"
    path: "/usr/bin/grep"
    interpreter: "bash"
    parameters:
      pattern:
        description: "Pattern to search for"
        required: true
        setter: "-rn {} ."
      file_type:
        description: "File extension to search (e.g., go, js, py)"
        required: false
        setter: "--include='*.{}'"

  - name: "git-diff-files"
    description: "Show diff for specific files"
    path: "/usr/bin/git"
    interpreter: "bash"
    parameters:
      file:
        description: "File to show diff for"
        required: true
        setter: "diff {}"

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

Before connecting to Cursor, validate your configuration:

```bash
# Validate configuration
mcp-shell validate

# List available scripts and aliases
mcp-shell list

# Test the server manually
mcp-shell serve
```

Press Ctrl+C to stop the test server.

## Step 4: Configure Cursor IDE

Cursor IDE uses a configuration file to define MCP servers.

### Find Configuration File Location

**macOS:**
```
~/Library/Application Support/Cursor/User/globalStorage/mcp-config.json
```

**Linux:**
```
~/.config/Cursor/User/globalStorage/mcp-config.json
```

**Windows:**
```
%APPDATA%\Cursor\User\globalStorage\mcp-config.json
```

### Edit Configuration

Add MCP Shell to Cursor's configuration:

```bash
# macOS
mkdir -p ~/Library/Application\ Support/Cursor/User/globalStorage
cat > ~/Library/Application\ Support/Cursor/User/globalStorage/mcp-config.json <<'EOF'
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

# Linux
mkdir -p ~/.config/Cursor/User/globalStorage
cat > ~/.config/Cursor/User/globalStorage/mcp-config.json <<'EOF'
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

## Step 5: Restart Cursor IDE

After saving the configuration:

1. **Quit Cursor IDE** completely (not just close the window)
2. **Reopen Cursor IDE**
3. Wait a few seconds for MCP servers to initialize

## Step 6: Verify Connection

In Cursor IDE, open the AI chat panel and ask:

```
Can you check what MCP tools are available?
```

Cursor should list the tools from your MCP Shell configuration, including:
- Your configured aliases (git-status, git-log, list-files, etc.)
- Your configured scripts (search-code, git-diff-files, check-port)

## Step 7: Use MCP Shell Through Cursor

Now you can ask Cursor to execute commands:

**Example 1: Check git status**
```
User: Can you check the git status of this project?

Cursor: I'll use the git-status tool...
[Cursor executes the git-status alias]
```

**Example 2: Find TODOs in code**
```
User: Find all TODO comments in the codebase

Cursor: I'll search for TODO comments...
[Cursor executes the find-todos alias]
```

**Example 3: Search for code patterns**
```
User: Search for all functions named "handleError" in Go files

Cursor: I'll search for that pattern...
[Cursor executes search-code with parameters]
```

**Example 4: Check if port is in use**
```
User: Is port 8080 in use?

Cursor: I'll check that port...
[Cursor executes check-port with port=8080]
```

## Advanced Configuration

### Project-Specific Configuration

You can create project-specific MCP configurations:

```bash
# Create project config
cat > .cursor/mcp-config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell-project": {
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
    command: "make build"

  - name: "test"
    description: "Run project tests"
    command: "make test"

  - name: "lint"
    description: "Run linters"
    command: "make lint"

execution:
  log_directory: "./.mcp-shell/logs"
  max_parallel_jobs: 5
  default_timeout: "10m"
EOF
```

### Adding Development Scripts

Create custom development scripts:

```bash
# Create a project analysis script
cat > ~/.local/bin/analyze-project.sh <<'EOF'
#!/bin/bash
echo "=== Project Analysis ==="
echo "Lines of code: $(find . -name '*.go' -o -name '*.js' -o -name '*.py' | xargs wc -l | tail -1)"
echo "Number of files: $(find . -type f -name '*.go' -o -name '*.js' -o -name '*.py' | wc -l)"
echo "Git commits: $(git rev-list --count HEAD 2>/dev/null || echo '0')"
echo "Open issues (TODO): $(grep -r 'TODO' --include='*.go' --include='*.js' --include='*.py' . | wc -l)"
echo "Recent contributors: $(git shortlog -sn --all | head -5)"
EOF

chmod +x ~/.local/bin/analyze-project.sh
```

Add to your config:

```yaml
scripts:
  - name: "analyze-project"
    description: "Analyze project metrics"
    path: "~/.local/bin/analyze-project.sh"
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
ls -l ~/.local/bin/analyze-project.sh

# Only use scripts you control
```

### Review Execution Logs

Monitor what Cursor executes:

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

### Cursor Can't Connect to MCP Shell

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

### Tools Not Appearing in Cursor

**Verify configuration is loaded:**
```bash
mcp-shell list
```

**Check Cursor's logs:**

macOS:
```bash
tail -f ~/Library/Logs/Cursor/mcp*.log
```

Linux:
```bash
tail -f ~/.config/Cursor/logs/mcp*.log
```

**Restart Cursor IDE:**
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

### Code Review Workflow

Ask Cursor to help with code review:

```
User: Can you check the git status and show me what files changed?

Cursor: [Executes git-status and analyzes changes]
```

### Testing Workflow

```
User: Run the tests and show me any TODOs in the codebase

Cursor: [Executes run-tests and find-todos]
```

### Debugging Workflow

```
User: Check if port 8080 is in use and show me the process

Cursor: [Executes check-port with port 8080]
```

### Project Analysis

```
User: Analyze this project and give me an overview

Cursor: [Executes analyze-project and provides insights]
```

## Integration with Cursor Features

### Composer Mode

In Composer mode, Cursor can use MCP Shell to:
- Check git status before committing
- Run tests before pushing
- Search for similar code patterns
- Verify builds before deployment

### Chat Mode

In Chat mode, you can ask Cursor to:
- Execute shell commands and explain output
- Find files and show their contents
- Check system status
- Run development workflows

### Inline Commands

Use Cursor's inline commands with MCP Shell:
- `Cmd+K` -> "Run tests for this file"
- `Cmd+L` -> "Check git diff for this file"

## Docker-Based MCP Server

For containerized environments, you can run MCP Shell in Docker:

```bash
# Pull the image
docker pull ghcr.io/rafa-dot-el/mcp-shell:latest

# Configure Cursor to use Docker
cat > ~/.config/Cursor/User/globalStorage/mcp-config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v", "${HOME}/.config/mcp-shell:/config",
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

## Next Steps

- [Configuration Guide](/docs/configuration/): Learn about advanced configuration options
- [Advanced Configuration Tutorial](/tutorials/advanced-configuration/): Master complex setups
- [Security Best Practices](/docs/security/): Secure your MCP Shell deployment

## Resources

- [Cursor IDE Documentation](https://cursor.sh/docs)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [MCP Shell GitHub](https://github.com/rafa-dot-el/mcp-shell)
- [Example Configurations](https://github.com/rafa-dot-el/mcp-shell/tree/main/examples)
