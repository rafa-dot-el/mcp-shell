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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupTestValidateConfig(t *testing.T, withScript bool) string {
	tmpDir := t.TempDir()

	var configContent string

	if withScript {
		// Create test script
		scriptPath := filepath.Join(tmpDir, "test.sh")
		scriptContent := `#!/bin/bash
echo "Test script"
exit 0
`
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		configContent = fmt.Sprintf(`mcp:
  name: "MCP Shell Test"
  version: "1.0.0-test"
  transport: "stdio"

execution:
  log_directory: "%s"
  max_parallel_jobs: 5
  default_timeout: 3600

scripts:
  - name: "test-script"
    description: "Test script"
    path: "%s"
    interpreter: "bash"
    parameters: {}

aliases:
  - name: "test-alias"
    description: "Test alias"
    command: "echo test"

security:
  allowed_interpreters:
    - "bash"

logging:
  format: "text"
  level: "info"
`, filepath.Join(tmpDir, "logs"), scriptPath)
	} else {
		configContent = fmt.Sprintf(`mcp:
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
	}

	// Write config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Reset viper for clean state
	viper.Reset()

	return tmpDir
}

func TestValidateCommand(t *testing.T) {
	tmpDir := setupTestValidateConfig(t, true)
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Capture output
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Run validate command
	rootCmd.SetArgs([]string{"validate", "--config", configPath})
	err := rootCmd.Execute()

	// Restore output
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("rootCmd.Execute() failed: %v", err)
	}

	// Read captured output
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestValidateCommand_EmptyConfig(t *testing.T) {
	tmpDir := setupTestValidateConfig(t, false)
	_ = tmpDir // TempDir automatically cleaned up by testing framework

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Run validate command
	rootCmd.SetArgs([]string{"validate", "--config", configPath})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("rootCmd.Execute() should not fail with empty config: %v", err)
	}
}

func TestValidateCommand_MissingScript(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "nonexistent.sh")
	configContent := fmt.Sprintf(`mcp:
  name: "MCP Shell Test"
  version: "1.0.0-test"
  transport: "stdio"

execution:
  log_directory: "%s"
  max_parallel_jobs: 5
  default_timeout: 3600

scripts:
  - name: "missing-script"
    description: "Missing script"
    path: "%s"
    interpreter: "bash"
    parameters: {}

aliases: []

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

	// Reset viper for clean state
	viper.Reset()

	// Run validate command - should fail
	rootCmd.SetArgs([]string{"validate", "--config", configPath})
	err := rootCmd.Execute()

	if err == nil {
		t.Error("Expected validation to fail with missing script")
	}
}

func TestValidateCommand_NonExecutableScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-executable script
	scriptPath := filepath.Join(tmpDir, "nonexec.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0644); err != nil {
		t.Fatalf("Failed to create script: %v", err)
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
  - name: "nonexec-script"
    description: "Non-executable script"
    path: "%s"
    interpreter: "bash"
    parameters: {}

aliases: []

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

	// Reset viper for clean state
	viper.Reset()

	// Run validate command - should fail
	rootCmd.SetArgs([]string{"validate", "--config", configPath})
	err := rootCmd.Execute()

	if err == nil {
		t.Error("Expected validation to fail with non-executable script")
	}
}

func TestValidateCommand_LogDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	logDir := filepath.Join(tmpDir, "newlogs")
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
`, logDir)

	// Write config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Reset viper for clean state
	viper.Reset()

	// Run validate command
	rootCmd.SetArgs([]string{"validate", "--config", configPath})
	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("rootCmd.Execute() failed: %v", err)
	}

	// Check if log directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Expected log directory to be created")
	}
}
