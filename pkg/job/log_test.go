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

package job

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

func setupTestManagerWithLog(t *testing.T) (*Manager, *script.Manager, string, *Job) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")
	cfg.Execution.MaxParallelJobs = 2

	scriptManager, err := script.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create script manager: %v", err)
	}

	executor := script.NewExecutor(scriptManager, cfg)

	manager, err := NewManager(cfg, executor)
	if err != nil {
		t.Fatalf("Failed to create job manager: %v", err)
	}

	// Create test script with multi-line output
	scriptPath := filepath.Join(tmpDir, "log_test.sh")
	scriptContent := `#!/bin/bash
echo "Line 1: Starting"
echo "Line 2: Processing"
echo "ERROR: Something went wrong"
echo "Line 4: Continuing"
echo "Line 5: Done"
echo "DEBUG: Debug information"
echo "Line 7: Final line"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "log-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute job and wait for completion
	ctx := context.Background()
	job, err := manager.Execute(ctx, "log-test", false, make(map[string]string))
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Wait for job to complete
	time.Sleep(1 * time.Second)

	return manager, scriptManager, tmpDir, job
}

func TestTailLog(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Test default tail (5 lines)
	lines, err := manager.TailLog(job.ID, TailOptions{})
	if err != nil {
		t.Fatalf("TailLog() failed: %v", err)
	}

	if len(lines) > 5 {
		t.Errorf("Expected at most 5 lines, got %d", len(lines))
	}

	// Test custom line count
	lines, err = manager.TailLog(job.ID, TailOptions{Lines: 3})
	if err != nil {
		t.Fatalf("TailLog() with custom lines failed: %v", err)
	}

	if len(lines) > 3 {
		t.Errorf("Expected at most 3 lines, got %d", len(lines))
	}
}

func TestTailLog_WithFilter(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Test with filter
	lines, err := manager.TailLog(job.ID, TailOptions{
		Lines:  10,
		Filter: "Line \\d+:",
	})
	if err != nil {
		t.Fatalf("TailLog() with filter failed: %v", err)
	}

	// All lines should match the filter
	for _, line := range lines {
		if !strings.Contains(line, "Line") {
			t.Errorf("Expected filtered line to contain 'Line', got: %s", line)
		}
	}
}

func TestTailLog_InvalidFilter(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Test with invalid regex
	_, err := manager.TailLog(job.ID, TailOptions{
		Lines:  10,
		Filter: "[invalid(regex",
	})
	if err == nil {
		t.Error("Expected error for invalid regex, got nil")
	}
}

func TestTailLog_JobNotFound(t *testing.T) {
	manager, _, _, _ := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	_, err := manager.TailLog("nonexistent", TailOptions{})
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestReadFullLog(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	content, err := manager.ReadFullLog(job.ID)
	if err != nil {
		t.Fatalf("ReadFullLog() failed: %v", err)
	}

	if content == "" {
		t.Error("Expected non-empty log content")
	}

	// Should contain job information
	if !strings.Contains(content, "Job "+job.ID) {
		t.Error("Expected log to contain job ID")
	}
}

func TestReadFullLog_JobNotFound(t *testing.T) {
	manager, _, _, _ := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	_, err := manager.ReadFullLog("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestSearchLog(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Search for ERROR lines
	matches, err := manager.SearchLog(job.ID, "ERROR:")
	if err != nil {
		t.Fatalf("SearchLog() failed: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected to find ERROR lines")
	}

	// All matches should contain ERROR
	for _, match := range matches {
		if !strings.Contains(match, "ERROR:") {
			t.Errorf("Expected match to contain 'ERROR:', got: %s", match)
		}
		// Should include line number
		if !strings.Contains(match, ":") {
			t.Errorf("Expected match to include line number, got: %s", match)
		}
	}
}

func TestSearchLog_NoMatches(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Search for non-existent pattern
	matches, err := manager.SearchLog(job.ID, "NONEXISTENT_PATTERN")
	if err != nil {
		t.Fatalf("SearchLog() failed: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected no matches, got %d", len(matches))
	}
}

func TestSearchLog_InvalidPattern(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	_, err := manager.SearchLog(job.ID, "[invalid(regex")
	if err == nil {
		t.Error("Expected error for invalid regex, got nil")
	}
}

func TestSearchLog_JobNotFound(t *testing.T) {
	manager, _, _, _ := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	_, err := manager.SearchLog("nonexistent", "pattern")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestGetLogStats(t *testing.T) {
	manager, _, _, job := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	stats, err := manager.GetLogStats(job.ID)
	if err != nil {
		t.Fatalf("GetLogStats() failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.Path != job.LogPath {
		t.Errorf("Expected path %s, got %s", job.LogPath, stats.Path)
	}

	if stats.Size == 0 {
		t.Error("Expected non-zero file size")
	}

	if stats.Lines == 0 {
		t.Error("Expected non-zero line count")
	}

	if stats.Words == 0 {
		t.Error("Expected non-zero word count")
	}

	if stats.Characters == 0 {
		t.Error("Expected non-zero character count")
	}

	if stats.Modified.IsZero() {
		t.Error("Expected non-zero modification time")
	}
}

func TestGetLogStats_JobNotFound(t *testing.T) {
	manager, _, _, _ := setupTestManagerWithLog(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	_, err := manager.GetLogStats("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestReadFullLog_MissingFile(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "missing-log-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute and wait for completion
	ctx := context.Background()
	job, _ := manager.Execute(ctx, "missing-log-test", false, make(map[string]string))
	time.Sleep(1 * time.Second)

	// Delete log file
	_ = os.Remove(job.LogPath)

	// Try to read missing log
	_, err := manager.ReadFullLog(job.ID)
	if err == nil {
		t.Error("Expected error for missing log file, got nil")
	}
}

func TestGetLogStats_MissingFile(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "missing-stats-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute and wait for completion
	ctx := context.Background()
	job, _ := manager.Execute(ctx, "missing-stats-test", false, make(map[string]string))
	time.Sleep(1 * time.Second)

	// Delete log file
	_ = os.Remove(job.LogPath)

	// Try to get stats for missing log
	_, err := manager.GetLogStats(job.ID)
	if err == nil {
		t.Error("Expected error for missing log file, got nil")
	}
}
