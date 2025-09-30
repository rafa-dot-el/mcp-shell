//go:build integration
// +build integration

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
	"path/filepath"
	"testing"
)

func TestConfigValidation_Integration(t *testing.T) {
	// Test configuration validation in a realistic scenario
	tmpDir := t.TempDir()

	// Create a config file with invalid content
	configPath := filepath.Join(tmpDir, "config.yaml")
	invalidConfig := `
verbose: true
debug: true
log_level: "INVALID_LEVEL"
`

	err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test that validation catches multiple errors
	config := &Config{
		Verbose:  true,
		Debug:    true,
		LogLevel: "INVALID_LEVEL",
	}

	errors := config.Validate()
	if len(errors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d: %v", len(errors), errors)
	}

	// Test that the config is marked as invalid
	if config.IsValid() {
		t.Error("Expected config to be invalid")
	}
}

func TestConfigDirectory_Integration(t *testing.T) {
	// Test the full config directory creation workflow
	tmpDir := t.TempDir()

	// Override HOME for this test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// Ensure the config directory doesn't exist initially
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("Config directory already exists: %s", configDir)
	}

	// Create the config directory
	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed: %v", err)
	}

	// Verify the directory exists and has correct permissions
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	expectedPerm := os.FileMode(0750)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected directory permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}

	// Test that calling EnsureConfigDir again doesn't fail
	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed on second call: %v", err)
	}
}
