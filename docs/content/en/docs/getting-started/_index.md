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
mcp-shell serve --log-level debug
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