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
