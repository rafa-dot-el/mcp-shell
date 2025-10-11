---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 20
description: >
  Learn how to use MCP Shell
---

# Getting Started

This guide will help you get started with MCP Shell after installation.

## Basic Usage

Check that MCP Shell is installed correctly:

```bash
mcp-shell --help
```

View the current version:

```bash
mcp-shell version
```

## Running MCP Shell

Start the MCP server:

```bash
mcp-shell serve
```

## Configuration

MCP Shell supports hierarchical configuration:

1. **Command Line Flags**: `--log-level debug`
2. **Environment Variables**: `MCP_SHELL_LOG_LEVEL=debug`
3. **Configuration Files**: YAML format
4. **Default Values**: Built-in defaults

### Configuration File

Create a configuration file at `~/.config/mcp-shell/config.yaml`:

```yaml
# Application configuration
log_level: "info"        # debug, info, warn, error
output_format: "text"    # text, json, yaml

# Add your application-specific configuration here
```

### Environment Variables

Set environment variables with the `MCP_SHELL_` prefix:

```bash
export MCP_SHELL_LOG_LEVEL="debug"
export MCP_SHELL_CONFIG_FILE="/path/to/config.yaml"
```

## Validating Configuration

Before running the server, validate your configuration:

```bash
# Validate default configuration
mcp-shell validate

# Validate specific config file
mcp-shell validate --config /path/to/config.yaml

# Validate with detailed output
mcp-shell validate --verbose
```

The `validate` command checks:
- Configuration file syntax and structure
- Script file existence and executable permissions
- Parameter definitions and valid values
- Alias command syntax
- Log directory accessibility and permissions
- MCP server configuration

**Exit Codes:**
- `0` - Configuration is valid
- `1` - Validation errors found

### Example Validation Output

```bash
$ mcp-shell validate
[+] Configuration validation successful
    Scripts: 3 valid
    Aliases: 2 valid
    Log directory: /var/log/mcp-shell
    Max parallel jobs: 5
```

## Listing Scripts and Aliases

View all configured scripts and aliases:

```bash
# List all in table format (default)
mcp-shell list

# List in JSON format (for programmatic use)
mcp-shell list --format json

# List in simple format (just names)
mcp-shell list --format simple

# List only scripts
mcp-shell list --scripts

# List only aliases
mcp-shell list --aliases

# Show detailed parameter information
mcp-shell list --details
```

### Output Formats

**Table Format** (default):
```
SCRIPTS
--------------------------------------------------------------------------------
NAME            DESCRIPTION                      PATH
backup-db       Backup PostgreSQL database       /opt/scripts/backup.sh
system-monitor  Monitor system resources         /opt/scripts/monitor.py

ALIASES
--------------------------------------------------------------------------------
NAME            DESCRIPTION                      COMMAND
git-status      Show git status                  git status -sb
disk-usage      Show disk usage                  df -h
```

**JSON Format**:
```json
{
  "scripts": [
    {
      "name": "backup-db",
      "description": "Backup PostgreSQL database",
      "path": "/opt/scripts/backup.sh",
      "interpreter": "bash",
      "parameters": {
        "database": {
          "description": "Database name",
          "required": true,
          "setter": "--db {}"
        }
      }
    }
  ],
  "aliases": [
    {
      "name": "git-status",
      "description": "Show git status",
      "command": "git status -sb"
    }
  ]
}
```

**Simple Format**:
```
backup-db
system-monitor
git-status
disk-usage
```

## Examples

### Basic Server

Start a basic MCP server:

```bash
mcp-shell serve --log-level info
```

### Custom Configuration

Use a custom configuration file:

```bash
mcp-shell serve --config /path/to/config.yaml
```

### Debug Mode

Run with verbose logging:

```bash
mcp-shell serve --verbose --log-level debug
```

### Complete Workflow

Typical workflow for starting the server:

```bash
# 1. Validate configuration
mcp-shell validate

# 2. Review available scripts and aliases
mcp-shell list --details

# 3. Start the server
mcp-shell serve
```

## Security Considerations

- MCP Shell follows security best practices
- Commands are executed with appropriate sandboxing
- Regular vulnerability scanning is performed
- Review configuration carefully in production environments

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure MCP Shell has appropriate permissions
2. **Configuration Not Found**: Check file paths and permissions
3. **Port Already in Use**: Specify a different port in configuration

### Getting Help

- Check the logs with `--log-level debug`
- Review the [configuration documentation](/docs/configuration/)
- Report issues on [GitHub](https://github.com/rafa-dot-el/mcp-shell/issues)

## Next Steps

- [Configuration](/docs/configuration/): Learn about advanced configuration options
- [Development](/docs/development/): Contribute to MCP Shell development