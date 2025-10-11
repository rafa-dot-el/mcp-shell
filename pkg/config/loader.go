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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from files and environment variables
// Configuration sources are merged in this order (highest to lowest priority):
// 1. Explicit config file via configPath parameter
// 2. Environment variables (MCP_SHELL_*)
// 3. User config (~/.config/mcp-shell/config.yaml)
// 4. Project config (./mcp-shell.yaml)
// 5. System config (/etc/mcp-shell/config.yaml)
// 6. Default values
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set up Viper configuration
	v.SetConfigName("mcp-shell")
	v.SetConfigType("yaml")

	// Configure environment variable binding
	v.SetEnvPrefix("MCP_SHELL")
	v.AutomaticEnv()

	// Add configuration search paths
	if configPath != "" {
		// Explicit config file
		v.SetConfigFile(configPath)
	} else {
		// Add search paths in priority order
		if configDir, err := GetConfigDir(); err == nil {
			v.AddConfigPath(configDir)
		}
		v.AddConfigPath(".")              // Current directory
		v.AddConfigPath("/etc/mcp-shell") // System config
	}

	// Set default values
	setDefaults(v)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is acceptable - we'll use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into Config struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if errors := config.Validate(); len(errors) > 0 {
		return nil, fmt.Errorf("configuration validation failed: %v", errors)
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, configPath string) error {
	// Validate before saving
	if errors := config.Validate(); len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Marshal config into Viper
	if err := v.MergeConfigMap(configToMap(config)); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	// Write to file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// setDefaults configures default values in Viper
func setDefaults(v *viper.Viper) {
	defaults := DefaultConfig()

	// MCP defaults
	v.SetDefault("mcp.name", defaults.MCP.Name)
	v.SetDefault("mcp.version", defaults.MCP.Version)
	v.SetDefault("mcp.transport", defaults.MCP.Transport)

	// Execution defaults
	v.SetDefault("execution.max_parallel_jobs", defaults.Execution.MaxParallelJobs)
	v.SetDefault("execution.default_timeout", defaults.Execution.DefaultTimeout)
	v.SetDefault("execution.log_directory", defaults.Execution.LogDirectory)
	v.SetDefault("execution.allow_background", defaults.Execution.AllowBackground)

	// Security defaults
	v.SetDefault("security.allow_script_creation", defaults.Security.AllowScriptCreation)
	v.SetDefault("security.allowed_interpreters", defaults.Security.AllowedInterpreters)
	v.SetDefault("security.script_creation_path", defaults.Security.ScriptCreationPath)

	// Logging defaults
	v.SetDefault("logging.level", defaults.Logging.Level)
	v.SetDefault("logging.format", defaults.Logging.Format)
}

// configToMap converts Config struct to map for Viper serialization
func configToMap(config *Config) map[string]interface{} {
	return map[string]interface{}{
		"mcp": map[string]interface{}{
			"name":      config.MCP.Name,
			"version":   config.MCP.Version,
			"transport": config.MCP.Transport,
		},
		"aliases":        config.Aliases,
		"scripts":        config.Scripts,
		"script_folders": config.ScriptFolders,
		"execution": map[string]interface{}{
			"max_parallel_jobs": config.Execution.MaxParallelJobs,
			"default_timeout":   config.Execution.DefaultTimeout,
			"log_directory":     config.Execution.LogDirectory,
			"allow_background":  config.Execution.AllowBackground,
		},
		"security": map[string]interface{}{
			"allow_script_creation": config.Security.AllowScriptCreation,
			"allowed_interpreters":  config.Security.AllowedInterpreters,
			"script_creation_path":  config.Security.ScriptCreationPath,
		},
		"logging": map[string]interface{}{
			"level":  config.Logging.Level,
			"format": config.Logging.Format,
		},
	}
}

// GetDefaultConfigPath returns the default user configuration file path
func GetDefaultConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}
