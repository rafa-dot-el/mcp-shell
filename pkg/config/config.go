/*
MCP Shell Server for serving shell AI models
Copyright (C) 2025 Rafael

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package config provides configuration management for MCP Shell.
//
// This package handles loading configuration from multiple sources:
// - Command line flags (highest priority)
// - Environment variables
// - Configuration files
// - Default values (lowest priority)
//
// Configuration files are searched in the following order:
// 1. Explicit config file via --config flag
// 2. ~/.config/mcp-shell/config.yaml
// 3. ./mcp-shell.yaml (project directory)
// 4. /etc/mcp-shell/config.yaml
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the application configuration
type Config struct {
	// MCP server configuration
	MCP MCPConfig `mapstructure:"mcp" yaml:"mcp"`

	// Aliases are simple one-liner commands with no arguments
	Aliases []Alias `mapstructure:"aliases" yaml:"aliases"`

	// Scripts are executable files with optional parameters
	Scripts []Script `mapstructure:"scripts" yaml:"scripts"`

	// ScriptFolders define directories containing discoverable scripts
	ScriptFolders []ScriptFolder `mapstructure:"script_folders" yaml:"script_folders"`

	// Execution settings control job execution behavior
	Execution ExecutionConfig `mapstructure:"execution" yaml:"execution"`

	// Security settings control script creation and execution permissions
	Security SecurityConfig `mapstructure:"security" yaml:"security"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`
}

// MCPConfig configures the MCP server
type MCPConfig struct {
	// Name of the MCP server
	Name string `mapstructure:"name" yaml:"name"`

	// Version of the MCP server
	Version string `mapstructure:"version" yaml:"version"`

	// Transport protocol (stdio, http, etc.)
	Transport string `mapstructure:"transport" yaml:"transport"`
}

// Alias represents a simple command with no arguments
type Alias struct {
	// Name is the alias identifier
	Name string `mapstructure:"name" yaml:"name"`

	// Description explains what the alias does
	Description string `mapstructure:"description" yaml:"description"`

	// Command is the shell one-liner to execute
	Command string `mapstructure:"command" yaml:"command"`
}

// Script represents an executable script with parameters
type Script struct {
	// Name is the script identifier
	Name string `mapstructure:"name" yaml:"name"`

	// Description explains what the script does
	Description string `mapstructure:"description" yaml:"description"`

	// Path is the file path to the script
	Path string `mapstructure:"path" yaml:"path"`

	// Interpreter to use (bash, python3, perl, etc.)
	Interpreter string `mapstructure:"interpreter" yaml:"interpreter"`

	// Parameters defines the script's input parameters
	Parameters map[string]Parameter `mapstructure:"parameters" yaml:"parameters"`
}

// Parameter defines a script parameter
type Parameter struct {
	// Description explains the parameter's purpose
	Description string `mapstructure:"description" yaml:"description"`

	// Required indicates if the parameter must be provided
	Required bool `mapstructure:"required" yaml:"required"`

	// Default value if not provided
	Default string `mapstructure:"default" yaml:"default"`

	// ValidValues restricts the allowed values
	ValidValues []string `mapstructure:"valid_values" yaml:"valid_values"`

	// Setter defines how the parameter is passed to the script
	// Examples: "--flag {}", "-f {}", "{}"
	Setter string `mapstructure:"setter" yaml:"setter"`
}

// ScriptFolder defines a directory containing discoverable scripts
type ScriptFolder struct {
	// Name is the folder identifier
	Name string `mapstructure:"name" yaml:"name"`

	// Description explains the folder's purpose
	Description string `mapstructure:"description" yaml:"description"`

	// Path is the glob pattern for discovering scripts
	Path string `mapstructure:"path" yaml:"path"`

	// DefaultInterpreter to use for scripts in this folder
	DefaultInterpreter string `mapstructure:"default_interpreter" yaml:"default_interpreter"`
}

// ExecutionConfig controls job execution behavior
type ExecutionConfig struct {
	// MaxParallelJobs limits concurrent job execution
	MaxParallelJobs int `mapstructure:"max_parallel_jobs" yaml:"max_parallel_jobs"`

	// DefaultTimeout in seconds for script execution
	DefaultTimeout int `mapstructure:"default_timeout" yaml:"default_timeout"`

	// LogDirectory stores job execution logs
	LogDirectory string `mapstructure:"log_directory" yaml:"log_directory"`

	// AllowBackground enables background job execution
	AllowBackground bool `mapstructure:"allow_background" yaml:"allow_background"`
}

// SecurityConfig controls security settings
type SecurityConfig struct {
	// AllowScriptCreation permits the model to create new scripts
	AllowScriptCreation bool `mapstructure:"allow_script_creation" yaml:"allow_script_creation"`

	// AllowedInterpreters restricts which interpreters can be used
	AllowedInterpreters []string `mapstructure:"allowed_interpreters" yaml:"allowed_interpreters"`

	// ScriptCreationPath is where the model can create scripts
	ScriptCreationPath string `mapstructure:"script_creation_path" yaml:"script_creation_path"`
}

// LoggingConfig controls logging behavior
type LoggingConfig struct {
	// Level sets the logging level (trace, debug, info, warn, error, fatal, panic)
	Level string `mapstructure:"level" yaml:"level"`

	// Format sets the log format (text, json)
	Format string `mapstructure:"format" yaml:"format"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MCP: MCPConfig{
			Name:      "mcp-shell",
			Version:   "1.0.0",
			Transport: "stdio",
		},
		Aliases:       []Alias{},
		Scripts:       []Script{},
		ScriptFolders: []ScriptFolder{},
		Execution: ExecutionConfig{
			MaxParallelJobs: 5,
			DefaultTimeout:  300,
			LogDirectory:    "/var/log/mcp-shell",
			AllowBackground: true,
		},
		Security: SecurityConfig{
			AllowScriptCreation: false,
			AllowedInterpreters: []string{"bash", "python3", "perl"},
			ScriptCreationPath:  "user-scripts/",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// GetConfigDir returns the default configuration directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "mcp-shell")
	return configDir, nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0750)
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' (value: %v): %s", e.Field, e.Value, e.Message)
}

// Validate validates the configuration and returns any validation errors
func (c *Config) Validate() []error {
	var errors []error

	// Validate MCP configuration
	errors = append(errors, c.validateMCP()...)

	// Validate aliases
	errors = append(errors, c.validateAliases()...)

	// Validate scripts
	errors = append(errors, c.validateScripts()...)

	// Validate script folders
	errors = append(errors, c.validateScriptFolders()...)

	// Validate execution configuration
	errors = append(errors, c.validateExecution()...)

	// Validate security configuration
	errors = append(errors, c.validateSecurity()...)

	// Validate logging configuration
	errors = append(errors, c.validateLogging()...)

	return errors
}

// validateMCP validates MCP server configuration
func (c *Config) validateMCP() []error {
	var errors []error

	if c.MCP.Name == "" {
		errors = append(errors, &ValidationError{
			Field:   "mcp.name",
			Value:   c.MCP.Name,
			Message: "MCP server name cannot be empty",
		})
	}

	if c.MCP.Version == "" {
		errors = append(errors, &ValidationError{
			Field:   "mcp.version",
			Value:   c.MCP.Version,
			Message: "MCP server version cannot be empty",
		})
	}

	validTransports := []string{"stdio", "http", "https"}
	if !contains(validTransports, c.MCP.Transport) {
		errors = append(errors, &ValidationError{
			Field:   "mcp.transport",
			Value:   c.MCP.Transport,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validTransports, ", ")),
		})
	}

	return errors
}

// validateAliases validates alias configuration
func (c *Config) validateAliases() []error {
	var errors []error
	names := make(map[string]bool)

	for i, alias := range c.Aliases {
		if alias.Name == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("aliases[%d].name", i),
				Value:   alias.Name,
				Message: "alias name cannot be empty",
			})
		}

		if names[alias.Name] {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("aliases[%d].name", i),
				Value:   alias.Name,
				Message: "duplicate alias name",
			})
		}
		names[alias.Name] = true

		if alias.Command == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("aliases[%d].command", i),
				Value:   alias.Command,
				Message: "alias command cannot be empty",
			})
		}
	}

	return errors
}

// validateScripts validates script configuration
func (c *Config) validateScripts() []error {
	var errors []error
	names := make(map[string]bool)

	for i, script := range c.Scripts {
		if script.Name == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("scripts[%d].name", i),
				Value:   script.Name,
				Message: "script name cannot be empty",
			})
		}

		if names[script.Name] {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("scripts[%d].name", i),
				Value:   script.Name,
				Message: "duplicate script name",
			})
		}
		names[script.Name] = true

		if script.Path == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("scripts[%d].path", i),
				Value:   script.Path,
				Message: "script path cannot be empty",
			})
		}

		if script.Interpreter != "" && !contains(c.Security.AllowedInterpreters, script.Interpreter) {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("scripts[%d].interpreter", i),
				Value:   script.Interpreter,
				Message: fmt.Sprintf("interpreter not in allowed list: %s", strings.Join(c.Security.AllowedInterpreters, ", ")),
			})
		}

		// Validate parameters
		for paramName, param := range script.Parameters {
			if param.Setter == "" {
				errors = append(errors, &ValidationError{
					Field:   fmt.Sprintf("scripts[%d].parameters[%s].setter", i, paramName),
					Value:   param.Setter,
					Message: "parameter setter cannot be empty",
				})
			}

			if !strings.Contains(param.Setter, "{}") {
				errors = append(errors, &ValidationError{
					Field:   fmt.Sprintf("scripts[%d].parameters[%s].setter", i, paramName),
					Value:   param.Setter,
					Message: "parameter setter must contain {} placeholder",
				})
			}

			if param.Required && param.Default != "" {
				errors = append(errors, &ValidationError{
					Field:   fmt.Sprintf("scripts[%d].parameters[%s]", i, paramName),
					Value:   fmt.Sprintf("required=%t, default=%s", param.Required, param.Default),
					Message: "required parameter cannot have a default value",
				})
			}
		}
	}

	return errors
}

// validateScriptFolders validates script folder configuration
func (c *Config) validateScriptFolders() []error {
	var errors []error
	names := make(map[string]bool)

	for i, folder := range c.ScriptFolders {
		if folder.Name == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("script_folders[%d].name", i),
				Value:   folder.Name,
				Message: "folder name cannot be empty",
			})
		}

		if names[folder.Name] {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("script_folders[%d].name", i),
				Value:   folder.Name,
				Message: "duplicate folder name",
			})
		}
		names[folder.Name] = true

		if folder.Path == "" {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("script_folders[%d].path", i),
				Value:   folder.Path,
				Message: "folder path cannot be empty",
			})
		}

		if folder.DefaultInterpreter != "" && !contains(c.Security.AllowedInterpreters, folder.DefaultInterpreter) {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("script_folders[%d].default_interpreter", i),
				Value:   folder.DefaultInterpreter,
				Message: fmt.Sprintf("interpreter not in allowed list: %s", strings.Join(c.Security.AllowedInterpreters, ", ")),
			})
		}
	}

	return errors
}

// validateExecution validates execution configuration
func (c *Config) validateExecution() []error {
	var errors []error

	if c.Execution.MaxParallelJobs < 1 {
		errors = append(errors, &ValidationError{
			Field:   "execution.max_parallel_jobs",
			Value:   c.Execution.MaxParallelJobs,
			Message: "must be at least 1",
		})
	}

	if c.Execution.DefaultTimeout < 1 {
		errors = append(errors, &ValidationError{
			Field:   "execution.default_timeout",
			Value:   c.Execution.DefaultTimeout,
			Message: "must be at least 1 second",
		})
	}

	if c.Execution.LogDirectory == "" {
		errors = append(errors, &ValidationError{
			Field:   "execution.log_directory",
			Value:   c.Execution.LogDirectory,
			Message: "log directory cannot be empty",
		})
	}

	return errors
}

// validateSecurity validates security configuration
func (c *Config) validateSecurity() []error {
	var errors []error

	if len(c.Security.AllowedInterpreters) == 0 {
		errors = append(errors, &ValidationError{
			Field:   "security.allowed_interpreters",
			Value:   c.Security.AllowedInterpreters,
			Message: "must specify at least one allowed interpreter",
		})
	}

	if c.Security.AllowScriptCreation && c.Security.ScriptCreationPath == "" {
		errors = append(errors, &ValidationError{
			Field:   "security.script_creation_path",
			Value:   c.Security.ScriptCreationPath,
			Message: "script creation path must be specified when script creation is allowed",
		})
	}

	return errors
}

// validateLogging validates logging configuration
func (c *Config) validateLogging() []error {
	var errors []error

	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, strings.ToLower(c.Logging.Level)) {
		errors = append(errors, &ValidationError{
			Field:   "logging.level",
			Value:   c.Logging.Level,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")),
		})
	}

	validFormats := []string{"text", "json"}
	if !contains(validFormats, strings.ToLower(c.Logging.Format)) {
		errors = append(errors, &ValidationError{
			Field:   "logging.format",
			Value:   c.Logging.Format,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")),
		})
	}

	return errors
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(item) {
			return true
		}
	}
	return false
}

// IsValid checks if the configuration is valid
func (c *Config) IsValid() bool {
	return len(c.Validate()) == 0
}
