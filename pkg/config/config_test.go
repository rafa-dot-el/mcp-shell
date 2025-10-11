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
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Validate default config is valid
	if !config.IsValid() {
		t.Errorf("Default config is not valid: %v", config.Validate())
	}

	// Check MCP defaults
	if config.MCP.Name != "mcp-shell" {
		t.Errorf("Expected MCP name 'mcp-shell', got '%s'", config.MCP.Name)
	}

	if config.MCP.Transport != "stdio" {
		t.Errorf("Expected MCP transport 'stdio', got '%s'", config.MCP.Transport)
	}

	// Check execution defaults
	if config.Execution.MaxParallelJobs != 5 {
		t.Errorf("Expected MaxParallelJobs 5, got %d", config.Execution.MaxParallelJobs)
	}

	// Check security defaults
	if config.Security.AllowScriptCreation {
		t.Error("Expected AllowScriptCreation to be false by default")
	}

	if len(config.Security.AllowedInterpreters) == 0 {
		t.Error("Expected at least one allowed interpreter by default")
	}

	// Check logging defaults
	if config.Logging.Level != "info" {
		t.Errorf("Expected logging level 'info', got '%s'", config.Logging.Level)
	}
}

func TestValidateMCP(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid MCP config",
			config: Config{
				MCP: MCPConfig{
					Name:      "test-server",
					Version:   "1.0.0",
					Transport: "stdio",
				},
				Execution: ExecutionConfig{
					MaxParallelJobs: 1,
					DefaultTimeout:  1,
					LogDirectory:    "/tmp",
				},
				Security: SecurityConfig{
					AllowedInterpreters: []string{"bash"},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantError: false,
		},
		{
			name: "empty MCP name",
			config: Config{
				MCP: MCPConfig{
					Name:      "",
					Version:   "1.0.0",
					Transport: "stdio",
				},
				Execution: ExecutionConfig{
					MaxParallelJobs: 1,
					DefaultTimeout:  1,
					LogDirectory:    "/tmp",
				},
				Security: SecurityConfig{
					AllowedInterpreters: []string{"bash"},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantError: true,
		},
		{
			name: "invalid transport",
			config: Config{
				MCP: MCPConfig{
					Name:      "test",
					Version:   "1.0.0",
					Transport: "invalid",
				},
				Execution: ExecutionConfig{
					MaxParallelJobs: 1,
					DefaultTimeout:  1,
					LogDirectory:    "/tmp",
				},
				Security: SecurityConfig{
					AllowedInterpreters: []string{"bash"},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateAliases(t *testing.T) {
	tests := []struct {
		name      string
		aliases   []Alias
		wantError bool
	}{
		{
			name: "valid aliases",
			aliases: []Alias{
				{Name: "test1", Description: "Test 1", Command: "echo test1"},
				{Name: "test2", Description: "Test 2", Command: "echo test2"},
			},
			wantError: false,
		},
		{
			name: "empty alias name",
			aliases: []Alias{
				{Name: "", Description: "Test", Command: "echo test"},
			},
			wantError: true,
		},
		{
			name: "duplicate alias name",
			aliases: []Alias{
				{Name: "test", Description: "Test 1", Command: "echo test1"},
				{Name: "test", Description: "Test 2", Command: "echo test2"},
			},
			wantError: true,
		},
		{
			name: "empty command",
			aliases: []Alias{
				{Name: "test", Description: "Test", Command: ""},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Aliases = tt.aliases

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateScripts(t *testing.T) {
	tests := []struct {
		name      string
		scripts   []Script
		wantError bool
	}{
		{
			name: "valid script with parameters",
			scripts: []Script{
				{
					Name:        "test-script",
					Description: "Test script",
					Path:        "/tmp/test.sh",
					Interpreter: "bash",
					Parameters: map[string]Parameter{
						"param1": {
							Description: "Parameter 1",
							Required:    true,
							Setter:      "--param1 {}",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "empty script name",
			scripts: []Script{
				{
					Name:        "",
					Description: "Test",
					Path:        "/tmp/test.sh",
				},
			},
			wantError: true,
		},
		{
			name: "duplicate script name",
			scripts: []Script{
				{Name: "test", Path: "/tmp/test1.sh"},
				{Name: "test", Path: "/tmp/test2.sh"},
			},
			wantError: true,
		},
		{
			name: "invalid interpreter",
			scripts: []Script{
				{
					Name:        "test",
					Path:        "/tmp/test.sh",
					Interpreter: "invalid-interpreter",
				},
			},
			wantError: true,
		},
		{
			name: "parameter without setter",
			scripts: []Script{
				{
					Name: "test",
					Path: "/tmp/test.sh",
					Parameters: map[string]Parameter{
						"param1": {
							Description: "Parameter 1",
							Setter:      "",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "setter without placeholder",
			scripts: []Script{
				{
					Name: "test",
					Path: "/tmp/test.sh",
					Parameters: map[string]Parameter{
						"param1": {
							Description: "Parameter 1",
							Setter:      "--param1",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "required parameter with default",
			scripts: []Script{
				{
					Name: "test",
					Path: "/tmp/test.sh",
					Parameters: map[string]Parameter{
						"param1": {
							Description: "Parameter 1",
							Required:    true,
							Default:     "value",
							Setter:      "--param1 {}",
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Scripts = tt.scripts

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateScriptFolders(t *testing.T) {
	tests := []struct {
		name      string
		folders   []ScriptFolder
		wantError bool
	}{
		{
			name: "valid script folder",
			folders: []ScriptFolder{
				{
					Name:               "test-folder",
					Description:        "Test folder",
					Path:               "/tmp/*.sh",
					DefaultInterpreter: "bash",
				},
			},
			wantError: false,
		},
		{
			name: "empty folder name",
			folders: []ScriptFolder{
				{
					Name:        "",
					Description: "Test",
					Path:        "/tmp/*.sh",
				},
			},
			wantError: true,
		},
		{
			name: "duplicate folder names",
			folders: []ScriptFolder{
				{Name: "test", Path: "/tmp/*.sh"},
				{Name: "test", Path: "/opt/*.sh"},
			},
			wantError: true,
		},
		{
			name: "empty path",
			folders: []ScriptFolder{
				{Name: "test", Path: ""},
			},
			wantError: true,
		},
		{
			name: "disallowed interpreter",
			folders: []ScriptFolder{
				{
					Name:               "test",
					Path:               "/tmp/*.sh",
					DefaultInterpreter: "invalid-interpreter",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.ScriptFolders = tt.folders

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateExecution(t *testing.T) {
	tests := []struct {
		name      string
		execution ExecutionConfig
		wantError bool
	}{
		{
			name: "valid execution config",
			execution: ExecutionConfig{
				MaxParallelJobs: 5,
				DefaultTimeout:  300,
				LogDirectory:    "/var/log/mcp-shell",
				AllowBackground: true,
			},
			wantError: false,
		},
		{
			name: "invalid max parallel jobs (zero)",
			execution: ExecutionConfig{
				MaxParallelJobs: 0,
				DefaultTimeout:  300,
				LogDirectory:    "/var/log/mcp-shell",
			},
			wantError: true,
		},
		{
			name: "invalid max parallel jobs (negative)",
			execution: ExecutionConfig{
				MaxParallelJobs: -1,
				DefaultTimeout:  300,
				LogDirectory:    "/var/log/mcp-shell",
			},
			wantError: true,
		},
		{
			name: "invalid default timeout (zero)",
			execution: ExecutionConfig{
				MaxParallelJobs: 5,
				DefaultTimeout:  0,
				LogDirectory:    "/var/log/mcp-shell",
			},
			wantError: true,
		},
		{
			name: "invalid default timeout (negative)",
			execution: ExecutionConfig{
				MaxParallelJobs: 5,
				DefaultTimeout:  -1,
				LogDirectory:    "/var/log/mcp-shell",
			},
			wantError: true,
		},
		{
			name: "empty log directory",
			execution: ExecutionConfig{
				MaxParallelJobs: 5,
				DefaultTimeout:  300,
				LogDirectory:    "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Execution = tt.execution

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateSecurity(t *testing.T) {
	tests := []struct {
		name      string
		security  SecurityConfig
		wantError bool
	}{
		{
			name: "valid security config",
			security: SecurityConfig{
				AllowScriptCreation: false,
				AllowedInterpreters: []string{"bash", "python3"},
				ScriptCreationPath:  "scripts/",
			},
			wantError: false,
		},
		{
			name: "empty allowed interpreters",
			security: SecurityConfig{
				AllowScriptCreation: false,
				AllowedInterpreters: []string{},
			},
			wantError: true,
		},
		{
			name: "script creation enabled without path",
			security: SecurityConfig{
				AllowScriptCreation: true,
				AllowedInterpreters: []string{"bash"},
				ScriptCreationPath:  "",
			},
			wantError: true,
		},
		{
			name: "script creation enabled with path",
			security: SecurityConfig{
				AllowScriptCreation: true,
				AllowedInterpreters: []string{"bash"},
				ScriptCreationPath:  "/tmp/scripts/",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Security = tt.security

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateLogging(t *testing.T) {
	tests := []struct {
		name      string
		logging   LoggingConfig
		wantError bool
	}{
		{
			name:      "valid text format",
			logging:   LoggingConfig{Level: "info", Format: "text"},
			wantError: false,
		},
		{
			name:      "valid json format",
			logging:   LoggingConfig{Level: "debug", Format: "json"},
			wantError: false,
		},
		{
			name:      "all valid log levels - trace",
			logging:   LoggingConfig{Level: "trace", Format: "text"},
			wantError: false,
		},
		{
			name:      "all valid log levels - warn",
			logging:   LoggingConfig{Level: "warn", Format: "text"},
			wantError: false,
		},
		{
			name:      "all valid log levels - error",
			logging:   LoggingConfig{Level: "error", Format: "text"},
			wantError: false,
		},
		{
			name:      "all valid log levels - fatal",
			logging:   LoggingConfig{Level: "fatal", Format: "text"},
			wantError: false,
		},
		{
			name:      "all valid log levels - panic",
			logging:   LoggingConfig{Level: "panic", Format: "text"},
			wantError: false,
		},
		{
			name:      "case insensitive level",
			logging:   LoggingConfig{Level: "INFO", Format: "text"},
			wantError: false,
		},
		{
			name:      "case insensitive format",
			logging:   LoggingConfig{Level: "info", Format: "JSON"},
			wantError: false,
		},
		{
			name:      "invalid log level",
			logging:   LoggingConfig{Level: "invalid", Format: "text"},
			wantError: true,
		},
		{
			name:      "invalid log format",
			logging:   LoggingConfig{Level: "info", Format: "invalid"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Logging = tt.logging

			errors := config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "basic error",
			err: &ValidationError{
				Field:   "test.field",
				Value:   "value",
				Message: "test message",
			},
			expected: "validation error for field 'test.field' (value: value): test message",
		},
		{
			name: "numeric value",
			err: &ValidationError{
				Field:   "config.timeout",
				Value:   0,
				Message: "must be positive",
			},
			expected: "validation error for field 'config.timeout' (value: 0): must be positive",
		},
		{
			name: "nil value",
			err: &ValidationError{
				Field:   "config.value",
				Value:   nil,
				Message: "cannot be nil",
			},
			expected: "validation error for field 'config.value' (value: <nil>): cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should end with .config/mcp-shell
	expectedSuffix := ".config/mcp-shell"
	if len(dir) < len(expectedSuffix) {
		t.Errorf("GetConfigDir() path too short: %s", dir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// This test might create directories
	err := EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() error = %v", err)
	}

	// Verify the directory path is returned correctly
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if dir == "" {
		t.Error("Config directory is empty after EnsureConfigDir()")
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		cfg   *Config
		valid bool
	}{
		{
			name:  "default config is valid",
			cfg:   DefaultConfig(),
			valid: true,
		},
		{
			name: "config with empty MCP name",
			cfg: &Config{
				MCP: MCPConfig{
					Name:      "",
					Version:   "1.0.0",
					Transport: "stdio",
				},
				Execution: ExecutionConfig{
					MaxParallelJobs: 5,
					DefaultTimeout:  300,
					LogDirectory:    "/tmp",
				},
				Security: SecurityConfig{
					AllowedInterpreters: []string{"bash"},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			valid: false,
		},
		{
			name: "config with multiple validation errors",
			cfg: &Config{
				MCP: MCPConfig{
					Name:      "",
					Version:   "",
					Transport: "invalid",
				},
				Execution: ExecutionConfig{
					MaxParallelJobs: 0,
					DefaultTimeout:  0,
					LogDirectory:    "",
				},
				Security: SecurityConfig{
					AllowedInterpreters: []string{},
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "invalid",
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsValid(); got != tt.valid {
				t.Errorf("IsValid() = %v, want %v, errors: %v", got, tt.valid, tt.cfg.Validate())
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"bash", "python3", "perl"},
			item:  "bash",
			want:  true,
		},
		{
			name:  "item exists case insensitive",
			slice: []string{"bash", "python3", "perl"},
			item:  "BASH",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: []string{"bash", "python3", "perl"},
			item:  "ruby",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "bash",
			want:  false,
		},
		{
			name:  "empty item",
			slice: []string{"bash", "python3"},
			item:  "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.item); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
