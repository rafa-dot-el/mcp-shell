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

// Package mcp implements the Model Context Protocol server using mark3labs/mcp-go
package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/job"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

// Server implements the MCP server using mark3labs/mcp-go
type Server struct {
	config    *config.Config
	manager   *script.Manager
	executor  *script.Executor
	jobMgr    *job.Manager
	scheduler *job.Scheduler
	mcpServer *server.MCPServer
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

	// Create mcp-go server
	s.mcpServer = server.NewMCPServer(
		cfg.MCP.Name,
		cfg.MCP.Version,
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

	// Register all tools
	s.registerTools()

	// Register all resources
	s.registerResources()

	// Register all prompts
	s.registerPrompts()

	return s, nil
}

// Serve starts the MCP server using stdio transport
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
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

// registerTools registers all MCP tools with the server
func (s *Server) registerTools() {
	// Script execution tools
	s.mcpServer.AddTool(
		mcp.NewTool("execute_script",
			mcp.WithDescription("Execute a script immediately with parameters"),
			mcp.WithString("name",
				mcp.Description("Name of the script to execute"),
				mcp.Required()),
			mcp.WithObject("parameters",
				mcp.Description("Parameters for the script (key-value pairs)")),
		),
		s.handleExecuteScript,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("enqueue_script",
			mcp.WithDescription("Enqueue a script for background execution"),
			mcp.WithString("name",
				mcp.Description("Name of the script to enqueue"),
				mcp.Required()),
			mcp.WithObject("parameters",
				mcp.Description("Parameters for the script (key-value pairs)")),
		),
		s.handleEnqueueScript,
	)

	// Alias execution tools
	s.mcpServer.AddTool(
		mcp.NewTool("execute_alias",
			mcp.WithDescription("Execute an alias command immediately"),
			mcp.WithString("name",
				mcp.Description("Name of the alias to execute"),
				mcp.Required()),
		),
		s.handleExecuteAlias,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("enqueue_alias",
			mcp.WithDescription("Enqueue an alias for background execution"),
			mcp.WithString("name",
				mcp.Description("Name of the alias to enqueue"),
				mcp.Required()),
		),
		s.handleEnqueueAlias,
	)

	// Job management tools
	s.mcpServer.AddTool(
		mcp.NewTool("list_jobs",
			mcp.WithDescription("List jobs (running, pending, or completed)"),
			mcp.WithString("status",
				mcp.Description("Job status filter"),
				mcp.Enum("running", "pending", "completed")),
		),
		s.handleListJobs,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_job",
			mcp.WithDescription("Get details of a specific job"),
			mcp.WithString("job_id",
				mcp.Description("Job ID to retrieve"),
				mcp.Required()),
		),
		s.handleGetJob,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("cancel_job",
			mcp.WithDescription("Cancel a running or pending job"),
			mcp.WithString("job_id",
				mcp.Description("Job ID to cancel"),
				mcp.Required()),
		),
		s.handleCancelJob,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("cleanup_jobs",
			mcp.WithDescription("Remove completed jobs and their logs"),
			mcp.WithArray("job_ids",
				mcp.Description("Specific job IDs to clean up (empty for all)")),
		),
		s.handleCleanupJobs,
	)

	// Log management tools
	s.mcpServer.AddTool(
		mcp.NewTool("tail_log",
			mcp.WithDescription("Get the last N lines from a job's log"),
			mcp.WithString("job_id",
				mcp.Description("Job ID"),
				mcp.Required()),
			mcp.WithNumber("lines",
				mcp.Description("Number of lines to tail")),
			mcp.WithString("filter",
				mcp.Description("Regex pattern to filter log lines (optional)")),
		),
		s.handleTailLog,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("read_log",
			mcp.WithDescription("Read the full log file for a job"),
			mcp.WithString("job_id",
				mcp.Description("Job ID"),
				mcp.Required()),
		),
		s.handleReadLog,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("search_log",
			mcp.WithDescription("Search for a pattern in a job's log"),
			mcp.WithString("job_id",
				mcp.Description("Job ID"),
				mcp.Required()),
			mcp.WithString("pattern",
				mcp.Description("Regex pattern to search for"),
				mcp.Required()),
		),
		s.handleSearchLog,
	)

	// Scheduling tools
	s.mcpServer.AddTool(
		mcp.NewTool("schedule_script",
			mcp.WithDescription("Schedule a script with cron expression or one-time"),
			mcp.WithString("name",
				mcp.Description("Script name"),
				mcp.Required()),
			mcp.WithString("schedule",
				mcp.Description("Cron expression (e.g., '*/5 * * * *') or timestamp"),
				mcp.Required()),
			mcp.WithObject("parameters",
				mcp.Description("Script parameters (key-value pairs)")),
		),
		s.handleScheduleScript,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("schedule_alias",
			mcp.WithDescription("Schedule an alias with cron expression or one-time"),
			mcp.WithString("name",
				mcp.Description("Alias name"),
				mcp.Required()),
			mcp.WithString("schedule",
				mcp.Description("Cron expression (e.g., '*/5 * * * *') or timestamp"),
				mcp.Required()),
		),
		s.handleScheduleAlias,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("list_scheduled",
			mcp.WithDescription("List all scheduled jobs"),
		),
		s.handleListScheduled,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("unschedule_job",
			mcp.WithDescription("Remove a scheduled job"),
			mcp.WithString("schedule_id",
				mcp.Description("Scheduled job ID"),
				mcp.Required()),
		),
		s.handleUnscheduleJob,
	)

	// Script management tools
	s.mcpServer.AddTool(
		mcp.NewTool("list_scripts",
			mcp.WithDescription("List all available scripts"),
		),
		s.handleListScripts,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("list_aliases",
			mcp.WithDescription("List all available aliases"),
		),
		s.handleListAliases,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_script_info",
			mcp.WithDescription("Get detailed information about a script"),
			mcp.WithString("name",
				mcp.Description("Script name"),
				mcp.Required()),
		),
		s.handleGetScriptInfo,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("create_script",
			mcp.WithDescription("Create a new script (if allowed by security settings)"),
			mcp.WithString("name",
				mcp.Description("Script name"),
				mcp.Required()),
			mcp.WithString("content",
				mcp.Description("Script content"),
				mcp.Required()),
			mcp.WithString("interpreter",
				mcp.Description("Interpreter (bash, python3, perl, etc.)"),
				mcp.Required()),
		),
		s.handleCreateScript,
	)
}

