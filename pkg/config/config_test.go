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
	"os"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Verbose {
		t.Error("Expected Verbose to be false")
	}

	if config.Debug {
		t.Error("Expected Debug to be false")
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got '%s'", config.LogLevel)
	}
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "default config",
			config: &Config{
				Verbose:  false,
				Debug:    false,
				LogLevel: "info",
			},
		},
		{
			name: "verbose only",
			config: &Config{
				Verbose:  true,
				Debug:    false,
				LogLevel: "debug",
			},
		},
		{
			name: "debug only",
			config: &Config{
				Verbose:  false,
				Debug:    true,
				LogLevel: "trace",
			},
		},
		{
			name: "different log levels",
			config: &Config{
				Verbose:  false,
				Debug:    false,
				LogLevel: "error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			if len(errors) != 0 {
				t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
			}

			if !tt.config.IsValid() {
				t.Error("Expected config to be valid")
			}
		})
	}
}

func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{"invalid level", "invalid"},
		{"empty level", ""},
		{"numeric level", "123"},
		{"mixed case invalid", "Invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Verbose:  false,
				Debug:    false,
				LogLevel: tt.logLevel,
			}

			errors := config.Validate()
			if len(errors) != 1 {
				t.Errorf("Expected 1 validation error, got %d: %v", len(errors), errors)
			}

			if !strings.Contains(errors[0].Error(), "log_level") {
				t.Errorf("Expected error to mention 'log_level', got: %s", errors[0].Error())
			}

			if config.IsValid() {
				t.Error("Expected config to be invalid")
			}
		})
	}
}

func TestConfig_Validate_ConflictingOptions(t *testing.T) {
	config := &Config{
		Verbose:  true,
		Debug:    true,
		LogLevel: "info",
	}

	errors := config.Validate()
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d: %v", len(errors), errors)
	}

	if !strings.Contains(errors[0].Error(), "verbose/debug") {
		t.Errorf("Expected error to mention 'verbose/debug', got: %s", errors[0].Error())
	}

	if config.IsValid() {
		t.Error("Expected config to be invalid")
	}
}

func TestConfig_NormalizeLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase", "INFO", "info"},
		{"mixed case", "DeBuG", "debug"},
		{"already lowercase", "error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{LogLevel: tt.input}
			config.NormalizeLogLevel()

			if config.LogLevel != tt.expected {
				t.Errorf("Expected LogLevel to be '%s', got '%s'", tt.expected, config.LogLevel)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}

	expected := "validation error for field 'test_field' (value: test_value): test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	if configDir == "" {
		t.Error("Expected non-empty config directory path")
	}

	if !strings.Contains(configDir, ".config") {
		t.Errorf("Expected config directory to contain '.config', got '%s'", configDir)
	}

	if !strings.Contains(configDir, "mcp-shell") {
		t.Errorf("Expected config directory to contain 'mcp-shell', got '%s'", configDir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed: %v", err)
	}

	// Check that the directory was created
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected config directory to be created at '%s'", configDir)
	}
}

// Integration test with realistic scenarios
func TestConfig_IntegrationScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		config      *Config
		shouldBeValid bool
		expectedErrors int
	}{
		{
			name: "production config",
			config: &Config{
				Verbose:  false,
				Debug:    false,
				LogLevel: "warn",
			},
			shouldBeValid: true,
			expectedErrors: 0,
		},
		{
			name: "development config",
			config: &Config{
				Verbose:  true,
				Debug:    false,
				LogLevel: "debug",
			},
			shouldBeValid: true,
			expectedErrors: 0,
		},
		{
			name: "invalid production config",
			config: &Config{
				Verbose:  false,
				Debug:    false,
				LogLevel: "PRODUCTION",
			},
			shouldBeValid: false,
			expectedErrors: 1,
		},
		{
			name: "conflicting development config",
			config: &Config{
				Verbose:  true,
				Debug:    true,
				LogLevel: "trace",
			},
			shouldBeValid: false,
			expectedErrors: 1,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			errors := scenario.config.Validate()

			if len(errors) != scenario.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", scenario.expectedErrors, len(errors), errors)
			}

			if scenario.config.IsValid() != scenario.shouldBeValid {
				t.Errorf("Expected IsValid() to be %t, got %t", scenario.shouldBeValid, scenario.config.IsValid())
			}

			// Test normalization
			originalLevel := scenario.config.LogLevel
			scenario.config.NormalizeLogLevel()
			if scenario.config.LogLevel != strings.ToLower(originalLevel) {
				t.Errorf("Expected normalized log level '%s', got '%s'", strings.ToLower(originalLevel), scenario.config.LogLevel)
			}
		})
	}
}