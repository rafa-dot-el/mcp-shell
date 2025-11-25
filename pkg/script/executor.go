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
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

// ExecutionRequest represents a request to execute a script or alias
type ExecutionRequest struct {
	// Name of the script or alias
	Name string

	// Parameters for script execution (ignored for aliases)
	Parameters map[string]string

	// Timeout for execution (0 means use default)
	Timeout time.Duration

	// WorkingDirectory for script execution
	WorkingDirectory string
}

// ExecutionResult represents the result of script execution
type ExecutionResult struct {
	// ExitCode from the executed command
	ExitCode int

	// Stdout from the command
	Stdout string

	// Stderr from the command
	Stderr string

	// Duration of execution
	Duration time.Duration

	// Error if execution failed
	Error error
}

// Executor handles script and alias execution
type Executor struct {
	manager *Manager
	config  *config.Config
}

// NewExecutor creates a new script executor
func NewExecutor(manager *Manager, cfg *config.Config) *Executor {
	return &Executor{
		manager: manager,
		config:  cfg,
	}
}

// ExecuteAlias executes an alias command
func (e *Executor) ExecuteAlias(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	alias, err := e.manager.GetAlias(req.Name)
	if err != nil {
		return nil, fmt.Errorf("alias not found: %w", err)
	}

	// Determine timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = time.Duration(e.config.Execution.DefaultTimeout) * time.Second
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute alias command using shell
	return e.executeCommand(execCtx, "bash", []string{"-c", alias.Command}, req.WorkingDirectory)
}

// ExecuteScript executes a script with parameters
func (e *Executor) ExecuteScript(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	script, err := e.manager.GetScript(req.Name)
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	// Validate script still exists and is valid
	if err := e.manager.ValidateScript(req.Name); err != nil {
		return nil, fmt.Errorf("script validation failed: %w", err)
	}

	// Validate required parameters are provided
	if err := e.validateParameters(script, req.Parameters); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Build command arguments with parameter substitution
	args, err := e.buildArguments(script, req.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to build arguments: %w", err)
	}

	// Determine interpreter
	interpreter := script.Config.Interpreter
	if interpreter == "" {
		// Default to bash if not specified
		interpreter = "bash"
	}

	// Determine timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = time.Duration(e.config.Execution.DefaultTimeout) * time.Second
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute script
	cmdArgs := append([]string{script.AbsolutePath}, args...)
	return e.executeCommand(execCtx, interpreter, cmdArgs, req.WorkingDirectory)
}

// validateParameters checks that all required parameters are provided
func (e *Executor) validateParameters(script *LoadedScript, params map[string]string) error {
	for paramName, paramConfig := range script.Config.Parameters {
		value, provided := params[paramName]

		// Check required parameters
		if paramConfig.Required && !provided {
			return fmt.Errorf("required parameter '%s' not provided", paramName)
		}

		// Validate against valid values if specified
		if provided && len(paramConfig.ValidValues) > 0 {
			valid := false
			for _, validValue := range paramConfig.ValidValues {
				if value == validValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("parameter '%s' has invalid value '%s', must be one of: %v",
					paramName, value, paramConfig.ValidValues)
			}
		}
	}

	return nil
}

// buildArguments constructs command arguments with parameter substitution
//
//nolint:unparam // error return reserved for future parameter validation
func (e *Executor) buildArguments(script *LoadedScript, params map[string]string) ([]string, error) {
	var args []string

	for paramName, paramConfig := range script.Config.Parameters {
		// Get parameter value (use default if not provided)
		value, provided := params[paramName]
		if !provided {
			value = paramConfig.Default
		}

		// Skip if no value and not required
		if value == "" && !paramConfig.Required {
			continue
		}

		// Substitute parameter using setter pattern
		setter := paramConfig.Setter
		substituted := strings.ReplaceAll(setter, "{}", value)

		// Split setter into arguments (handle multi-word setters)
		setterArgs := strings.Fields(substituted)
		args = append(args, setterArgs...)
	}

	return args, nil
}

// executeCommand executes a command and captures output
func (e *Executor) executeCommand(ctx context.Context, interpreter string, args []string, workDir string) (*ExecutionResult, error) {
	start := time.Now()

	// Create command
	cmd := exec.CommandContext(ctx, interpreter, args...)

	if workDir != "" {
		cmd.Dir = workDir
	}

	// Capture stdout and stderr
	stdout, err := cmd.Output()

	result := &ExecutionResult{
		Duration: time.Since(start),
		Stdout:   string(stdout),
	}

	if err != nil {
		// Handle exit errors
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Stderr = string(exitErr.Stderr)
			result.Error = fmt.Errorf("command exited with code %d", result.ExitCode)
		} else {
			result.Error = fmt.Errorf("command execution failed: %w", err)
		}
		return result, result.Error
	}

	result.ExitCode = 0
	return result, nil
}
