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

package script

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

func setupTestExecutor(t *testing.T) (*Executor, string) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	executor := NewExecutor(manager, cfg)
	return executor, tmpDir
}

func TestExecuteAlias(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	// Add test alias
	executor.manager.config.Aliases = []config.Alias{
		{
			Name:        "test-echo",
			Description: "Test echo command",
			Command:     "echo 'hello world'",
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute alias
	req := &ExecutionRequest{
		Name:    "test-echo",
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := executor.ExecuteAlias(ctx, req)

	if err != nil {
		t.Errorf("ExecuteAlias() failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got: %s", result.Stdout)
	}
}

func TestExecuteAlias_NotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	req := &ExecutionRequest{
		Name:    "nonexistent",
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	_, err := executor.ExecuteAlias(ctx, req)

	if err == nil {
		t.Error("Expected error for nonexistent alias, got nil")
	}
}

func TestExecuteScript_Simple(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
echo "test output"
exit 0
`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Add script to manager
	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Description: "Test script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute script
	req := &ExecutionRequest{
		Name:       "test-script",
		Parameters: make(map[string]string),
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	result, err := executor.ExecuteScript(ctx, req)

	if err != nil {
		t.Errorf("ExecuteScript() failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "test output") {
		t.Errorf("Expected output to contain 'test output', got: %s", result.Stdout)
	}
}

func TestExecuteScript_WithParameters(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	// Create test script that uses parameters
	scriptPath := filepath.Join(tmpDir, "param_test.sh")
	scriptContent := `#!/bin/bash
while [[ $# -gt 0 ]]; do
    case $1 in
        --name)
            NAME="$2"
            shift 2
            ;;
        --count)
            COUNT="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done
echo "Name: $NAME, Count: $COUNT"
`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Add script with parameters
	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "param-script",
			Description: "Script with parameters",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters: map[string]config.Parameter{
				"name": {
					Description: "Name parameter",
					Required:    true,
					Setter:      "--name {}",
				},
				"count": {
					Description: "Count parameter",
					Required:    false,
					Default:     "5",
					Setter:      "--count {}",
				},
			},
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute script with parameters
	req := &ExecutionRequest{
		Name: "param-script",
		Parameters: map[string]string{
			"name":  "test",
			"count": "10",
		},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := executor.ExecuteScript(ctx, req)

	if err != nil {
		t.Errorf("ExecuteScript() failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "Name: test") {
		t.Errorf("Expected output to contain 'Name: test', got: %s", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "Count: 10") {
		t.Errorf("Expected output to contain 'Count: 10', got: %s", result.Stdout)
	}
}

func TestExecuteScript_MissingRequiredParameter(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "param-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters: map[string]config.Parameter{
				"required_param": {
					Description: "Required parameter",
					Required:    true,
					Setter:      "--param {}",
				},
			},
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute without required parameter
	req := &ExecutionRequest{
		Name:       "param-script",
		Parameters: make(map[string]string),
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	_, err := executor.ExecuteScript(ctx, req)

	if err == nil {
		t.Error("Expected error for missing required parameter, got nil")
	}
}

func TestExecuteScript_InvalidParameterValue(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "valid-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters: map[string]config.Parameter{
				"level": {
					Description: "Log level",
					Required:    true,
					ValidValues: []string{"debug", "info", "warn", "error"},
					Setter:      "--level {}",
				},
			},
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute with invalid parameter value
	req := &ExecutionRequest{
		Name: "valid-script",
		Parameters: map[string]string{
			"level": "invalid",
		},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	_, err := executor.ExecuteScript(ctx, req)

	if err == nil {
		t.Error("Expected error for invalid parameter value, got nil")
	}
}

func TestExecuteScript_DefaultParameter(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	scriptPath := filepath.Join(tmpDir, "default_test.sh")
	scriptContent := `#!/bin/bash
while [[ $# -gt 0 ]]; do
    case $1 in
        --value)
            VALUE="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done
echo "Value: $VALUE"
`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "default-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters: map[string]config.Parameter{
				"value": {
					Description: "Value with default",
					Required:    false,
					Default:     "default_value",
					Setter:      "--value {}",
				},
			},
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute without providing parameter (should use default)
	req := &ExecutionRequest{
		Name:       "default-script",
		Parameters: make(map[string]string),
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	result, err := executor.ExecuteScript(ctx, req)

	if err != nil {
		t.Errorf("ExecuteScript() failed: %v", err)
	}

	if !strings.Contains(result.Stdout, "Value: default_value") {
		t.Errorf("Expected output to contain default value, got: %s", result.Stdout)
	}
}

func TestExecuteScript_Timeout(t *testing.T) {
	executor, tmpDir := setupTestExecutor(t)

	// Create script that sleeps
	scriptPath := filepath.Join(tmpDir, "sleep.sh")
	scriptContent := `#!/bin/bash
sleep 10
echo "done"
`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	executor.manager.config.Scripts = []config.Script{
		{
			Name:        "sleep-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := executor.manager.Reload(); err != nil {
		t.Fatalf("Failed to reload manager: %v", err)
	}

	// Execute with short timeout
	req := &ExecutionRequest{
		Name:       "sleep-script",
		Parameters: make(map[string]string),
		Timeout:    1 * time.Second,
	}

	ctx := context.Background()
	result, err := executor.ExecuteScript(ctx, req)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result != nil && result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for timeout")
	}
}

func TestExecuteScript_NotFound(t *testing.T) {
	executor, _ := setupTestExecutor(t)

	req := &ExecutionRequest{
		Name:       "nonexistent",
		Parameters: make(map[string]string),
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	_, err := executor.ExecuteScript(ctx, req)

	if err == nil {
		t.Error("Expected error for nonexistent script, got nil")
	}
}
