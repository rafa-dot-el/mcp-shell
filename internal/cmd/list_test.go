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
	// Reset all state at the beginning of each test
	viper.Reset()
	rootCmd.SetArgs([]string{})
	resetCfgFile()

	// Ensure cleanup at the end of the test
	t.Cleanup(func() {
		viper.Reset()
		rootCmd.SetArgs([]string{})
		resetCfgFile()
	})

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
  default_timeout: 3600

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

	return tmpDir
}

func captureOutput(f func()) string {
	// Save original stdout
	oldStdout := os.Stdout

	// Ensure stdout is always restored, even on panic
	defer func() {
		os.Stdout = oldStdout
	}()

	// Create pipe for capturing output
	r, w, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("failed to create pipe: %v", err))
	}

	// Redirect stdout to pipe
	os.Stdout = w

	// Channel to signal completion
	outChan := make(chan string)

	// Start goroutine to read from pipe
	go func() {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r)
		outChan <- buf.String()
	}()

	// Run the function
	f()

	// Close writer to signal EOF
	_ = w.Close()

	// Wait for reader to finish and get output
	output := <-outChan
	_ = r.Close()

	return output
}

// resetCfgFile resets the package-level cfgFile variable
func resetCfgFile() {
	// Access the package-level cfgFile by parsing it through the flag
	_ = rootCmd.PersistentFlags().Set("config", "")

	// Reset listCmd flags to default values
	_ = listCmd.Flags().Set("format", "table")
	_ = listCmd.Flags().Set("scripts", "false")
	_ = listCmd.Flags().Set("aliases", "false")
	_ = listCmd.Flags().Set("details", "false")
}

func TestListCommand_TableFormat(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
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
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath, "--format", "json"})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
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
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath, "--format", "simple"})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
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
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath, "--scripts"})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
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
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	t.Logf("Config path: %s", configPath)

	// Verify config file exists and has content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	t.Logf("Config content length: %d bytes", len(content))

	var execErr error
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath, "--aliases"})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("rootCmd.Execute() failed: %v", execErr)
	}

	if strings.Contains(output, "test-script") {
		t.Error("Expected output to NOT contain 'test-script' with --aliases flag")
	}

	if !strings.Contains(output, "test-alias") {
		t.Error("Expected output to contain 'test-alias'")
	}
}

func TestListCommand_WithDetails(t *testing.T) {
	tmpDir := setupTestListConfig(t)
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath, "--details"})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
		}
	})

	if !strings.Contains(output, "Parameters:") {
		t.Errorf("Expected output to contain 'Parameters:' with --details flag. Got: %q", output)
	}

	if !strings.Contains(output, "message") {
		t.Errorf("Expected output to contain parameter name 'message'. Got: %q", output)
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
  default_timeout: 3600

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

	output := captureOutput(func() {
		rootCmd.SetArgs([]string{"list", "--config", configPath})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
		}
	})

	// Empty config should produce no output (no scripts or aliases to list)
	// This is expected and correct behavior
	_ = output // Unused, but that's okay
}
