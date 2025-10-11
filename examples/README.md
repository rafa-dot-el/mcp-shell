# MCP Shell Examples

This directory contains example configurations and scripts demonstrating MCP Shell functionality.

## Quick Start

1. **Build the binary**:
   ```bash
   task build
   # or: go build -o mcp-shell ./cmd/mcp-shell
   ```

2. **Run the server with example configuration**:
   ```bash
   ./mcp-shell serve --config examples/config.yaml --verbose
   ```

3. **Test the server** (in another terminal):
   ```bash
   # Initialize the connection
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./mcp-shell serve --config examples/config.yaml

   # List available resources
   echo '{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}' | ./mcp-shell serve --config examples/config.yaml

   # List available tools
   echo '{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}' | ./mcp-shell serve --config examples/config.yaml

   # Execute an alias
   echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"execute_alias","arguments":{"name":"git-status"}}}' | ./mcp-shell serve --config examples/config.yaml

   # Execute a script with parameters
   echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"execute_script","arguments":{"name":"hello","parameters":{"name":"MCP"}}}}' | ./mcp-shell serve --config examples/config.yaml
   ```

## Configuration Overview

The `config.yaml` demonstrates all major MCP Shell features:

### Aliases
Simple one-line commands with no parameters:
- `git-status` - Show git repository status
- `list-files` - List files in current directory
- `disk-usage` - Show disk usage
- `memory-info` - Show memory usage

### Scripts
Executable files with parameters:
- `hello` - Simple greeting script
  - Parameter: `name` (optional, default: "World")

- `system-info` - Display system information
  - Parameter: `format` (optional, values: "text" or "json")

- `backup-example` - Demonstration backup script
  - Parameter: `source` (required) - Source directory
  - Parameter: `destination` (required) - Destination directory
  - Parameter: `compress` (optional, default: "true")

### Script Folders
Auto-discovery of scripts in directories:
- `example-scripts` - Discovers all `.sh` files in `examples/scripts/`

## Example Scripts

### hello.sh
Simple script demonstrating parameter handling:
```bash
./examples/scripts/hello.sh "Alice"
```

### system-info.sh
Display system information in text or JSON format:
```bash
./examples/scripts/system-info.sh --format text
./examples/scripts/system-info.sh --format json
```

### backup-example.sh
Simulate a backup operation (doesn't actually backup):
```bash
./examples/scripts/backup-example.sh --source /tmp --dest /tmp/backups --compress true
```

## Testing with MCP Protocol

### 1. Initialize Connection
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }
}
```

### 2. List Available Scripts
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "list_scripts",
    "arguments": {}
  }
}
```

### 3. Execute Script with Parameters
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "execute_script",
    "arguments": {
      "name": "hello",
      "parameters": {
        "name": "Claude"
      }
    }
  }
}
```

### 4. Execute Alias
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "execute_alias",
    "arguments": {
      "name": "git-status"
    }
  }
}
```

### 5. Get Job Status
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "get_job",
    "arguments": {
      "job_id": "abc12345"
    }
  }
}
```

## Directory Structure

```
examples/
├── config.yaml           # Example configuration
├── scripts/              # Example scripts
│   ├── hello.sh         # Simple greeting script
│   ├── system-info.sh   # System information script
│   └── backup-example.sh # Backup simulation script
├── logs/                 # Job execution logs (created automatically)
└── README.md            # This file
```

## Configuration Details

### Execution Settings
- `max_parallel_jobs: 5` - Maximum 5 concurrent jobs
- `default_timeout: 300` - Scripts timeout after 5 minutes
- `log_directory: ./examples/logs` - Logs stored in examples/logs
- `allow_background: true` - Background execution enabled

### Security Settings
- `allow_script_creation: false` - Model cannot create new scripts
- `allowed_interpreters: [bash, sh, python3, perl]` - Allowed interpreters
- `script_creation_path: ./examples/user-scripts/` - Where scripts can be created

### Logging
- `level: info` - Log level (trace, debug, info, warn, error, fatal, panic)
- `format: text` - Log format (text or json)

## Customization

1. **Add your own scripts**:
   - Place scripts in `examples/scripts/`
   - Update `config.yaml` to reference them
   - Make scripts executable: `chmod +x examples/scripts/your-script.sh`

2. **Modify aliases**:
   - Edit the `aliases` section in `config.yaml`
   - Add your frequently-used commands

3. **Adjust security settings**:
   - Enable `allow_script_creation: true` to let the model create scripts
   - Modify `allowed_interpreters` to restrict/expand allowed languages

4. **Change execution limits**:
   - Adjust `max_parallel_jobs` based on your system resources
   - Modify `default_timeout` for longer-running scripts

## Troubleshooting

### Scripts not found
- Verify script paths are correct relative to where you run `mcp-shell`
- Check scripts are executable: `chmod +x examples/scripts/*.sh`
- Use absolute paths if needed

### Permission denied
- Make scripts executable: `chmod +x examples/scripts/*.sh`
- Check security settings in `config.yaml`
- Verify interpreter is in `allowed_interpreters` list

### Log directory errors
- Ensure log directory exists and is writable
- Create it manually: `mkdir -p examples/logs`
- Or update `execution.log_directory` in config

## Integration with Claude Desktop

To use this MCP server with Claude Desktop, add to your Claude configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "shell": {
      "command": "/path/to/mcp-shell",
      "args": ["serve", "--config", "/path/to/examples/config.yaml"]
    }
  }
}
```

Replace `/path/to/` with the actual paths on your system.

## Next Steps

1. Review the example scripts to understand parameter handling
2. Customize `config.yaml` for your use case
3. Add your own scripts and aliases
4. Test with the MCP protocol JSON-RPC requests
5. Integrate with Claude Desktop or other MCP clients
