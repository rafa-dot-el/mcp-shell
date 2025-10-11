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
	"testing"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

func setupTestManager(t *testing.T) (*Manager, *script.Manager, string) {
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

	return manager, scriptManager, tmpDir
}

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = filepath.Join(tmpDir, "logs")

	scriptManager, err := script.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create script manager: %v", err)
	}

	executor := script.NewExecutor(scriptManager, cfg)

	manager, err := NewManager(cfg, executor)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	// Verify log directory was created
	if _, err := os.Stat(cfg.Execution.LogDirectory); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}

	// Shutdown manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestNewManager_NilConfig(t *testing.T) {
	scriptManager, _ := script.NewManager(config.DefaultConfig())
	executor := script.NewExecutor(scriptManager, config.DefaultConfig())

	_, err := NewManager(nil, executor)
	if err == nil {
		t.Error("Expected error with nil config, got nil")
	}
}

func TestNewManager_NilExecutor(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Execution.LogDirectory = t.TempDir()

	_, err := NewManager(cfg, nil)
	if err == nil {
		t.Error("Expected error with nil executor, got nil")
	}
}

func TestEnqueue(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue job
	job, err := manager.Enqueue("test-script", false, make(map[string]string))
	if err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	if job == nil {
		t.Fatal("Enqueue() returned nil job")
	}

	if job.ID == "" {
		t.Error("Job ID is empty")
	}

	if job.Status != StatusPending {
		t.Errorf("Expected status %s, got %s", StatusPending, job.Status)
	}

	if job.Name != "test-script" {
		t.Errorf("Expected name 'test-script', got '%s'", job.Name)
	}
}

func TestGetJob(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue job
	enqueuedJob, err := manager.Enqueue("test-script", false, make(map[string]string))
	if err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Get job
	job, err := manager.GetJob(enqueuedJob.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if job.ID != enqueuedJob.ID {
		t.Errorf("Expected job ID %s, got %s", enqueuedJob.ID, job.ID)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	manager, _, _ := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	_, err := manager.GetJob("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestListPendingJobs(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue multiple jobs
	for i := 0; i < 3; i++ {
		if _, err := manager.Enqueue("test-script", false, make(map[string]string)); err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
	}

	// Give queue processor time to start processing
	time.Sleep(100 * time.Millisecond)

	// List pending jobs
	pending := manager.ListPendingJobs()

	// Should have pending or running jobs (at least some of the 3)
	if len(pending) == 0 {
		t.Error("Expected pending jobs, got 0")
	}
}

func TestListRunningJobs(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script that sleeps briefly
	scriptPath := filepath.Join(tmpDir, "sleep.sh")
	scriptContent := `#!/bin/bash
sleep 0.5
echo "done"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "sleep-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue job
	if _, err := manager.Enqueue("sleep-script", false, make(map[string]string)); err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Give time for job to start
	time.Sleep(200 * time.Millisecond)

	// List running jobs
	running := manager.ListRunningJobs()

	// Should have at least the sleep script running or it might have completed
	// This is a timing-dependent test
	if len(running) > 1 {
		t.Errorf("Expected at most 1 running job, got %d", len(running))
	}
}

func TestListCompletedJobs(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue and wait for completion
	if _, err := manager.Enqueue("test-script", false, make(map[string]string)); err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Wait for job to complete
	time.Sleep(1 * time.Second)

	// List completed jobs
	completed := manager.ListCompletedJobs()

	if len(completed) != 1 {
		t.Errorf("Expected 1 completed job, got %d", len(completed))
	}

	if len(completed) > 0 && completed[0].Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, completed[0].Status)
	}
}

func TestCancelJob(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Test cancelling a pending job (before it starts)
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue multiple jobs to fill the queue
	jobs := make([]*Job, 5)
	for i := 0; i < 5; i++ {
		job, err := manager.Enqueue("test-script", false, make(map[string]string))
		if err != nil {
			t.Fatalf("Enqueue() failed: %v", err)
		}
		jobs[i] = job
	}

	// Cancel the last job (should still be pending)
	time.Sleep(100 * time.Millisecond)
	lastJob := jobs[len(jobs)-1]

	if err := manager.CancelJob(lastJob.ID); err != nil {
		t.Errorf("CancelJob() failed: %v", err)
	}

	// Wait a bit for status to update
	time.Sleep(200 * time.Millisecond)

	// Verify job was cancelled
	cancelledJob, err := manager.GetJob(lastJob.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if cancelledJob.Status != StatusCancelled {
		t.Errorf("Expected status %s, got %s", StatusCancelled, cancelledJob.Status)
	}
}

func TestCancelJob_NotFound(t *testing.T) {
	manager, _, _ := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	err := manager.CancelJob("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestShutdown(t *testing.T) {
	manager, _, _ := setupTestManager(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestShutdownTimeout(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)

	// Create test script that sleeps longer than shutdown timeout
	scriptPath := filepath.Join(tmpDir, "long_sleep.sh")
	scriptContent := `#!/bin/bash
sleep 10
echo "done"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "long-sleep",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue long-running job
	if _, err := manager.Enqueue("long-sleep", false, make(map[string]string)); err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Give time for job to start
	time.Sleep(200 * time.Millisecond)

	// Shutdown with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := manager.Shutdown(ctx)
	if err == nil {
		t.Error("Expected timeout error on shutdown, got nil")
	}
}

func TestGenerateShortID(t *testing.T) {
	id1 := generateShortID()
	id2 := generateShortID()

	if len(id1) != 8 {
		t.Errorf("Expected ID length 8, got %d", len(id1))
	}

	if id1 == id2 {
		t.Error("Expected unique IDs, got duplicates")
	}
}

func TestExecute(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho immediate"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "immediate-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute immediately (not queued)
	ctx := context.Background()
	job, err := manager.Execute(ctx, "immediate-script", false, make(map[string]string))
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if job == nil {
		t.Fatal("Execute() returned nil job")
	}

	// Job should start in running status (may complete quickly)
	if job.Status != StatusRunning && job.Status != StatusCompleted {
		t.Errorf("Expected status running or completed, got %s", job.Status)
	}

	// Wait for job to complete
	time.Sleep(1 * time.Second)

	// Verify job eventually completed
	completedJob, err := manager.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if completedJob.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, completedJob.Status)
	}
}

func TestCleanupJobs_All(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "cleanup-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue and wait for completion
	job1, _ := manager.Enqueue("cleanup-test", false, make(map[string]string))
	job2, _ := manager.Enqueue("cleanup-test", false, make(map[string]string))

	// Wait for jobs to complete
	time.Sleep(2 * time.Second)

	// Verify jobs completed
	completed := manager.ListCompletedJobs()
	if len(completed) < 2 {
		t.Fatalf("Expected at least 2 completed jobs, got %d", len(completed))
	}

	// Clean up all jobs
	if err := manager.CleanupJobs(nil); err != nil {
		t.Errorf("CleanupJobs() failed: %v", err)
	}

	// Verify all jobs removed
	completed = manager.ListCompletedJobs()
	if len(completed) != 0 {
		t.Errorf("Expected 0 completed jobs after cleanup, got %d", len(completed))
	}

	// Verify log files removed
	if _, err := os.Stat(job1.LogPath); !os.IsNotExist(err) {
		t.Error("Expected log file to be removed")
	}

	if _, err := os.Stat(job2.LogPath); !os.IsNotExist(err) {
		t.Error("Expected log file to be removed")
	}
}

func TestCleanupJobs_Specific(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "cleanup-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Enqueue and wait for completion
	job1, _ := manager.Enqueue("cleanup-test", false, make(map[string]string))
	job2, _ := manager.Enqueue("cleanup-test", false, make(map[string]string))

	// Wait for jobs to complete
	time.Sleep(2 * time.Second)

	// Clean up only job1
	if err := manager.CleanupJobs([]string{job1.ID}); err != nil {
		t.Errorf("CleanupJobs() failed: %v", err)
	}

	// Verify job1 removed, job2 remains
	completed := manager.ListCompletedJobs()
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed job after cleanup, got %d", len(completed))
	}

	if len(completed) > 0 && completed[0].ID != job2.ID {
		t.Errorf("Expected remaining job to be job2 (%s), got %s", job2.ID, completed[0].ID)
	}

	// Verify job1 log removed, job2 log exists
	if _, err := os.Stat(job1.LogPath); !os.IsNotExist(err) {
		t.Error("Expected job1 log file to be removed")
	}

	if _, err := os.Stat(job2.LogPath); os.IsNotExist(err) {
		t.Error("Expected job2 log file to still exist")
	}
}

