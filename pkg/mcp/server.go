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

// Package mcp implements the Model Context Protocol server
package mcp

import (
	"context"
	"fmt"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/job"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

// Server implements the MCP server
type Server struct {
	config    *config.Config
	manager   *script.Manager
	executor  *script.Executor
	jobMgr    *job.Manager
	scheduler *job.Scheduler

	// MCP handlers
	resources *ResourceHandler
	prompts   *PromptHandler
	tools     *ToolHandler
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create script manager
	manager, err := script.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create script manager: %w", err)
	}

	// Create executor
	executor := script.NewExecutor(manager, cfg)

	// Create job manager
	jobMgr, err := job.NewManager(cfg, executor)
	if err != nil {
		return nil, fmt.Errorf("failed to create job manager: %w", err)
	}

	// Create scheduler
	scheduler := job.NewScheduler(jobMgr)
	scheduler.Start()

	s := &Server{
		config:    cfg,
		manager:   manager,
		executor:  executor,
		jobMgr:    jobMgr,
		scheduler: scheduler,
	}

	// Initialize handlers
	s.resources = NewResourceHandler(s)
	s.prompts = NewPromptHandler(s)
	s.tools = NewToolHandler(s)

	return s, nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop scheduler
	if err := s.scheduler.Stop(ctx); err != nil {
		return fmt.Errorf("scheduler shutdown failed: %w", err)
	}

	// Stop job manager
	if err := s.jobMgr.Shutdown(ctx); err != nil {
		return fmt.Errorf("job manager shutdown failed: %w", err)
	}

	return nil
}

// GetServerInfo returns server information
func (s *Server) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:    s.config.MCP.Name,
		Version: s.config.MCP.Version,
	}
}

// ServerInfo contains server metadata
type ServerInfo struct {
	Name    string
	Version string
}
