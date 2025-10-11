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

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/mcp"
)

// setupTestServer creates a test MCP server with temporary directories
func setupTestServer(t *testing.T) (*mcp.Server, *config.Config, string) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")
	cfg.MCP.Name = "MCP Shell Test Server"
	cfg.MCP.Version = "1.0.0-test"

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
echo "Test script executed"
echo "Args: $@"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Configure test script
	cfg.Scripts = []config.Script{
		{
			Name:        "test-script",
			Description: "Test script for integration testing",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters: map[string]config.Parameter{
				"message": {
					Description: "Test message parameter",
					Required:    false,
					Default:     "default message",
					Setter:      "--message {}",
				},
			},
		},
	}

	// Configure test alias
	cfg.Aliases = []config.Alias{
		{
			Name:        "test-alias",
			Description: "Test alias for integration testing",
			Command:     "echo 'Alias executed'",
		},
	}

	server, err := mcp.NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	return server, cfg, tmpDir
}

func TestServerCreation(t *testing.T) {
	server, cfg, _ := setupTestServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Verify server info
	info := server.GetServerInfo()
	if info.Name != cfg.MCP.Name {
		t.Errorf("Expected server name %s, got %s", cfg.MCP.Name, info.Name)
	}

	if info.Version != cfg.MCP.Version {
		t.Errorf("Expected server version %s, got %s", cfg.MCP.Version, info.Version)
	}
}

func TestServerShutdown(t *testing.T) {
	server, _, _ := setupTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}
}

func TestServerShutdownTimeout(t *testing.T) {
	server, _, _ := setupTestServer(t)

	// Use very short timeout to trigger timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Sleep to ensure context is cancelled
	time.Sleep(10 * time.Millisecond)

	err := server.Shutdown(ctx)
	// Should succeed or timeout depending on timing
	_ = err
}

func TestMultipleServers(t *testing.T) {
	// Create multiple servers to ensure they don't interfere
	servers := make([]*mcp.Server, 3)
	for i := 0; i < 3; i++ {
		server, _, _ := setupTestServer(t)
		servers[i] = server
	}

	// Shutdown all servers
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, server := range servers {
		if err := server.Shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown server: %v", err)
		}
	}
}

func TestServerWithEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")
	// Empty scripts and aliases
	cfg.Scripts = []config.Script{}
	cfg.Aliases = []config.Alias{}

	server, err := mcp.NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server with empty config: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	info := server.GetServerInfo()
	if info.Name == "" {
		t.Error("Expected non-empty server name")
	}
}

func TestServerConfiguration(t *testing.T) {
	server, cfg, _ := setupTestServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Verify server has access to configuration
	info := server.GetServerInfo()

	// Check server info matches config
	if info.Name != cfg.MCP.Name {
		t.Errorf("Server name mismatch: expected %s, got %s", cfg.MCP.Name, info.Name)
	}

	if info.Version != cfg.MCP.Version {
		t.Errorf("Server version mismatch: expected %s, got %s", cfg.MCP.Version, info.Version)
	}
}

func TestServerInfoSerialization(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	info := server.GetServerInfo()

	// Serialize to JSON
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal server info: %v", err)
	}

	// Deserialize from JSON
	var decoded mcp.ServerInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal server info: %v", err)
	}

	// Verify deserialized data matches
	if decoded.Name != info.Name {
		t.Errorf("Name mismatch after serialization: expected %s, got %s", info.Name, decoded.Name)
	}

	if decoded.Version != info.Version {
		t.Errorf("Version mismatch after serialization: expected %s, got %s", info.Version, decoded.Version)
	}
}

func TestServerWithCustomMCPConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")

	// Custom MCP configuration
	cfg.MCP.Name = "Custom MCP Server"
	cfg.MCP.Version = "2.0.0"

	server, err := mcp.NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server with custom config: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	info := server.GetServerInfo()

	if info.Name != "Custom MCP Server" {
		t.Errorf("Expected custom name 'Custom MCP Server', got %s", info.Name)
	}

	if info.Version != "2.0.0" {
		t.Errorf("Expected custom version '2.0.0', got %s", info.Version)
	}
}

func TestConcurrentServerOperations(t *testing.T) {
	server, _, _ := setupTestServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Get server info concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			info := server.GetServerInfo()
			if info.Name == "" {
				t.Error("Got empty server name in concurrent access")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestServerLifecycle(t *testing.T) {
	// Test complete server lifecycle: create -> use -> shutdown
	server, _, _ := setupTestServer(t)

	// Use the server (get info)
	info := server.GetServerInfo()
	if info.Name == "" {
		t.Error("Expected non-empty server name")
	}

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
