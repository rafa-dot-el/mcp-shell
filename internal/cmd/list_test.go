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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func setupTestListConfig(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
echo "Test script"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	configContent := fmt.Sprintf(`mcp:
  name: "MCP Shell Test"
  version: "1.0.0-test"
  transport: "stdio"

execution:
  log_directory: "%s"
  max_parallel_jobs: 5
  default_timeout: "1h"

scripts:
  - name: "test-script"
    description: "Test script for listing"
    path: "%s"
    interpreter: "bash"
    parameters:
      message:
        description: "Test message"
        required: false
        default: "default"
        setter: "--message {}"

aliases:
  - name: "test-alias"
    description: "Test alias for listing"
    command: "echo 'Test alias'"

security:
  allowed_interpreters:
    - "bash"

logging:
  format: "text"
  level: "info"
`, filepath.Join(tmpDir, "logs"), scriptPath)

	// Write config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Reset viper
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	return tmpDir
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

func TestListCommand_TableFormat(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	if !strings.Contains(output, "test-script") {
		t.Error("Expected output to contain 'test-script'")
	}

	if !strings.Contains(output, "test-alias") {
		t.Error("Expected output to contain 'test-alias'")
	}

	if !strings.Contains(output, "SCRIPTS") {
		t.Error("Expected output to contain 'SCRIPTS' header")
	}

	if !strings.Contains(output, "ALIASES") {
		t.Error("Expected output to contain 'ALIASES' header")
	}
}

func TestListCommand_JSONFormat(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{"--format", "json"})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	scripts, ok := result["scripts"].([]interface{})
	if !ok {
		t.Fatal("Expected 'scripts' field in JSON output")
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}

	aliases, ok := result["aliases"].([]interface{})
	if !ok {
		t.Fatal("Expected 'aliases' field in JSON output")
	}

	if len(aliases) != 1 {
		t.Errorf("Expected 1 alias, got %d", len(aliases))
	}
}

func TestListCommand_SimpleFormat(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{"--format", "simple"})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	if !strings.Contains(output, "test-script") {
		t.Error("Expected output to contain 'test-script'")
	}

	if !strings.Contains(output, "test-alias") {
		t.Error("Expected output to contain 'test-alias'")
	}
}

func TestListCommand_ScriptsOnly(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{"--scripts"})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	if !strings.Contains(output, "test-script") {
		t.Error("Expected output to contain 'test-script'")
	}

	if strings.Contains(output, "test-alias") {
		t.Error("Expected output to NOT contain 'test-alias' with --scripts flag")
	}
}

func TestListCommand_AliasesOnly(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{"--aliases"})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	if strings.Contains(output, "test-script") {
		t.Error("Expected output to NOT contain 'test-script' with --aliases flag")
	}

	if !strings.Contains(output, "test-alias") {
		t.Error("Expected output to contain 'test-alias'")
	}
}

func TestListCommand_WithDetails(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	defer os.RemoveAll(tmpDir)

	output := captureOutput(func() {
		listCmd.SetArgs([]string{"--details"})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	if !strings.Contains(output, "Parameters:") {
		t.Error("Expected output to contain 'Parameters:' with --details flag")
	}

	if !strings.Contains(output, "message") {
		t.Error("Expected output to contain parameter name 'message'")
	}
}

func TestListCommand_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := fmt.Sprintf(`mcp:
  name: "MCP Shell Test"
  version: "1.0.0-test"
  transport: "stdio"

execution:
  log_directory: "%s"
  max_parallel_jobs: 5
  default_timeout: "1h"

scripts: []
aliases: []

security:
  allowed_interpreters:
    - "bash"

logging:
  format: "text"
  level: "info"
`, filepath.Join(tmpDir, "logs"))

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	output := captureOutput(func() {
		listCmd.SetArgs([]string{})
		if err := listCmd.Execute(); err != nil {
			t.Errorf("listCmd.Execute() failed: %v", err)
		}
	})

	// Should still work, just with empty output
	if output == "" {
		t.Error("Expected some output even with empty config")
	}
}