// registerResources registers all MCP resources with the server
func (s *Server) registerResources() {
	// Register resource templates
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"script://{name}",
			"Script Resource",
			mcp.WithTemplateDescription("Access a specific script by name"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		s.handleReadScriptResource,
	)

	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"alias://{name}",
			"Alias Resource",
			mcp.WithTemplateDescription("Access a specific alias by name"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		s.handleReadAliasResource,
	)

	// Register static resources for listing
	s.updateResourceList()
}

// updateResourceList updates the list of available resources
func (s *Server) updateResourceList() {
	// Add script resources
	scripts := s.manager.ListScripts()
	for _, script := range scripts {
		s.mcpServer.AddResource(
			mcp.NewResource(
				fmt.Sprintf("script://%s", script.Config.Name),
				script.Config.Name,
				mcp.WithResourceDescription(script.Config.Description),
				mcp.WithMIMEType("text/markdown"),
			),
			s.handleReadScriptResource,
		)
	}

	// Add alias resources
	aliases := s.manager.ListAliases()
	for _, alias := range aliases {
		s.mcpServer.AddResource(
			mcp.NewResource(
				fmt.Sprintf("alias://%s", alias.Name),
				alias.Name,
				mcp.WithResourceDescription(alias.Description),
				mcp.WithMIMEType("text/markdown"),
			),
			s.handleReadAliasResource,
		)
	}
}

// registerPrompts registers all MCP prompts with the server
func (s *Server) registerPrompts() {
	// Execute script prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("execute-script",
			mcp.WithPromptDescription("Execute a script with parameters"),
			mcp.WithArgument("script_name",
				mcp.ArgumentDescription("Name of the script to execute"),
				mcp.RequiredArgument()),
			mcp.WithArgument("parameters",
				mcp.ArgumentDescription("JSON object of parameters (optional)")),
		),
		s.handleExecuteScriptPrompt,
	)

	// Execute alias prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("execute-alias",
			mcp.WithPromptDescription("Execute an alias command"),
			mcp.WithArgument("alias_name",
				mcp.ArgumentDescription("Name of the alias to execute"),
				mcp.RequiredArgument()),
		),
		s.handleExecuteAliasPrompt,
	)

	// Schedule script prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("schedule-script",
			mcp.WithPromptDescription("Schedule a script to run at a specific time or with cron"),
			mcp.WithArgument("script_name",
				mcp.ArgumentDescription("Name of the script to schedule"),
				mcp.RequiredArgument()),
			mcp.WithArgument("schedule",
				mcp.ArgumentDescription("Cron expression or timestamp"),
				mcp.RequiredArgument()),
			mcp.WithArgument("parameters",
				mcp.ArgumentDescription("JSON object of parameters (optional)")),
		),
		s.handleScheduleScriptPrompt,
	)

	// Monitor job prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("monitor-job",
			mcp.WithPromptDescription("Monitor a running job and view its logs"),
			mcp.WithArgument("job_id",
				mcp.ArgumentDescription("ID of the job to monitor"),
				mcp.RequiredArgument()),
		),
		s.handleMonitorJobPrompt,
	)

	// List available scripts prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("list-available-scripts",
			mcp.WithPromptDescription("List all available scripts and aliases"),
		),
		s.handleListScriptsPrompt,
	)

	// Create script prompt
	s.mcpServer.AddPrompt(
		mcp.NewPrompt("create-script",
			mcp.WithPromptDescription("Create a new script (if allowed by security settings)"),
			mcp.WithArgument("name",
				mcp.ArgumentDescription("Name for the new script"),
				mcp.RequiredArgument()),
			mcp.WithArgument("interpreter",
				mcp.ArgumentDescription("Interpreter to use (bash, python3, etc.)"),
				mcp.RequiredArgument()),
			mcp.WithArgument("content",
				mcp.ArgumentDescription("Script content"),
				mcp.RequiredArgument()),
		),
		s.handleCreateScriptPrompt,
	)
}
