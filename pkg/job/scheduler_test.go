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

func setupTestScheduler(t *testing.T) (*Scheduler, *Manager, *script.Manager, string) {
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

	scheduler := NewScheduler(manager)

	return scheduler, manager, scriptManager, tmpDir
}

func TestNewScheduler(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	if scheduler == nil {
		t.Fatal("NewScheduler() returned nil")
	}

	if scheduler.manager != manager {
		t.Error("Scheduler manager not set correctly")
	}

	if scheduler.cron == nil {
		t.Error("Scheduler cron not initialized")
	}

	if scheduler.scheduled == nil {
		t.Error("Scheduler scheduled map not initialized")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		manager.Shutdown(ctx)
	}()

	// Start scheduler
	scheduler.Start()

	// Stop scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := scheduler.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}

func TestSchedule(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "scheduled.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho scheduled"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "scheduled-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job to run every minute
	scheduledJob, err := scheduler.Schedule("scheduled-script", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	if scheduledJob == nil {
		t.Fatal("Schedule() returned nil job")
	}

	if scheduledJob.ID == "" {
		t.Error("Scheduled job ID is empty")
	}

	if scheduledJob.Name != "scheduled-script" {
		t.Errorf("Expected name 'scheduled-script', got '%s'", scheduledJob.Name)
	}

	if scheduledJob.Schedule != "* * * * *" {
		t.Errorf("Expected schedule '* * * * *', got '%s'", scheduledJob.Schedule)
	}

	if scheduledJob.NextRun.IsZero() {
		t.Error("NextRun time is zero")
	}

	if scheduledJob.EntryID == 0 {
		t.Error("EntryID is zero")
	}
}

func TestSchedule_InvalidCron(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Schedule with invalid cron expression
	_, err := scheduler.Schedule("test", false, make(map[string]string), "invalid cron")
	if err == nil {
		t.Error("Expected error for invalid cron expression, got nil")
	}
}

func TestScheduleOnce_Future(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "once.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho once"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "once-script",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job to run 2 seconds in the future
	runAt := time.Now().Add(2 * time.Second)
	scheduledJob, err := scheduler.ScheduleOnce("once-script", false, make(map[string]string), runAt)
	if err != nil {
		t.Fatalf("ScheduleOnce() failed: %v", err)
	}

	if scheduledJob == nil {
		t.Fatal("ScheduleOnce() returned nil job")
	}

	if scheduledJob.ID == "" {
		t.Error("Scheduled job ID is empty")
	}

	if scheduledJob.Name != "once-script" {
		t.Errorf("Expected name 'once-script', got '%s'", scheduledJob.Name)
	}

	if scheduledJob.NextRun != runAt {
		t.Errorf("Expected NextRun %v, got %v", runAt, scheduledJob.NextRun)
	}

	// Verify job is in scheduled list
	jobs := scheduler.ListScheduledJobs()
	if len(jobs) != 1 {
		t.Errorf("Expected 1 scheduled job, got %d", len(jobs))
	}

	// Wait for job to execute and be removed
	time.Sleep(3 * time.Second)

	// Verify job was removed from scheduled list
	jobs = scheduler.ListScheduledJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 scheduled jobs after execution, got %d", len(jobs))
	}

	// Verify job was enqueued/executed
	pending := manager.ListPendingJobs()
	completed := manager.ListCompletedJobs()
	total := len(pending) + len(completed)
	if total == 0 {
		t.Error("Expected job to be enqueued, found none")
	}
}

func TestScheduleOnce_Past(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Schedule job in the past
	runAt := time.Now().Add(-1 * time.Hour)
	_, err := scheduler.ScheduleOnce("test", false, make(map[string]string), runAt)
	if err == nil {
		t.Error("Expected error for past time, got nil")
	}
}

func TestUnschedule(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "unschedule.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "unschedule-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job
	scheduledJob, err := scheduler.Schedule("unschedule-test", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	// Verify job is scheduled
	jobs := scheduler.ListScheduledJobs()
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 scheduled job, got %d", len(jobs))
	}

	// Unschedule job
	if err := scheduler.Unschedule(scheduledJob.ID); err != nil {
		t.Errorf("Unschedule() failed: %v", err)
	}

	// Verify job is removed
	jobs = scheduler.ListScheduledJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 scheduled jobs after unschedule, got %d", len(jobs))
	}
}

