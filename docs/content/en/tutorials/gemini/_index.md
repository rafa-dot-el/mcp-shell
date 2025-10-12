---
title: "Configuring Google Gemini with MCP Shell"
linkTitle: "Gemini"
weight: 31
tutorials: ["integration"]
tags: ["gemini", "mcp", "integration", "ai", "google"]
description: "Learn how to integrate MCP Shell with Google Gemini for AI-powered shell operations"
---

# Configuring Google Gemini with MCP Shell

This tutorial shows you how to configure Google Gemini to use MCP Shell as an MCP server, enabling Gemini to execute shell scripts and commands on your behalf.

## Prerequisites

Before starting, ensure you have:

- Google AI Studio account ([Sign up here](https://ai.google.dev/))
- MCP Shell installed ([Installation guide](/docs/installation/))
- Node.js 18+ installed (for Gemini MCP client)
- Basic understanding of shell scripts and JSON configuration
- 15-20 minutes of time

## What You'll Learn

By the end of this tutorial, you'll be able to:
- Configure Gemini to connect to MCP Shell
- Create scripts that Gemini can execute
- Use Gemini to run shell commands safely
- Validate and list available scripts
- Integrate MCP Shell with Gemini API

## Architecture Overview

Gemini connects to MCP Shell via the Model Context Protocol (MCP):

```
Gemini API / AI Studio
       ↓
   MCP Client (Node.js)
       ↓
   MCP Protocol (stdio/HTTP)
       ↓
   MCP Shell Server
       ↓
   Shell Scripts/Aliases
```

MCP Shell acts as a bridge between Gemini's AI capabilities and your system's shell environment.

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

  - name: "check-memory"
    description: "Display memory usage"
    command: "free -h"

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

  - name: "system-info"
    description: "Display system information"
    path: "/usr/bin/uname"
    interpreter: "bash"
    parameters:
      option:
        description: "Information to display (all, kernel, hardware)"
        required: false
        default: "-a"
        valid_values: ["-a", "-s", "-m", "-r"]
        setter: "{}"

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

Before connecting to Gemini, validate your configuration:

```bash
# Validate configuration
mcp-shell validate

# List available scripts and aliases
mcp-shell list

# Test the server manually
mcp-shell serve
```

Press Ctrl+C to stop the test server.

## Step 4: Set Up MCP Client for Gemini

Gemini requires an MCP client to communicate with MCP servers. Create a Node.js client:

```bash
# Create project directory
mkdir -p ~/gemini-mcp-client
cd ~/gemini-mcp-client

# Initialize Node.js project
npm init -y

# Install MCP SDK and Gemini API
npm install @modelcontextprotocol/sdk @google/generative-ai
```

## Step 5: Create MCP Client Configuration

Create a configuration file for the MCP client:

```bash
cat > ~/gemini-mcp-client/mcp-config.json <<'EOF'
{
  "mcpServers": {
    "mcp-shell": {
      "command": "mcp-shell",
      "args": ["serve"],
      "env": {
        "MCP_SHELL_CONFIG": "~/.config/mcp-shell/config.yaml",
        "MCP_SHELL_LOGGING_LEVEL": "info"
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

## Step 6: Create Gemini MCP Client

Create a Node.js script to connect Gemini to MCP Shell:

```bash
cat > ~/gemini-mcp-client/gemini-mcp.js <<'EOF'
import { GoogleGenerativeAI } from '@google/generative-ai';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import { spawn } from 'child_process';
import { readFileSync } from 'fs';

// Load configuration
const config = JSON.parse(readFileSync('./mcp-config.json', 'utf-8'));
const mcpServerConfig = config.mcpServers['mcp-shell'];

// Initialize Gemini
const genAI = new GoogleGenerativeAI(process.env.GEMINI_API_KEY);
const model = genAI.getGenerativeModel({ model: 'gemini-pro' });

// Initialize MCP client
const mcpClient = new Client({
  name: 'gemini-mcp-client',
  version: '1.0.0',
});

async function startMCPServer() {
  const serverProcess = spawn(mcpServerConfig.command, mcpServerConfig.args, {
    env: { ...process.env, ...mcpServerConfig.env },
  });

  const transport = new StdioClientTransport({
    stdin: serverProcess.stdout,
    stdout: serverProcess.stdin,
  });

  await mcpClient.connect(transport);
  console.log('Connected to MCP Shell server');

  return serverProcess;
}

async function listAvailableTools() {
  const tools = await mcpClient.listTools();
  console.log('\nAvailable tools:');
  tools.tools.forEach(tool => {
    console.log(`- ${tool.name}: ${tool.description}`);
  });
  return tools.tools;
}

async function executeCommand(toolName, args) {
  console.log(`\nExecuting: ${toolName}`);
  const result = await mcpClient.callTool({
    name: toolName,
    arguments: args,
  });
  console.log('Result:', result);
  return result;
}

async function geminiWithMCP(prompt) {
  // Get available tools
  const tools = await listAvailableTools();

  // Convert MCP tools to Gemini function declarations
  const functionDeclarations = tools.map(tool => ({
    name: tool.name,
    description: tool.description,
    parameters: tool.inputSchema,
  }));

  // Create chat with function calling
  const chat = model.startChat({
    tools: [{ functionDeclarations }],
  });

  // Send message
  const result = await chat.sendMessage(prompt);
  const response = result.response;

  // Check if Gemini wants to call a function
  const functionCall = response.functionCalls()?.[0];
  if (functionCall) {
    console.log(`\nGemini wants to call: ${functionCall.name}`);
    console.log('Arguments:', functionCall.args);

    // Execute the tool via MCP
    const toolResult = await executeCommand(functionCall.name, functionCall.args);

    // Send result back to Gemini
    const finalResult = await chat.sendMessage([{
      functionResponse: {
        name: functionCall.name,
        response: toolResult,
      },
    }]);

    return finalResult.response.text();
  }

  return response.text();
}

// Main execution
async function main() {
  const serverProcess = await startMCPServer();

  try {
    // Example: Ask Gemini to check git status
    const prompt = process.argv[2] || 'Check the git status of the current directory';
    console.log(`\nPrompt: ${prompt}\n`);

    const response = await geminiWithMCP(prompt);
    console.log('\nGemini response:', response);
  } finally {
    serverProcess.kill();
    process.exit(0);
  }
}

main().catch(console.error);
EOF
```

## Step 7: Set Up Gemini API Key

Obtain an API key from Google AI Studio and configure it:

```bash
# Get API key from https://ai.google.dev/
export GEMINI_API_KEY="your-api-key-here"

# Add to your shell profile for persistence
echo 'export GEMINI_API_KEY="your-api-key-here"' >> ~/.bashrc
```

## Step 8: Run Gemini with MCP Shell

Execute the client to use Gemini with MCP Shell:

```bash
cd ~/gemini-mcp-client

# Run with default prompt
node gemini-mcp.js

# Run with custom prompt
node gemini-mcp.js "Show me disk usage and list files"

# Run with specific command
node gemini-mcp.js "Search for all Python files in /home/user/projects"
```

## Example Usage

### Example 1: Check git status

```bash
node gemini-mcp.js "Can you check the git status?"
```

**Output:**
```
Prompt: Can you check the git status?

Available tools:
- git-status: Show git repository status
- list-files: List files in current directory
- disk-usage: Show disk usage
- search-files: Search for files by name

Gemini wants to call: git-status
Arguments: {}

Executing: git-status
Result: { content: '## main...origin/main\n M docs/config.yaml\n?? test.sh' }

Gemini response: The git repository is on the main branch, tracking origin/main.
There is one modified file (docs/config.yaml) and one untracked file (test.sh).
```

### Example 2: System information

```bash
node gemini-mcp.js "What's the system information?"
```

### Example 3: Search for files

```bash
node gemini-mcp.js "Find all markdown files in the current directory"
```

## Advanced Configuration

### Using MCP Shell with Gemini API in Python

You can also use Python to integrate Gemini with MCP Shell:

```python
#!/usr/bin/env python3
import google.generativeai as genai
import subprocess
import json
import os

# Configure Gemini
genai.configure(api_key=os.environ['GEMINI_API_KEY'])
model = genai.GenerativeModel('gemini-pro')

# Start MCP Shell server
mcp_process = subprocess.Popen(
    ['mcp-shell', 'serve'],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    env={**os.environ, 'MCP_SHELL_CONFIG': '~/.config/mcp-shell/config.yaml'}
)

# List available tools
# (MCP protocol communication would go here)

# Use Gemini with function calling
# (Implementation would integrate MCP tools with Gemini function calling)
```

### Environment Variables for MCP Client

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

### Adding Custom Scripts

Create a custom script and add it to your configuration:

```bash
# Create a script
cat > ~/.local/bin/git-summary.sh <<'EOF'
#!/bin/bash
echo "=== Git Repository Summary ==="
echo "Branch: $(git branch --show-current)"
echo "Remote: $(git remote get-url origin 2>/dev/null || echo 'No remote')"
echo "Commits: $(git rev-list --count HEAD 2>/dev/null || echo '0')"
echo "Uncommitted changes: $(git status --short | wc -l)"
echo "Last commit: $(git log -1 --format='%h - %s (%ar)' 2>/dev/null || echo 'No commits')"
EOF

chmod +x ~/.local/bin/git-summary.sh
```

Add to your config:

```yaml
scripts:
  - name: "git-summary"
    description: "Display git repository summary"
    path: "~/.local/bin/git-summary.sh"
    interpreter: "bash"
```

Restart the MCP client to pick up the changes.

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

### API Key Security

Protect your Gemini API key:

```bash
# Store in environment variable, never in code
export GEMINI_API_KEY="your-key-here"

# Use environment files
cat > ~/gemini-mcp-client/.env <<'EOF'
GEMINI_API_KEY=your-key-here
EOF

# Add to .gitignore
echo ".env" >> ~/gemini-mcp-client/.gitignore
```

### Restrict Script Locations

Use absolute paths and verify script ownership:

```bash
# Check script ownership
ls -l ~/.local/bin/git-summary.sh

# Only use scripts you control
```

### Review Execution Logs

Monitor what Gemini executes:

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

### MCP Client Can't Connect to MCP Shell

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

### Gemini API Key Issues

**Verify API key is set:**
```bash
echo $GEMINI_API_KEY
```

**Test API key:**
```bash
curl -H "Content-Type: application/json" \
  -d '{"contents":[{"parts":[{"text":"Hello"}]}]}' \
  "https://generativelanguage.googleapis.com/v1/models/gemini-pro:generateContent?key=$GEMINI_API_KEY"
```

### Tools Not Appearing

**Verify configuration is loaded:**
```bash
mcp-shell list
```

**Check MCP Shell logs:**
```bash
tail -f ~/.local/state/mcp-shell/logs/*.log
```

**Restart MCP client:**
```bash
# Kill any running instances
pkill -f "node gemini-mcp.js"

# Restart
node gemini-mcp.js
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

### Node.js Module Errors

**Reinstall dependencies:**
```bash
cd ~/gemini-mcp-client
rm -rf node_modules package-lock.json
npm install
```

**Check Node.js version:**
```bash
node --version  # Should be 18+
```

## Example Workflows

### Development Workflow

Ask Gemini to help with development tasks:

```bash
node gemini-mcp.js "Show me the git status and list recent files"
```

### System Administration

```bash
node gemini-mcp.js "Check disk usage and display system information"
```

### File Management

```bash
node gemini-mcp.js "Find all markdown files in my documents folder"
```

### Automated Reports

```bash
# Create a script that runs periodically
cat > ~/gemini-mcp-client/daily-report.sh <<'EOF'
#!/bin/bash
node gemini-mcp.js "Create a summary of git activity and system status" > daily-report.txt
cat daily-report.txt
EOF

chmod +x ~/gemini-mcp-client/daily-report.sh
```

## Integration with Gemini AI Studio

You can also use Gemini AI Studio's web interface with MCP Shell by deploying the MCP client as a web service:

1. Create an Express.js server that wraps the MCP client
2. Expose an API endpoint that Gemini can call
3. Configure Gemini AI Studio to use your endpoint
4. Use webhooks or polling to get responses

**Example Express server:**

```javascript
import express from 'express';
// ... MCP client code from above ...

const app = express();
app.use(express.json());

app.post('/execute', async (req, res) => {
  const { prompt } = req.body;
  const response = await geminiWithMCP(prompt);
  res.json({ response });
});

app.listen(3000, () => {
  console.log('MCP client listening on port 3000');
});
```

## Next Steps

- [Configuration Guide](/docs/configuration/): Learn about advanced configuration options
- [Advanced Configuration Tutorial](/tutorials/advanced-configuration/): Master complex setups
- [Security Best Practices](/docs/security/): Secure your MCP Shell deployment

## Resources

- [Gemini API Documentation](https://ai.google.dev/docs)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [MCP Shell GitHub](https://github.com/rafa-dot-el/mcp-shell)
- [Example Configurations](https://github.com/rafa-dot-el/mcp-shell/tree/main/examples)
- [Google AI Studio](https://ai.google.dev/)
