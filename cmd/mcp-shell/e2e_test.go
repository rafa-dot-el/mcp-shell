//go:build e2e

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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	// Test timeout for e2e operations
	testTimeout = 30 * time.Second

	// Binary name for testing
	testBinaryName = "mcp-shell"
)

// buildTestBinary builds the CLI binary for end-to-end testing.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Get the project root directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file path")
	}
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// Create a temporary directory for the test binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, testBinaryName)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Build the binary
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/mcp-shell")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v\nOutput: %s", err, output)
	}

	// Verify the binary was created
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Test binary not found after build: %v", err)
	}

	return binaryPath
}

// runCommand executes the CLI binary with given arguments and returns stdout, stderr, and exit code.
func runCommand(t *testing.T, binaryPath string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)

	// Capture stdout and stderr
	stdoutBuf := &strings.Builder{}
	stderrBuf := &strings.Builder{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	// Run the command
	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	// Get exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Logf("Command execution error (not exit error): %v", err)
			exitCode = -1
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode
}

// TestE2EVersion tests the version command end-to-end.
func TestE2EVersion(t *testing.T) {
	binaryPath := buildTestBinary(t)

	// Test version command
	stdout, stderr, exitCode := runCommand(t, binaryPath, "version")

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	// Verify output contains expected information
	expectedStrings := []string{
		"MCP Shell",
		"version",
		"Built with go",
		"Licensed under GPL-3.0",
		"Copyright (C) 2025 Rafael",
	}

	for _, expectedString := range expectedStrings {
		if !strings.Contains(stdout, expectedString) {
			t.Errorf("Expected version output to contain '%s', but it didn't.\nActual output: %s", expectedString, stdout)
		}
	}
}

// TestE2ERootCommand tests the root command behavior.
func TestE2ERootCommand(t *testing.T) {
	binaryPath := buildTestBinary(t)

	// Test root command (no arguments)
	stdout, stderr, exitCode := runCommand(t, binaryPath)

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	// Verify output contains expected information
	expectedStrings := []string{
		"MCP Shell",
		"MCP Shell Server for serving shell AI models",
		"Use 'mcp-shell help'",
	}

	for _, expectedString := range expectedStrings {
		if !strings.Contains(stdout, expectedString) {
			t.Errorf("Expected root output to contain '%s', but it didn't.\nActual output: %s", expectedString, stdout)
		}
	}
}

// TestE2EHelp tests the help command functionality.
func TestE2EHelp(t *testing.T) {
	binaryPath := buildTestBinary(t)

	testCases := []struct {
		name string
		args []string
	}{
		{"help flag", []string{"--help"}},
		{"help command", []string{"help"}},
		{"help for version", []string{"help", "version"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCommand(t, binaryPath, tc.args...)

			// Help should exit with code 0
			if exitCode != 0 {
				t.Errorf("Expected exit code 0 for help, got %d. Stderr: %s", exitCode, stderr)
			}

			// Help output should contain usage information
			if !strings.Contains(stdout, "Usage:") {
				t.Errorf("Expected help output to contain 'Usage:', but it didn't.\nActual output: %s", stdout)
			}
		})
	}
}

// TestE2EInvalidCommand tests behavior with invalid commands.
func TestE2EInvalidCommand(t *testing.T) {
	binaryPath := buildTestBinary(t)

	// Test invalid command
	stdout, stderr, exitCode := runCommand(t, binaryPath, "nonexistent-command")

	// Should exit with non-zero code
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid command, got %d", exitCode)
	}

	// Should show error message
	combinedOutput := stdout + stderr
	if !strings.Contains(combinedOutput, "unknown command") && !strings.Contains(combinedOutput, "Error:") {
		t.Errorf("Expected error message for invalid command.\nStdout: %s\nStderr: %s", stdout, stderr)
	}
}

// TestE2EConfigFile tests configuration file handling.
func TestE2EConfigFile(t *testing.T) {
	binaryPath := buildTestBinary(t)
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

	// Test with config file
	stdout, stderr, exitCode := runCommand(t, binaryPath, "--config", configPath, "--help")

	// Should still work with config file
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 with config file, got %d. Stderr: %s", exitCode, stderr)
	}

	// Should show help (config file shouldn't interfere with basic operations)
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("Expected help output with config file.\nStdout: %s", stdout)
	}
}

// TestE2EFlags tests various command-line flags.
func TestE2EFlags(t *testing.T) {
	binaryPath := buildTestBinary(t)

	testCases := []struct {
		name string
		args []string
	}{
		{"verbose flag", []string{"--verbose", "--help"}},
		{"debug flag", []string{"--debug", "--help"}},
		{"both flags", []string{"--verbose", "--debug", "--help"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCommand(t, binaryPath, tc.args...)

			// Flags shouldn't cause errors when combined with help
			if exitCode != 0 {
				t.Errorf("Expected exit code 0 with flags, got %d. Stderr: %s", exitCode, stderr)
			}

			// Should still show help
			if !strings.Contains(stdout, "Usage:") {
				t.Errorf("Expected help output with flags.\nStdout: %s", stdout)
			}
		})
	}
}

// TestE2EEnvironmentVariables tests environment variable handling.
func TestE2EEnvironmentVariables(t *testing.T) {
	binaryPath := buildTestBinary(t)

	// Set environment variables
	envVars := map[string]string{
		"MCP-SHELL_VERBOSE": "true",
		"MCP-SHELL_DEBUG":   "true",
	}

	// Prepare environment
	env := os.Environ()
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Run command with environment variables
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--help")
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
			t.Errorf("Command failed with environment variables: %v\nOutput: %s", err, output)
		}
	}

	// Should still work with environment variables
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output with environment variables.\nOutput: %s", string(output))
	}
}

// TestE2EInteractiveMode tests interactive behavior (if applicable).
func TestE2EInteractiveMode(t *testing.T) {
	// Skip this test if running in CI or non-interactive environment
	if os.Getenv("CI") != "" || !isTerminal() {
		t.Skip("Skipping interactive test in non-interactive environment")
	}

	binaryPath := buildTestBinary(t)

	// This is a placeholder for interactive testing
	// In a real application, you might test interactive prompts,
	// menu navigation, or other user interactions

	// For now, just verify the binary can be executed
	stdout, stderr, exitCode := runCommand(t, binaryPath, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 in interactive test, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("Expected help output in interactive test.\nStdout: %s", stdout)
	}
}

// isTerminal checks if the current environment supports terminal interaction.
func isTerminal() bool {
	// Simple check for terminal - you might want to use a more sophisticated method
	if os.Getenv("TERM") == "" {
		return false
	}

	// Check if stdin is a terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Mode()&os.ModeCharDevice != 0
}
