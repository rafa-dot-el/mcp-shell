---
title: "Configuration"
linkTitle: "Configuration"
weight: 30
description: >
  Configure MCP Shell for your environment
---

# Configuration

MCP Shell uses a flexible hierarchical configuration system that allows you to customize its behavior through multiple methods.

## Configuration Hierarchy

Configuration values are resolved in the following order (highest to lowest priority):

1. **Command Line Flags** - Explicit flags passed to commands
2. **Environment Variables** - Variables prefixed with `MCP_SHELL_`
3. **Configuration Files** - YAML files in standard locations
4. **Default Values** - Built-in defaults

## Configuration File Locations

MCP Shell searches for configuration files in this order:

1. `--config /path/to/config.yaml` - Explicit path via flag
2. `~/.config/mcp-shell/config.yaml` - User configuration
3. `./mcp-shell.yaml` - Project-local configuration
4. `/etc/mcp-shell/config.yaml` - System-wide configuration

The first file found is used; subsequent files are ignored.

## Configuration File Format

Configuration files use YAML format. Here's a complete example:

```yaml
# MCP Server Configuration
mcp:
  name: "mcp-shell"
  version: "1.0.0"
  transport: "stdio"

# Simple command aliases with no arguments
aliases:
  - name: "git-status"
    description: "Show git repository status"
    command: "git status --short --branch"

  - name: "docker-ps"
    description: "List running containers"
    command: "docker ps --format 'table {{.Names}}\t{{.Status}}'"

# Scripts with parameters
scripts:
  - name: "backup-db"
    description: "Backup PostgreSQL database"
    path: "/opt/scripts/backup.sh"
    interpreter: "bash"
    parameters:
      database:
        description: "Database name to backup"
        required: true
        setter: "--db {}"
      output:
        description: "Output directory"
        required: false
        default: "/backups"
        setter: "--output {}"

# Script folders for auto-discovery
script_folders:
  - name: "admin-scripts"
    description: "Administration scripts"
    path: "/opt/scripts/admin/*.sh"
    default_interpreter: "bash"

# Execution configuration
execution:
  max_parallel_jobs: 5
  default_timeout: "1h"
  log_directory: "/var/log/mcp-shell"
  allow_background: true

# Security configuration
security:
  allow_script_creation: false
  allowed_interpreters:
    - "bash"
    - "python3"
    - "perl"
  script_creation_path: "/opt/mcp-shell/user-scripts"

# Logging configuration
logging:
  level: "info"
  format: "text"
```

## Configuration Sections

### MCP Server Settings

```yaml
mcp:
  name: "mcp-shell"        # Server name
  version: "1.0.0"         # Server version
  transport: "stdio"       # Transport protocol
```

The MCP section configures the Model Context Protocol server settings.

### Aliases

Aliases are simple one-line commands with no parameters:

```yaml
aliases:
  - name: "disk-usage"
    description: "Show disk usage"
    command: "df -h | grep -E '^/dev/'"
```

**Fields:**
- `name` (required): Unique identifier for the alias
- `description` (required): Human-readable description
- `command` (required): Shell command to execute

### Scripts

Scripts are executable files that can accept parameters:

```yaml
scripts:
  - name: "system-backup"
    description: "Backup system files"
    path: "/usr/local/bin/backup.sh"
    interpreter: "bash"
    parameters:
      target:
        description: "Backup target directory"
        required: true
        setter: "--target {}"
      compression:
        description: "Compression level (1-9)"
        required: false
        default: "6"
        valid_values: ["1", "2", "3", "4", "5", "6", "7", "8", "9"]
        setter: "--level {}"
```

**Script Fields:**
- `name` (required): Unique identifier
- `description` (required): Description of what the script does
- `path` (required): Absolute or relative path to script file
- `interpreter` (required): Interpreter to use (must be in allowed list)
- `parameters` (optional): Map of parameter definitions

**Parameter Fields:**
- `description` (required): Parameter description
- `required` (required): Whether parameter is mandatory
- `default` (optional): Default value if not provided
- `valid_values` (optional): List of acceptable values
- `setter` (required): How to pass the value to the script
  - `{}` - positional argument
  - `--flag {}` - flag with value
  - `--flag={}` - flag with equals sign

### Script Folders

Auto-discover scripts in directories using glob patterns:

```yaml
script_folders:
  - name: "maintenance-scripts"
    description: "Database and system maintenance"
    path: "/opt/scripts/maintenance/**/*.sh"
    default_interpreter: "bash"
```

