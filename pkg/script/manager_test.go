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
	"os"
	"path/filepath"
	"testing"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

func TestNewManager(t *testing.T) {
	cfg := config.DefaultConfig()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}
}

func TestNewManager_NilConfig(t *testing.T) {
	_, err := NewManager(nil)
	if err == nil {
		t.Error("Expected error with nil config, got nil")
	}
}

func TestManager_LoadAliases(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Aliases = []config.Alias{
		{
			Name:        "test-alias",
			Description: "Test alias",
			Command:     "echo test",
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Get the alias
	alias, err := manager.GetAlias("test-alias")
	if err != nil {
		t.Errorf("GetAlias() failed: %v", err)
	}

	if alias.Command != "echo test" {
		t.Errorf("Expected command 'echo test', got '%s'", alias.Command)
	}
}

func TestManager_GetAlias_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	_, err = manager.GetAlias("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent alias, got nil")
	}
}

func TestManager_LoadScript(t *testing.T) {
	// Create temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	scriptContent := `#!/bin/bash
echo "test script"
`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{
			Name:        "test-script",
			Description: "Test script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Get the script
	script, err := manager.GetScript("test-script")
	if err != nil {
		t.Errorf("GetScript() failed: %v", err)
	}

	if script.Config.Name != "test-script" {
		t.Errorf("Expected script name 'test-script', got '%s'", script.Config.Name)
	}

	if !script.IsExecutable {
		t.Error("Expected script to be executable")
	}
}

func TestManager_LoadScript_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{
			Name:        "missing",
			Description: "Missing script",
			Path:        "/nonexistent/script.sh",
			Interpreter: "bash",
		},
	}

	_, err := NewManager(cfg)
	if err == nil {
		t.Error("Expected error for nonexistent script, got nil")
	}
}

func TestManager_LoadScript_InvalidInterpreter(t *testing.T) {
	// Create temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	if err := os.WriteFile(scriptPath, []byte("echo test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{
			Name:        "test-script",
			Description: "Test script",
			Path:        scriptPath,
			Interpreter: "invalid-interpreter",
		},
	}

	_, err := NewManager(cfg)
	if err == nil {
		t.Error("Expected error for invalid interpreter, got nil")
	}
}

func TestManager_DiscoverFolder(t *testing.T) {
	// Create temporary directory with scripts
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0750); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	// Create test scripts
	for i := 1; i <= 3; i++ {
		scriptPath := filepath.Join(scriptsDir, filepath.Join("script"+string(rune('0'+i))+".sh"))
		if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
	}

	cfg := config.DefaultConfig()
	cfg.ScriptFolders = []config.ScriptFolder{
		{
			Name:               "test-folder",
			Description:        "Test scripts",
			Path:               filepath.Join(scriptsDir, "*.sh"),
			DefaultInterpreter: "bash",
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	scripts := manager.ListScripts()
	if len(scripts) != 3 {
		t.Errorf("Expected 3 discovered scripts, got %d", len(scripts))
	}
}

func TestManager_ListScripts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test scripts
	script1Path := filepath.Join(tmpDir, "script1.sh")
	script2Path := filepath.Join(tmpDir, "script2.sh")

	for _, path := range []string{script1Path, script2Path} {
		if err := os.WriteFile(path, []byte("#!/bin/bash\necho test"), 0750); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{Name: "script1", Path: script1Path, Interpreter: "bash"},
		{Name: "script2", Path: script2Path, Interpreter: "bash"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	scripts := manager.ListScripts()
	if len(scripts) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(scripts))
	}
}

func TestManager_ListAliases(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Aliases = []config.Alias{
		{Name: "alias1", Command: "echo 1"},
		{Name: "alias2", Command: "echo 2"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	aliases := manager.ListAliases()
	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}
}

func TestManager_ValidateScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{Name: "test", Path: scriptPath, Interpreter: "bash"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	err = manager.ValidateScript("test")
	if err != nil {
		t.Errorf("ValidateScript() failed: %v", err)
	}
}

func TestManager_ValidateScript_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	err = manager.ValidateScript("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent script, got nil")
	}
}

func TestManager_CreateScript(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = true
	cfg.Security.ScriptCreationPath = tmpDir

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	scriptContent := "#!/bin/bash\necho 'created script'"
	err = manager.CreateScript("created.sh", scriptContent, "bash")
	if err != nil {
		t.Errorf("CreateScript() failed: %v", err)
	}

	// Verify script was created and loaded
	script, err := manager.GetScript("created.sh")
	if err != nil {
		t.Errorf("GetScript() failed after creation: %v", err)
	}

	if script.Source != "user-created" {
		t.Errorf("Expected source 'user-created', got '%s'", script.Source)
	}

	// Verify file exists on filesystem
	createdPath := filepath.Join(tmpDir, "created.sh")
	if _, err := os.Stat(createdPath); err != nil {
		t.Errorf("Created script file not found: %v", err)
	}
}

func TestManager_CreateScript_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = false

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	err = manager.CreateScript("test.sh", "echo test", "bash")
	if err == nil {
		t.Error("Expected error when script creation is disabled, got nil")
	}
}

func TestManager_CreateScript_InvalidInterpreter(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = true
	cfg.Security.ScriptCreationPath = tmpDir

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	err = manager.CreateScript("test.sh", "echo test", "invalid-interpreter")
	if err == nil {
		t.Error("Expected error for invalid interpreter, got nil")
	}
}

func TestManager_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{Name: "test", Path: scriptPath, Interpreter: "bash"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Verify initial load
	scripts := manager.ListScripts()
	if len(scripts) != 1 {
		t.Errorf("Expected 1 script after initial load, got %d", len(scripts))
	}

	// Add another script to config
	script2Path := filepath.Join(tmpDir, "test2.sh")
	if err := os.WriteFile(script2Path, []byte("#!/bin/bash\necho test2"), 0750); err != nil {
		t.Fatalf("Failed to create second test script: %v", err)
	}

	cfg.Scripts = append(cfg.Scripts, config.Script{
		Name:        "test2",
		Path:        script2Path,
		Interpreter: "bash",
	})

	// Reload
	if err := manager.Reload(); err != nil {
		t.Errorf("Reload() failed: %v", err)
	}

	// Verify reload loaded new script
	scripts = manager.ListScripts()
	if len(scripts) != 2 {
		t.Errorf("Expected 2 scripts after reload, got %d", len(scripts))
	}
}

func TestManager_DiscoverFolder_InvalidGlob(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ScriptFolders = []config.ScriptFolder{
		{
			Name:               "invalid-glob",
			Description:        "Test invalid glob",
			Path:               "\\invalid[glob",
			DefaultInterpreter: "bash",
		},
	}

	_, err := NewManager(cfg)
	if err == nil {
		t.Error("Expected error for invalid glob pattern, got nil")
	}
}

func TestManager_DiscoverFolder_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0750); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	// Create a script file
	scriptPath := filepath.Join(scriptsDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create a subdirectory that should be skipped
	subDir := filepath.Join(scriptsDir, "subdir.sh")
	if err := os.MkdirAll(subDir, 0750); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.ScriptFolders = []config.ScriptFolder{
		{
			Name:               "test-folder",
			Description:        "Test scripts",
			Path:               filepath.Join(scriptsDir, "*.sh"),
			DefaultInterpreter: "bash",
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	scripts := manager.ListScripts()
	// Should only find the script file, not the directory
	if len(scripts) != 1 {
		t.Errorf("Expected 1 discovered script (directories should be skipped), got %d", len(scripts))
	}
}

func TestManager_DiscoverFolder_NonExecutableFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0750); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	// Create a non-executable script file
	scriptPath := filepath.Join(scriptsDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0640); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.ScriptFolders = []config.ScriptFolder{
		{
			Name:               "test-folder",
			Description:        "Test scripts",
			Path:               filepath.Join(scriptsDir, "*.sh"),
			DefaultInterpreter: "bash",
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Should still discover the script even if not executable
	scripts := manager.ListScripts()
	if len(scripts) != 1 {
		t.Errorf("Expected 1 discovered script, got %d", len(scripts))
	}

	// Verify IsExecutable is false
	script, err := manager.GetScript("test-folder:script.sh")
	if err != nil {
		t.Fatalf("GetScript() failed: %v", err)
	}

	if script.IsExecutable {
		t.Error("Expected script to not be executable")
	}
}

func TestManager_DiscoverFolder_ConfigScriptPriority(t *testing.T) {
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0750); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	// Create a script file
	scriptPath := filepath.Join(scriptsDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	// Add script folder that will try to discover the script
	cfg.ScriptFolders = []config.ScriptFolder{
		{
			Name:               "test-folder",
			Description:        "Discovered scripts",
			Path:               filepath.Join(scriptsDir, "*.sh"),
			DefaultInterpreter: "bash",
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// First discovery should work
	scripts := manager.ListScripts()
	if len(scripts) != 1 {
		t.Errorf("Expected 1 script after first discovery, got %d", len(scripts))
	}

	// Verify discovered script name has folder prefix
	script, err := manager.GetScript("test-folder:test.sh")
	if err != nil {
		t.Errorf("GetScript() failed for discovered script: %v", err)
	}

	if script.Config.Description != "Discovered scripts" {
		t.Errorf("Expected discovered script description, got '%s'", script.Config.Description)
	}

	// Reload should skip already-existing script
	if err := manager.Reload(); err != nil {
		t.Errorf("Reload() failed: %v", err)
	}

	scripts = manager.ListScripts()
	if len(scripts) != 1 {
		t.Errorf("Expected 1 script after reload (duplicate should be skipped), got %d", len(scripts))
	}
}

func TestManager_ValidateScript_FileDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{Name: "test", Path: scriptPath, Interpreter: "bash"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Delete the script file
	if err := os.Remove(scriptPath); err != nil {
		t.Fatalf("Failed to delete script file: %v", err)
	}

	// Validation should fail
	err = manager.ValidateScript("test")
	if err == nil {
		t.Error("Expected error when script file is deleted, got nil")
	}
}

func TestManager_ValidateScript_InterpreterDisallowed(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Scripts = []config.Script{
		{Name: "test", Path: scriptPath, Interpreter: "bash"},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Change config to disallow bash interpreter
	cfg.Security.AllowedInterpreters = []string{"python3"}

	// Validation should fail
	err = manager.ValidateScript("test")
	if err == nil {
		t.Error("Expected error when interpreter is no longer allowed, got nil")
	}
}

func TestManager_CreateScript_WithSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = true
	cfg.Security.ScriptCreationPath = tmpDir

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Create script with subdirectory in name
	scriptContent := "#!/bin/bash\necho 'subdirectory script'"
	err = manager.CreateScript("subdir/script.sh", scriptContent, "bash")
	if err != nil {
		t.Errorf("CreateScript() with subdirectory failed: %v", err)
	}

	// Verify file exists
	createdPath := filepath.Join(tmpDir, "subdir", "script.sh")
	if _, err := os.Stat(createdPath); err != nil {
		t.Errorf("Created script file not found at subdirectory path: %v", err)
	}
}

func TestManager_CreateScript_DirectoryCreationFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file where we want a directory
	blockingFile := filepath.Join(tmpDir, "blocked")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0640); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = true
	cfg.Security.ScriptCreationPath = blockingFile // This is a file, not a directory

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Attempt to create script should fail
	err = manager.CreateScript("script.sh", "#!/bin/bash\necho test", "bash")
	if err == nil {
		t.Error("Expected error when directory creation is blocked, got nil")
	}
}

func TestManager_CreateScript_ReadOnlyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	// Create directory
	if err := os.MkdirAll(readOnlyDir, 0750); err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}

	// Make directory read-only
	if err := os.Chmod(readOnlyDir, 0550); err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0750) // Restore permissions for cleanup

	cfg := config.DefaultConfig()
	cfg.Security.AllowScriptCreation = true
	cfg.Security.ScriptCreationPath = readOnlyDir

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Attempt to create script in read-only directory should fail
	err = manager.CreateScript("test.sh", "#!/bin/bash\necho test", "bash")
	if err == nil {
		t.Error("Expected error when writing to read-only directory, got nil")
	}
}