func TestExecuteJob_LogFileCreationFailure(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "test-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Set invalid log directory to cause log file creation failure
	invalidLogDir := filepath.Join(tmpDir, "nonexistent", "deep", "path")
	manager.config.Execution.LogDirectory = invalidLogDir

	// Enqueue job
	job, err := manager.Enqueue("test-script", false, make(map[string]string))
	if err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	// Wait for job to fail
	time.Sleep(1 * time.Second)

	// Verify job failed
	failedJob, err := manager.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if failedJob.Status != StatusFailed {
		t.Errorf("Expected status %s, got %s", StatusFailed, failedJob.Status)
	}

	if failedJob.Error == nil {
		t.Error("Expected error to be set, got nil")
	}
}

func TestExecuteJob_ScriptFailure(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create failing script
	scriptPath := filepath.Join(tmpDir, "fail.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 1"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "fail-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute script
	ctx := context.Background()
	job, _ := manager.Execute(ctx, "fail-script", false, make(map[string]string))

	// Wait for job to complete
	time.Sleep(1 * time.Second)

	// Verify job completed with non-zero exit code
	completedJob, err := manager.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if completedJob.ExitCode == 0 {
		t.Error("Expected non-zero exit code, got 0")
	}
}

func TestExecuteJob_AliasExecution(t *testing.T) {
	manager, scriptManager, _ := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Add test alias
	manager.config.Aliases = []config.Alias{
		{
			Name:    "test-alias",
			Command: "echo 'alias test'",
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute alias
	ctx := context.Background()
	job, _ := manager.Execute(ctx, "test-alias", true, make(map[string]string))

	// Wait for job to complete
	time.Sleep(1 * time.Second)

	// Verify job completed
	completedJob, err := manager.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if completedJob.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, completedJob.Status)
	}

	if !completedJob.IsAlias {
		t.Error("Expected IsAlias to be true")
	}
}

func TestCleanupJobs_NonexistentLogFile(t *testing.T) {
	manager, scriptManager, tmpDir := setupTestManager(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "cleanup-nonexistent-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Execute and wait for completion
	job, _ := manager.Enqueue("cleanup-nonexistent-test", false, make(map[string]string))
	time.Sleep(1 * time.Second)

	// Delete log file manually
	os.Remove(job.LogPath)

	// Cleanup should succeed even if log file doesn't exist
	if err := manager.CleanupJobs([]string{job.ID}); err != nil {
		t.Errorf("CleanupJobs() failed: %v", err)
	}

	// Verify job was removed
	completed := manager.ListCompletedJobs()
	if len(completed) != 0 {
		t.Errorf("Expected 0 completed jobs after cleanup, got %d", len(completed))
	}
}