**Fields:**
- `name` (required): Folder identifier
- `description` (required): Folder description
- `path` (required): Glob pattern for script discovery
- `default_interpreter` (required): Default interpreter for discovered scripts

**Supported Patterns:**
- `*.sh` - All `.sh` files in directory
- `**/*.py` - All `.py` files recursively
- `/path/to/scripts/backup*.sh` - Specific pattern matching

### Execution Settings

Control job execution behavior:

```yaml
execution:
  max_parallel_jobs: 5          # Maximum concurrent jobs
  default_timeout: "1h"         # Default execution timeout
  log_directory: "/var/log/mcp-shell"  # Job log storage
  allow_background: true        # Allow background job execution
```

**Timeout Format:**
- `"30s"` - 30 seconds
- `"5m"` - 5 minutes
- `"1h"` - 1 hour
- `"24h"` - 24 hours

### Security Settings

Configure security policies:

```yaml
security:
  allow_script_creation: false  # Allow AI to create new scripts
  allowed_interpreters:         # Permitted interpreters
    - "bash"
    - "sh"
    - "python3"
  script_creation_path: "/opt/mcp-shell/user-scripts"
```

**Security Considerations:**
- Only enable `allow_script_creation` in trusted environments
- Limit `allowed_interpreters` to necessary ones only
- Ensure `script_creation_path` has appropriate permissions
- Review all script configurations before deployment

### Logging Configuration

Configure logging behavior:

```yaml
logging:
  level: "info"    # trace, debug, info, warn, error, fatal, panic
  format: "text"   # text or json
```

**Log Levels:**
- `trace` - Most verbose, includes all details
- `debug` - Debugging information
- `info` - General informational messages (default)
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (application exits)
- `panic` - Panic-level errors

## Environment Variables

Override configuration with environment variables using the `MCP_SHELL_` prefix:

```bash
# Logging
export MCP_SHELL_LOGGING_LEVEL="debug"
export MCP_SHELL_LOGGING_FORMAT="json"

# Execution
export MCP_SHELL_EXECUTION_LOG_DIRECTORY="/custom/logs"
export MCP_SHELL_EXECUTION_MAX_PARALLEL_JOBS="10"

# Security
export MCP_SHELL_SECURITY_ALLOW_SCRIPT_CREATION="false"

# Explicit config file
export MCP_SHELL_CONFIG="/path/to/config.yaml"
```

Environment variable naming convention:
- Prefix: `MCP_SHELL_`
- Nested structure uses underscores: `SECTION_FIELD`
- Arrays use comma-separated values

## Validating Configuration

Use the `validate` command to check your configuration:

```bash
# Validate default configuration
mcp-shell validate

# Validate specific config file
mcp-shell validate --config /path/to/config.yaml

# Validate with verbose output
mcp-shell validate --verbose
```

The validator checks:
- Configuration file syntax and structure
- Script file existence and permissions
- Parameter definitions
- Alias command syntax
- Log directory accessibility
- MCP server configuration

## Listing Configured Items

View all configured scripts and aliases:

```bash
# List in table format (default)
mcp-shell list

# List in JSON format
mcp-shell list --format json

# List only scripts
mcp-shell list --scripts

# List only aliases
mcp-shell list --aliases

# Show detailed parameter information
mcp-shell list --details
```

## Examples

### Minimal Configuration

```yaml
mcp:
  name: "mcp-shell"
  version: "1.0.0"

execution:
  log_directory: "./logs"

security:
  allowed_interpreters: ["bash"]

logging:
  level: "info"
  format: "text"
```

### Development Configuration

```yaml
mcp:
  name: "mcp-shell-dev"
  version: "dev"

execution:
  max_parallel_jobs: 10
  default_timeout: "30m"
  log_directory: "./dev-logs"
  allow_background: true

security:
  allow_script_creation: true
  allowed_interpreters: ["bash", "python3", "node"]
  script_creation_path: "./user-scripts"

logging:
  level: "debug"
  format: "text"
```

### Production Configuration

```yaml
mcp:
  name: "mcp-shell-prod"
  version: "1.0.0"

scripts:
  - name: "health-check"
    description: "System health check"
    path: "/opt/monitoring/health.sh"
    interpreter: "bash"

execution:
  max_parallel_jobs: 5
  default_timeout: "1h"
  log_directory: "/var/log/mcp-shell"
  allow_background: false

security:
  allow_script_creation: false
  allowed_interpreters: ["bash"]

logging:
  level: "info"
  format: "json"
```

## Next Steps

- [Getting Started](/docs/getting-started/): Learn basic usage
- [Development](/docs/development/): Contribute to MCP Shell
