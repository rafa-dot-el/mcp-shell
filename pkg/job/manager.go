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

// Package job provides job queue, execution, and lifecycle management
package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

// JobStatus represents the current state of a job
//
//nolint:revive // JobStatus is an established API name, changing would break compatibility
type JobStatus string

//nolint:revive // Const names are self-documenting
const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
)

// Job represents a script execution job
type Job struct {
	// ID is a short UUID (8 hex characters) for token efficiency
	ID string

	// Name of the script or alias to execute
	Name string

	// IsAlias indicates if this is an alias (true) or script (false)
	IsAlias bool

	// Parameters for script execution
	Parameters map[string]string

	// Status of the job
	Status JobStatus

	// CreatedAt timestamp
	CreatedAt time.Time

	// StartedAt timestamp (nil if not started)
	StartedAt *time.Time

	// CompletedAt timestamp (nil if not completed)
	CompletedAt *time.Time

	// Duration of execution
	Duration time.Duration

	// ExitCode from the executed command
	ExitCode int

	// LogPath to the job's log file
	LogPath string

	// Error if execution failed
	Error error

	// ctx for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// Manager handles job queue and execution
type Manager struct {
	config   *config.Config
	executor *script.Executor

	// Job queue and tracking
	queue       []*Job
	running     map[string]*Job
	completed   []*Job
	queueMu     sync.Mutex
	runningMu   sync.RWMutex
	completedMu sync.Mutex

	// Execution control
	maxParallel int
	executing   chan struct{}
	trigger     chan struct{} // signals when queue should be processed

	// Shutdown control
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewManager creates a new job manager
func NewManager(cfg *config.Config, executor *script.Executor) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if executor == nil {
		return nil, fmt.Errorf("executor cannot be nil")
	}

	// Ensure log directory exists
	if err := os.MkdirAll(cfg.Execution.LogDirectory, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	m := &Manager{
		config:      cfg,
		executor:    executor,
		queue:       make([]*Job, 0),
		running:     make(map[string]*Job),
		completed:   make([]*Job, 0),
		maxParallel: cfg.Execution.MaxParallelJobs,
		executing:   make(chan struct{}, cfg.Execution.MaxParallelJobs),
		trigger:     make(chan struct{}, 1), // buffered to prevent blocking
		shutdown:    make(chan struct{}),
	}

	// Start queue processor
	m.wg.Add(1)
	go m.processQueue()

	return m, nil
}

// Shutdown gracefully shuts down the job manager
func (m *Manager) Shutdown(ctx context.Context) error {
	close(m.shutdown)

	// Wait for queue processor to finish or context cancellation
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// generateShortID generates a short 8-character hex ID
func generateShortID() string {
	fullUUID := uuid.New()
	// Take first 4 bytes (8 hex characters) for token efficiency
	return fmt.Sprintf("%08x", fullUUID.ID())
}

// Enqueue adds a job to the queue
func (m *Manager) Enqueue(name string, isAlias bool, parameters map[string]string) (*Job, error) {
	job := &Job{
		ID:         generateShortID(),
		Name:       name,
		IsAlias:    isAlias,
		Parameters: parameters,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		LogPath:    filepath.Join(m.config.Execution.LogDirectory, fmt.Sprintf("%s.log", generateShortID())),
	}

	job.ctx, job.cancel = context.WithCancel(context.Background())

	m.queueMu.Lock()
	m.queue = append(m.queue, job)
	m.queueMu.Unlock()

	// Trigger queue processing (non-blocking)
	select {
	case m.trigger <- struct{}{}:
	default:
	}

	return job, nil
}

// Execute runs a job immediately (bypasses queue)
func (m *Manager) Execute(ctx context.Context, name string, isAlias bool, parameters map[string]string) (*Job, error) {
	job := &Job{
		ID:         generateShortID(),
		Name:       name,
		IsAlias:    isAlias,
		Parameters: parameters,
		Status:     StatusRunning,
		CreatedAt:  time.Now(),
		LogPath:    filepath.Join(m.config.Execution.LogDirectory, fmt.Sprintf("%s.log", generateShortID())),
	}

	job.ctx, job.cancel = context.WithCancel(ctx)

	// Execute immediately
	m.executeJob(job)

	return job, nil
}

// processQueue continuously processes the job queue
func (m *Manager) processQueue() {
	defer m.wg.Done()

	for {
		select {
		case <-m.shutdown:
			return
		case <-m.trigger:
			m.processNextJob()
		}
	}
}

// processNextJob processes the next pending job if capacity is available
func (m *Manager) processNextJob() {
	// Check if we have capacity
	select {
	case m.executing <- struct{}{}:
		// Got capacity, proceed
	default:
		// No capacity, return
		return
	}

	// Get next pending job
	m.queueMu.Lock()
	var job *Job
	if len(m.queue) > 0 {
		job = m.queue[0]
		m.queue = m.queue[1:]
	}
	m.queueMu.Unlock()

	if job == nil {
		// No pending jobs, release capacity
		<-m.executing
		return
	}

	// Execute job in background
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer func() { <-m.executing }()

		m.executeJob(job)
	}()
}

// executeJob executes a single job
func (m *Manager) executeJob(job *Job) {
	now := time.Now()
	job.StartedAt = &now
	job.Status = StatusRunning

	// Add to running jobs
	m.runningMu.Lock()
	m.running[job.ID] = job
	m.runningMu.Unlock()

	// Create log file
	logFile, err := os.Create(job.LogPath)
	if err != nil {
		job.Error = fmt.Errorf("failed to create log file: %w", err)
		job.Status = StatusFailed
		m.completeJob(job)
		return
	}
	defer func() { _ = logFile.Close() }()

	// Execute script or alias
	var result *script.ExecutionResult
	req := &script.ExecutionRequest{
		Name:       job.Name,
		Parameters: job.Parameters,
	}

	if job.IsAlias {
		result, err = m.executor.ExecuteAlias(job.ctx, req)
	} else {
		result, err = m.executor.ExecuteScript(job.ctx, req)
	}

	// Write output to log file
	if result != nil {
		_, _ = fmt.Fprintf(logFile, "=== Job %s ===\n", job.ID)
		_, _ = fmt.Fprintf(logFile, "Name: %s\n", job.Name)
		_, _ = fmt.Fprintf(logFile, "Started: %s\n", job.StartedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(logFile, "\n=== STDOUT ===\n%s\n", result.Stdout)
		if result.Stderr != "" {
			_, _ = fmt.Fprintf(logFile, "\n=== STDERR ===\n%s\n", result.Stderr)
		}
		_, _ = fmt.Fprintf(logFile, "\n=== EXIT CODE ===\n%d\n", result.ExitCode)
		_, _ = fmt.Fprintf(logFile, "\n=== DURATION ===\n%s\n", result.Duration)

		job.Duration = result.Duration
		job.ExitCode = result.ExitCode
	}

	// Update job status
	//nolint:gocritic // if-else chain is clearer for error handling than switch
	if err != nil {
		job.Error = err
		job.Status = StatusFailed
		_, _ = fmt.Fprintf(logFile, "\n=== ERROR ===\n%s\n", err)
	} else if job.ctx.Err() == context.Canceled {
		job.Status = StatusCancelled
		_, _ = fmt.Fprintf(logFile, "\n=== CANCELLED ===\n")
	} else {
		job.Status = StatusCompleted
	}

	m.completeJob(job)
}

// completeJob marks a job as complete and moves it to completed list
func (m *Manager) completeJob(job *Job) {
	now := time.Now()
	job.CompletedAt = &now

	// Remove from running jobs
	m.runningMu.Lock()
	delete(m.running, job.ID)
	m.runningMu.Unlock()

	// Add to completed jobs
	m.completedMu.Lock()
	m.completed = append(m.completed, job)
	m.completedMu.Unlock()

	// Trigger queue processing for next job (non-blocking)
	select {
	case m.trigger <- struct{}{}:
	default:
	}
}

// GetJob retrieves a job by ID (checks queue, running, and completed)
func (m *Manager) GetJob(id string) (*Job, error) {
	// Check running jobs
	m.runningMu.RLock()
	if job, exists := m.running[id]; exists {
		m.runningMu.RUnlock()
		return job, nil
	}
	m.runningMu.RUnlock()

	// Check queue
	m.queueMu.Lock()
	for _, job := range m.queue {
		if job.ID == id {
			m.queueMu.Unlock()
			return job, nil
		}
	}
	m.queueMu.Unlock()

	// Check completed
	m.completedMu.Lock()
	for _, job := range m.completed {
		if job.ID == id {
			m.completedMu.Unlock()
			return job, nil
		}
	}
	m.completedMu.Unlock()

	return nil, fmt.Errorf("job '%s' not found", id)
}

// CancelJob cancels a running or pending job
func (m *Manager) CancelJob(id string) error {
	job, err := m.GetJob(id)
	if err != nil {
		return err
	}

	if job.Status == StatusCompleted || job.Status == StatusFailed || job.Status == StatusCancelled {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	// Cancel job context
	if job.cancel != nil {
		job.cancel()
	}

	// If pending, remove from queue
	if job.Status == StatusPending {
		m.queueMu.Lock()
		for i, queuedJob := range m.queue {
			if queuedJob.ID == id {
				m.queue = append(m.queue[:i], m.queue[i+1:]...)
				break
			}
		}
		m.queueMu.Unlock()

		job.Status = StatusCancelled
		now := time.Now()
		job.CompletedAt = &now

		m.completedMu.Lock()
		m.completed = append(m.completed, job)
		m.completedMu.Unlock()
	}

	return nil
}

// ListRunningJobs returns all currently running jobs
func (m *Manager) ListRunningJobs() []*Job {
	m.runningMu.RLock()
	defer m.runningMu.RUnlock()

	jobs := make([]*Job, 0, len(m.running))
	for _, job := range m.running {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListPendingJobs returns all jobs in the queue
func (m *Manager) ListPendingJobs() []*Job {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()

	jobs := make([]*Job, len(m.queue))
	copy(jobs, m.queue)
	return jobs
}

// ListCompletedJobs returns all completed jobs
func (m *Manager) ListCompletedJobs() []*Job {
	m.completedMu.Lock()
	defer m.completedMu.Unlock()

	jobs := make([]*Job, len(m.completed))
	copy(jobs, m.completed)
	return jobs
}

// CleanupJobs removes completed jobs and their log files
func (m *Manager) CleanupJobs(ids []string) error {
	m.completedMu.Lock()
	defer m.completedMu.Unlock()

	if len(ids) == 0 {
		// Clean up all completed jobs
		for _, job := range m.completed {
			if err := os.Remove(job.LogPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove log file for job %s: %w", job.ID, err)
			}
		}
		m.completed = make([]*Job, 0)
		return nil
	}

	// Clean up specific jobs
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	remaining := make([]*Job, 0)
	for _, job := range m.completed {
		if idSet[job.ID] {
			if err := os.Remove(job.LogPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove log file for job %s: %w", job.ID, err)
			}
		} else {
			remaining = append(remaining, job)
		}
	}

	m.completed = remaining
	return nil
}
