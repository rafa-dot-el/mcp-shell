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

package mcp

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

func setupTestConfig(t *testing.T) *config.Config {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")

	return cfg
}

func TestNewServer(t *testing.T) {
	cfg := setupTestConfig(t)

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() returned nil server")
	}

	// Verify components initialized
	if server.config == nil {
		t.Error("Server config is nil")
	}

	if server.manager == nil {
		t.Error("Server manager is nil")
	}

	if server.executor == nil {
		t.Error("Server executor is nil")
	}

	if server.jobMgr == nil {
		t.Error("Server job manager is nil")
	}

	if server.scheduler == nil {
		t.Error("Server scheduler is nil")
	}

	if server.resources == nil {
		t.Error("Server resource handler is nil")
	}

	if server.prompts == nil {
		t.Error("Server prompt handler is nil")
	}

	if server.tools == nil {
		t.Error("Server tool handler is nil")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestNewServer_NilConfig(t *testing.T) {
	_, err := NewServer(nil)
	if err == nil {
		t.Error("Expected error for nil config, got nil")
	}
}

func TestServer_Shutdown(t *testing.T) {
	cfg := setupTestConfig(t)

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestServer_GetServerInfo(t *testing.T) {
	cfg := setupTestConfig(t)
	cfg.MCP.Name = "Test Server"
	cfg.MCP.Version = "1.0.0"

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	info := server.GetServerInfo()

	if info.Name != "Test Server" {
		t.Errorf("Expected name 'Test Server', got '%s'", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
}
