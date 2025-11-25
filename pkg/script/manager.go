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

// Package script provides script discovery, loading, and management capabilities
package script

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

// Manager handles script discovery, loading, and lifecycle
type Manager struct {
	config  *config.Config
	scripts map[string]*LoadedScript
	aliases map[string]*config.Alias
}

// LoadedScript represents a discovered and validated script
type LoadedScript struct {
	// Configuration from config file
	Config config.Script

	// Discovered metadata
	AbsolutePath string
	FileInfo     os.FileInfo
	IsExecutable bool
	Source       string // "config" or folder name
}

// NewManager creates a new script manager
func NewManager(cfg *config.Config) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	m := &Manager{
		config:  cfg,
		scripts: make(map[string]*LoadedScript),
		aliases: make(map[string]*config.Alias),
	}

	// Load all configured scripts and aliases
	if err := m.Reload(); err != nil {
		return nil, fmt.Errorf("failed to load scripts: %w", err)
	}

	return m, nil
}

// Reload reloads all scripts and aliases from configuration
func (m *Manager) Reload() error {
	// Clear existing scripts and aliases
	m.scripts = make(map[string]*LoadedScript)
	m.aliases = make(map[string]*config.Alias)

	// Load aliases
	for i := range m.config.Aliases {
		alias := &m.config.Aliases[i]
		m.aliases[alias.Name] = alias
	}

	// Load explicitly configured scripts
	for i := range m.config.Scripts {
		script := &m.config.Scripts[i]
		if err := m.loadScript(script, "config"); err != nil {
			return fmt.Errorf("failed to load script '%s': %w", script.Name, err)
		}
	}

	// Discover scripts from folders
	for _, folder := range m.config.ScriptFolders {
		if err := m.discoverFolder(folder); err != nil {
			return fmt.Errorf("failed to discover folder '%s': %w", folder.Name, err)
		}
	}

	return nil
}

// loadScript loads and validates a single script
func (m *Manager) loadScript(script *config.Script, source string) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(script.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("script file not found: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Check if executable (Unix permissions)
	isExecutable := fileInfo.Mode()&0111 != 0

	// Validate interpreter is allowed
	if script.Interpreter != "" {
		if !m.isInterpreterAllowed(script.Interpreter) {
			return fmt.Errorf("interpreter '%s' is not in allowed list", script.Interpreter)
		}
	}

	// Store loaded script
	loaded := &LoadedScript{
		Config:       *script,
		AbsolutePath: absPath,
		FileInfo:     fileInfo,
		IsExecutable: isExecutable,
		Source:       source,
	}

	m.scripts[script.Name] = loaded
	return nil
}

// discoverFolder discovers scripts from a folder using glob patterns
func (m *Manager) discoverFolder(folder config.ScriptFolder) error {
	// Use glob to find matching files
	matches, err := filepath.Glob(folder.Path)
	if err != nil {
		return fmt.Errorf("glob pattern error: %w", err)
	}

	for _, match := range matches {
		// Get file info
		fileInfo, err := os.Stat(match)
		if err != nil {
			continue // Skip files that can't be accessed
		}

		if fileInfo.IsDir() {
			continue // Skip directories
		}

		// Create script configuration from discovered file
		scriptName := fmt.Sprintf("%s:%s", folder.Name, filepath.Base(match))

		// Skip if already exists (explicit config takes priority)
		if _, exists := m.scripts[scriptName]; exists {
			continue
		}

		script := config.Script{
			Name:        scriptName,
			Description: folder.Description,
			Path:        match,
			Interpreter: folder.DefaultInterpreter,
			Parameters:  make(map[string]config.Parameter),
		}

		if err := m.loadScript(&script, folder.Name); err != nil {
			// Log warning but continue discovering other scripts
			continue
		}
	}

	return nil
}

// isInterpreterAllowed checks if an interpreter is in the allowed list
func (m *Manager) isInterpreterAllowed(interpreter string) bool {
	for _, allowed := range m.config.Security.AllowedInterpreters {
		if interpreter == allowed {
			return true
		}
	}
	return false
}

// GetScript retrieves a loaded script by name
func (m *Manager) GetScript(name string) (*LoadedScript, error) {
	script, exists := m.scripts[name]
	if !exists {
		return nil, fmt.Errorf("script '%s' not found", name)
	}
	return script, nil
}

// GetAlias retrieves an alias by name
func (m *Manager) GetAlias(name string) (*config.Alias, error) {
	alias, exists := m.aliases[name]
	if !exists {
		return nil, fmt.Errorf("alias '%s' not found", name)
	}
	return alias, nil
}

// ListScripts returns all loaded scripts
func (m *Manager) ListScripts() []*LoadedScript {
	scripts := make([]*LoadedScript, 0, len(m.scripts))
	for _, script := range m.scripts {
		scripts = append(scripts, script)
	}
	return scripts
}

// ListAliases returns all configured aliases
func (m *Manager) ListAliases() []*config.Alias {
	aliases := make([]*config.Alias, 0, len(m.aliases))
	for _, alias := range m.aliases {
		aliases = append(aliases, alias)
	}
	return aliases
}

// ValidateScript performs validation checks on a script
func (m *Manager) ValidateScript(name string) error {
	script, err := m.GetScript(name)
	if err != nil {
		return err
	}

	// Check file still exists
	if _, err := os.Stat(script.AbsolutePath); err != nil {
		return fmt.Errorf("script file no longer exists: %w", err)
	}

	// Check interpreter is still allowed
	if script.Config.Interpreter != "" {
		if !m.isInterpreterAllowed(script.Config.Interpreter) {
			return fmt.Errorf("interpreter '%s' is no longer allowed", script.Config.Interpreter)
		}
	}

	return nil
}

// CreateScript creates a new script file (if allowed by security settings)
func (m *Manager) CreateScript(name, content, interpreter string) error {
	if !m.config.Security.AllowScriptCreation {
		return fmt.Errorf("script creation is disabled by security settings")
	}

	if !m.isInterpreterAllowed(interpreter) {
		return fmt.Errorf("interpreter '%s' is not allowed", interpreter)
	}

	// Create script in the designated creation path
	scriptPath := filepath.Join(m.config.Security.ScriptCreationPath, name)
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Ensure creation directory exists
	if err := os.MkdirAll(filepath.Dir(absPath), 0750); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// Write script file
	// #nosec G306 -- Script files need execute permission (0700 = rwx------)
	if err := os.WriteFile(absPath, []byte(content), 0700); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// Add to loaded scripts
	script := config.Script{
		Name:        name,
		Description: "User-created script",
		Path:        absPath,
		Interpreter: interpreter,
		Parameters:  make(map[string]config.Parameter),
	}

	if err := m.loadScript(&script, "user-created"); err != nil {
		// Clean up file if loading fails
		_ = os.Remove(absPath)
		return fmt.Errorf("failed to load created script: %w", err)
	}

	return nil
}
