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
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID         string
	Name       string
	IsAlias    bool
	Parameters map[string]string
	Schedule   string // Cron expression
	NextRun    time.Time
	LastRun    *time.Time
	EntryID    cron.EntryID
}

// Scheduler handles job scheduling with cron support
type Scheduler struct {
	manager   *Manager
	cron      *cron.Cron
	scheduled map[string]*ScheduledJob
	mu        sync.RWMutex
}

// NewScheduler creates a new job scheduler
func NewScheduler(manager *Manager) *Scheduler {
	return &Scheduler{
		manager:   manager,
		cron:      cron.New(),
		scheduled: make(map[string]*ScheduledJob),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop halts the scheduler and waits for running jobs
func (s *Scheduler) Stop(ctx context.Context) error {
	stopCtx := s.cron.Stop()

	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return fmt.Errorf("scheduler stop timeout: %w", ctx.Err())
	}
}

// Schedule adds a job to the scheduler with a cron expression
func (s *Scheduler) Schedule(name string, isAlias bool, parameters map[string]string, schedule string) (*ScheduledJob, error) {
	id := generateShortID()

	// Create scheduled job
	scheduledJob := &ScheduledJob{
		ID:         id,
		Name:       name,
		IsAlias:    isAlias,
		Parameters: parameters,
		Schedule:   schedule,
	}

	// Add to cron
	entryID, err := s.cron.AddFunc(schedule, func() {
		s.executeScheduledJob(scheduledJob)
	})

	if err != nil {
		return nil, fmt.Errorf("invalid cron schedule: %w", err)
	}

	scheduledJob.EntryID = entryID

	// Get next run time
	entry := s.cron.Entry(entryID)
	scheduledJob.NextRun = entry.Next

	// Store scheduled job
	s.mu.Lock()
	s.scheduled[id] = scheduledJob
	s.mu.Unlock()

	return scheduledJob, nil
}

// ScheduleOnce adds a job to run once at a specific time
func (s *Scheduler) ScheduleOnce(name string, isAlias bool, parameters map[string]string, runAt time.Time) (*ScheduledJob, error) {
	id := generateShortID()

	// Create scheduled job
	scheduledJob := &ScheduledJob{
		ID:         id,
		Name:       name,
		IsAlias:    isAlias,
		Parameters: parameters,
		Schedule:   fmt.Sprintf("once at %s", runAt.Format(time.RFC3339)),
		NextRun:    runAt,
	}

	// Calculate delay until execution
	delay := time.Until(runAt)
	if delay < 0 {
		return nil, fmt.Errorf("scheduled time is in the past")
	}

	// Store scheduled job
	s.mu.Lock()
	s.scheduled[id] = scheduledJob
	s.mu.Unlock()

	// Schedule execution
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		<-timer.C

		s.executeScheduledJob(scheduledJob)

		// Remove from scheduled jobs after execution
		s.mu.Lock()
		delete(s.scheduled, id)
		s.mu.Unlock()
	}()

	return scheduledJob, nil
}

// executeScheduledJob executes a scheduled job
func (s *Scheduler) executeScheduledJob(scheduledJob *ScheduledJob) {
	// Update last run time
	now := time.Now()
	scheduledJob.LastRun = &now

	// Update next run time for recurring jobs
	if scheduledJob.EntryID != 0 {
		entry := s.cron.Entry(scheduledJob.EntryID)
		scheduledJob.NextRun = entry.Next
	}

	// Enqueue job for execution
	_, _ = s.manager.Enqueue(scheduledJob.Name, scheduledJob.IsAlias, scheduledJob.Parameters)
}

// Unschedule removes a scheduled job
func (s *Scheduler) Unschedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledJob, exists := s.scheduled[id]
	if !exists {
		return fmt.Errorf("scheduled job '%s' not found", id)
	}

	// Remove from cron if it has an entry ID
	if scheduledJob.EntryID != 0 {
		s.cron.Remove(scheduledJob.EntryID)
	}

	// Remove from scheduled jobs
	delete(s.scheduled, id)

	return nil
}

// GetScheduledJob retrieves a scheduled job by ID
func (s *Scheduler) GetScheduledJob(id string) (*ScheduledJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.scheduled[id]
	if !exists {
		return nil, fmt.Errorf("scheduled job '%s' not found", id)
	}

	return job, nil
}

// ListScheduledJobs returns all scheduled jobs
func (s *Scheduler) ListScheduledJobs() []*ScheduledJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*ScheduledJob, 0, len(s.scheduled))
	for _, job := range s.scheduled {
		jobs = append(jobs, job)
	}

	return jobs
}

// UpdateSchedule updates the schedule of an existing scheduled job
func (s *Scheduler) UpdateSchedule(id string, newSchedule string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledJob, exists := s.scheduled[id]
	if !exists {
		return fmt.Errorf("scheduled job '%s' not found", id)
	}

	// Remove old cron entry
	if scheduledJob.EntryID != 0 {
		s.cron.Remove(scheduledJob.EntryID)
	}

	// Add new cron entry
	entryID, err := s.cron.AddFunc(newSchedule, func() {
		s.executeScheduledJob(scheduledJob)
	})

	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	// Update scheduled job
	scheduledJob.Schedule = newSchedule
	scheduledJob.EntryID = entryID

	// Get next run time
	entry := s.cron.Entry(entryID)
	scheduledJob.NextRun = entry.Next

	return nil
}
