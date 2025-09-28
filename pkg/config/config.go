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
	// Verbose enables verbose logging
	Verbose bool `mapstructure:"verbose"`

	// Debug enables debug logging
	Debug bool `mapstructure:"debug"`

	// LogLevel sets the logging level
	LogLevel string `mapstructure:"log_level"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Verbose:  false,
		Debug:    false,
		LogLevel: "info",
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

	// Validate log level
	if err := c.validateLogLevel(); err != nil {
		errors = append(errors, err)
	}

	// Validate conflicting options
	if c.Verbose && c.Debug {
		errors = append(errors, &ValidationError{
			Field:   "verbose/debug",
			Value:   fmt.Sprintf("verbose=%t, debug=%t", c.Verbose, c.Debug),
			Message: "verbose and debug cannot both be true",
		})
	}

	return errors
}

// validateLogLevel validates that the log level is valid
func (c *Config) validateLogLevel() error {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}

	level := strings.ToLower(c.LogLevel)
	for _, valid := range validLevels {
		if level == valid {
			return nil
		}
	}

	return &ValidationError{
		Field:   "log_level",
		Value:   c.LogLevel,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")),
	}
}

// NormalizeLogLevel normalizes the log level to lowercase
func (c *Config) NormalizeLogLevel() {
	c.LogLevel = strings.ToLower(c.LogLevel)
}

// IsValid checks if the configuration is valid
func (c *Config) IsValid() bool {
	return len(c.Validate()) == 0
}
