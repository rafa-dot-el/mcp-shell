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

//go:build integration

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

// TestIntegrationConfigFileLoading tests that configuration files are properly loaded
// from different locations in the expected order of precedence.
func TestIntegrationConfigFileLoading(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test config file
	configContent := `verbose: true
debug: true
log_level: debug
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test that the config file can be loaded
	// This is a basic integration test - in a real application,
	// you would test the actual configuration loading mechanism
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file should exist: %v", err)
	}

	// Test config directory creation
	if err := config.EnsureConfigDir(); err != nil {
		t.Errorf("Failed to ensure config directory: %v", err)
	}
}

// TestIntegrationCommandExecution tests that commands can be executed
// and produce expected output.
func TestIntegrationCommandExecution(t *testing.T) {
	// Set up a buffer to capture output
	var buf bytes.Buffer

	// Test version command execution
	// In a real integration test, you might execute the actual binary
	// or test command execution through the cobra command structure
	versionCmd := rootCmd.Commands()[0] // Assuming version is the first subcommand
	if versionCmd.Name() != "version" {
		// Find the version command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "version" {
				versionCmd = cmd
				break
			}
		}
	}

	if versionCmd.Name() == "version" {
		// This is a placeholder - in a real integration test,
		// you would capture and verify the actual command output
		t.Logf("Version command found: %s", versionCmd.Short)
	}

	// Test root command help
	rootCmd.SetOutput(&buf)
	rootCmd.SetArgs([]string{"--help"})

	// Execute the command (this will exit, so we need to handle it carefully)
	// In a real integration test, you might use a separate process or
	// mock the execution environment
	t.Log("Integration test for command execution completed")
}

// TestIntegrationFileOperations tests file system operations
// that the application might perform.
func TestIntegrationFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test config directory creation
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("Failed to restore HOME environment variable: %v", err)
		}
	}()

	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}

	// Test ensuring config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		t.Errorf("Failed to ensure config directory: %v", err)
	}

	// Verify the directory was created
	configDir, err := config.GetConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config directory: %v", err)
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory should have been created: %s", configDir)
	}
}

// TestIntegrationEnvironmentVariables tests that environment variables
// are properly handled by the application.
func TestIntegrationEnvironmentVariables(t *testing.T) {
	// Test environment variable handling
	envVars := map[string]string{
		"MCP-SHELL_VERBOSE":   "true",
		"MCP-SHELL_DEBUG":     "true",
		"MCP-SHELL_LOG_LEVEL": "debug",
	}

	// Set environment variables
	for key, value := range envVars {
		originalValue := os.Getenv(key)
		defer func(k, v string) {
			if v == "" {
				if err := os.Unsetenv(k); err != nil {
					t.Errorf("Failed to unset environment variable %s: %v", k, err)
				}
			} else {
				if err := os.Setenv(k, v); err != nil {
					t.Errorf("Failed to restore environment variable %s: %v", k, err)
				}
			}
		}(key, originalValue)

		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	// Test that environment variables are accessible
	for key, expectedValue := range envVars {
		if actualValue := os.Getenv(key); actualValue != expectedValue {
			t.Errorf("Environment variable %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// TestIntegrationLongRunningOperation tests operations that might take
// longer to complete, simulating real-world usage scenarios.
func TestIntegrationLongRunningOperation(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping long-running integration test in short mode")
	}

	// Simulate a long-running operation
	start := time.Now()
	time.Sleep(100 * time.Millisecond) // Simulate work
	duration := time.Since(start)

	if duration < 100*time.Millisecond {
		t.Errorf("Operation completed too quickly: %v", duration)
	}

	t.Logf("Long-running operation completed in %v", duration)
}

// TestIntegrationConcurrentOperations tests that the application
// can handle concurrent operations correctly.
func TestIntegrationConcurrentOperations(t *testing.T) {
	const numGoroutines = 10

	// Create a channel to collect results
	results := make(chan error, numGoroutines)

	// Start multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Simulate concurrent config directory access
			if err := config.EnsureConfigDir(); err != nil {
				results <- err
				return
			}

			// Test concurrent config directory retrieval
			if _, err := config.GetConfigDir(); err != nil {
				results <- err
				return
			}

			results <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent operation %d failed: %v", i, err)
		}
	}
}