func TestUnschedule_NotFound(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	err := scheduler.Unschedule("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestGetScheduledJob(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "get.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "get-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job
	scheduledJob, err := scheduler.Schedule("get-test", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	// Get scheduled job
	retrievedJob, err := scheduler.GetScheduledJob(scheduledJob.ID)
	if err != nil {
		t.Fatalf("GetScheduledJob() failed: %v", err)
	}

	if retrievedJob.ID != scheduledJob.ID {
		t.Errorf("Expected job ID %s, got %s", scheduledJob.ID, retrievedJob.ID)
	}
}

func TestGetScheduledJob_NotFound(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	_, err := scheduler.GetScheduledJob("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestListScheduledJobs(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Initially empty
	jobs := scheduler.ListScheduledJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs initially, got %d", len(jobs))
	}

	// Create test script
	scriptPath := filepath.Join(tmpDir, "list.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "list-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule multiple jobs
	_, err := scheduler.Schedule("list-test", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	_, err = scheduler.Schedule("list-test", false, make(map[string]string), "0 * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	// List jobs
	jobs = scheduler.ListScheduledJobs()
	if len(jobs) != 2 {
		t.Errorf("Expected 2 scheduled jobs, got %d", len(jobs))
	}
}

func TestUpdateSchedule(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "update.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "update-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job
	scheduledJob, err := scheduler.Schedule("update-test", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	originalSchedule := scheduledJob.Schedule
	originalEntryID := scheduledJob.EntryID

	// Update schedule
	newSchedule := "0 * * * *"
	if err := scheduler.UpdateSchedule(scheduledJob.ID, newSchedule); err != nil {
		t.Errorf("UpdateSchedule() failed: %v", err)
	}

	// Verify schedule was updated
	updatedJob, err := scheduler.GetScheduledJob(scheduledJob.ID)
	if err != nil {
		t.Fatalf("GetScheduledJob() failed: %v", err)
	}

	if updatedJob.Schedule != newSchedule {
		t.Errorf("Expected schedule '%s', got '%s'", newSchedule, updatedJob.Schedule)
	}

	if updatedJob.Schedule == originalSchedule {
		t.Error("Schedule was not updated")
	}

	if updatedJob.EntryID == originalEntryID {
		t.Error("EntryID was not updated")
	}
}

func TestUpdateSchedule_NotFound(t *testing.T) {
	scheduler, manager, _, _ := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	err := scheduler.UpdateSchedule("nonexistent", "* * * * *")
	if err == nil {
		t.Error("Expected error for nonexistent job, got nil")
	}
}

func TestUpdateSchedule_InvalidCron(t *testing.T) {
	scheduler, manager, scriptManager, tmpDir := setupTestScheduler(t)
	scheduler.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		scheduler.Stop(ctx)
		manager.Shutdown(ctx)
	}()

	// Create test script
	scriptPath := filepath.Join(tmpDir, "invalid.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0750); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	manager.config.Scripts = []config.Script{
		{
			Name:        "invalid-test",
			Path:        scriptPath,
			Interpreter: "bash",
			Parameters:  make(map[string]config.Parameter),
		},
	}

	if err := scriptManager.Reload(); err != nil {
		t.Fatalf("Failed to reload script manager: %v", err)
	}

	// Schedule job
	scheduledJob, err := scheduler.Schedule("invalid-test", false, make(map[string]string), "* * * * *")
	if err != nil {
		t.Fatalf("Schedule() failed: %v", err)
	}

	// Update with invalid schedule
	err = scheduler.UpdateSchedule(scheduledJob.ID, "invalid cron")
	if err == nil {
		t.Error("Expected error for invalid cron expression, got nil")
	}
}
