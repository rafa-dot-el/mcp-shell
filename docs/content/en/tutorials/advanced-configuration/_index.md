---
title: "Advanced Configuration Tutorial"
linkTitle: "Advanced Configuration"
weight: 20
tutorials: ["advanced"]
tags: ["configuration", "advanced"]
description: "Master MCP Shell configuration with environment variables, multiple config files, and best practices"
---

# Advanced Configuration Tutorial

This tutorial covers advanced configuration techniques for MCP Shell, including environment variables, configuration hierarchies, and optimization strategies.

## Prerequisites

- Completed [Basic Usage Tutorial](/tutorials/basic-usage/)
- Familiarity with YAML and environment variables
- Understanding of command-line flags

## Configuration Hierarchy

MCP Shell uses a hierarchical configuration system (highest to lowest priority):

1. Command-line flags
2. Environment variables
3. Configuration files
4. Default values

## Step 1: Environment Variables

Set configuration through environment variables:

```bash
# Set log level
export MCP_SHELL_LOG_LEVEL=debug

# Set output format
export MCP_SHELL_OUTPUT_FORMAT=json

# Verify settings
mcp-shell version
```

All environment variables follow the pattern: `MCP_SHELL_<SETTING_NAME>`

## Step 2: Multiple Configuration Files

Create environment-specific configurations:

```bash
# Development configuration
cat > ~/.config/mcp-shell/dev.yaml <<EOF
log_level: debug
output_format: json
EOF

# Production configuration
cat > ~/.config/mcp-shell/prod.yaml <<EOF
log_level: warn
output_format: text
EOF

# Use specific config
mcp-shell --config ~/.config/mcp-shell/dev.yaml version
```

## Step 3: Override Priority

Demonstrate configuration priority:

```bash
# Config file sets log_level=info
# Environment variable overrides to debug
export MCP_SHELL_LOG_LEVEL=debug

# Command flag has highest priority
mcp-shell --log-level=error version
```

## Step 4: Configuration Validation

Validate your configuration:

```bash
# Check current configuration
mcp-shell config show

# Test configuration file
mcp-shell --config myconfig.yaml config validate
```

## Best Practices

### Development vs Production

**Development**:
```yaml
log_level: debug
output_format: json
```

**Production**:
```yaml
log_level: warn
output_format: text
```

### Security Considerations

- Never commit sensitive data to configuration files
- Use environment variables for secrets
- Set restrictive permissions: `chmod 600 ~/.config/mcp-shell/config.yaml`

### Performance Optimization

- Use appropriate log levels (info or warn in production)
- Configure output format based on needs
- Consider file size limits for logs

## Advanced Patterns

### Dynamic Configuration

Load configuration based on environment:

```bash
#!/bin/bash
ENV=${APP_ENV:-development}
mcp-shell --config ~/.config/mcp-shell/${ENV}.yaml "$@"
```

### Configuration Templates

Use templates for team standardization:

```yaml
# template.yaml
log_level: ${LOG_LEVEL:-info}
output_format: ${OUTPUT_FORMAT:-text}
```

## Next Steps

- Review [Configuration Documentation](/docs/configuration/)
- Explore [Development Guide](/docs/development/)
- Read about [Architecture](/docs/development/architecture/)

## Troubleshooting

**Environment variables not working?**
- Check variable names follow `MCP_SHELL_*` pattern
- Verify variables are exported: `export MCP_SHELL_LOG_LEVEL=debug`

**Config file ignored?**
- Ensure proper YAML syntax
- Check file permissions are readable
- Verify file path is correct
