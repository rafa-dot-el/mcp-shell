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
	"encoding/json"
	"fmt"
	"time"

	"github.com/rafa-dot-el/mcp-shell/pkg/job"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
)

// ToolHandler handles MCP tool requests
type ToolHandler struct {
	server *Server
}

// NewToolHandler creates a new tool handler
func NewToolHandler(server *Server) *ToolHandler {
	return &ToolHandler{
		server: server,
	}
}

// Tool represents an MCP tool
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`
}

// ToolInputSchema defines the JSON schema for tool inputs
type ToolInputSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]ToolPropertySchema `json:"properties"`
	Required   []string                      `json:"required,omitempty"`
}

// ToolPropertySchema defines a property in the input schema
type ToolPropertySchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content interface{} `json:"content"`
	IsError bool        `json:"isError"`
}

// ListTools returns all available tools
func (h *ToolHandler) ListTools() []Tool {
	return []Tool{
		// Script execution
		{
			Name:        "execute_script",
			Description: "Execute a script immediately with parameters",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Name of the script to execute",
					},
					"parameters": {
						Type:        "object",
						Description: "Parameters for the script (key-value pairs)",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "enqueue_script",
			Description: "Enqueue a script for background execution",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Name of the script to enqueue",
					},
					"parameters": {
						Type:        "object",
						Description: "Parameters for the script (key-value pairs)",
					},
				},
				Required: []string{"name"},
			},
		},

		// Alias execution
		{
			Name:        "execute_alias",
			Description: "Execute an alias command immediately",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Name of the alias to execute",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "enqueue_alias",
			Description: "Enqueue an alias for background execution",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Name of the alias to enqueue",
					},
				},
				Required: []string{"name"},
			},
		},

		// Job management
		{
			Name:        "list_jobs",
			Description: "List jobs (running, pending, or completed)",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"status": {
						Type:        "string",
						Description: "Job status filter",
						Enum:        []string{"running", "pending", "completed"},
						Default:     "running",
					},
				},
			},
		},
		{
			Name:        "get_job",
			Description: "Get details of a specific job",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Job ID to retrieve",
					},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "cancel_job",
			Description: "Cancel a running or pending job",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Job ID to cancel",
					},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "cleanup_jobs",
			Description: "Remove completed jobs and their logs",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_ids": {
						Type:        "array",
						Description: "Specific job IDs to clean up (empty for all)",
					},
				},
			},
		},

		// Log management
		{
			Name:        "tail_log",
			Description: "Get the last N lines from a job's log",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Job ID",
					},
					"lines": {
						Type:        "integer",
						Description: "Number of lines to tail",
						Default:     5,
					},
					"filter": {
						Type:        "string",
						Description: "Regex pattern to filter log lines (optional)",
					},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "read_log",
			Description: "Read the full log file for a job",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Job ID",
					},
				},
				Required: []string{"job_id"},
			},
		},
		{
			Name:        "search_log",
			Description: "Search for a pattern in a job's log",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Job ID",
					},
					"pattern": {
						Type:        "string",
						Description: "Regex pattern to search for",
					},
				},
				Required: []string{"job_id", "pattern"},
			},
		},

		// Scheduling
		{
			Name:        "schedule_script",
			Description: "Schedule a script with cron expression or one-time",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Script name",
					},
					"schedule": {
						Type:        "string",
						Description: "Cron expression (e.g., '*/5 * * * *') or timestamp",
					},
					"parameters": {
						Type:        "object",
						Description: "Script parameters (key-value pairs)",
					},
				},
				Required: []string{"name", "schedule"},
			},
		},
		{
			Name:        "schedule_alias",
			Description: "Schedule an alias with cron expression or one-time",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Alias name",
					},
					"schedule": {
						Type:        "string",
						Description: "Cron expression (e.g., '*/5 * * * *') or timestamp",
					},
				},
				Required: []string{"name", "schedule"},
			},
		},
		{
			Name:        "list_scheduled",
			Description: "List all scheduled jobs",
			InputSchema: ToolInputSchema{
				Type:       "object",
				Properties: map[string]ToolPropertySchema{},
			},
		},
		{
			Name:        "unschedule_job",
			Description: "Remove a scheduled job",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"schedule_id": {
						Type:        "string",
						Description: "Scheduled job ID",
					},
				},
				Required: []string{"schedule_id"},
			},
		},

		// Script management
		{
			Name:        "list_scripts",
			Description: "List all available scripts",
			InputSchema: ToolInputSchema{
				Type:       "object",
				Properties: map[string]ToolPropertySchema{},
			},
		},
		{
			Name:        "list_aliases",
			Description: "List all available aliases",
			InputSchema: ToolInputSchema{
				Type:       "object",
				Properties: map[string]ToolPropertySchema{},
			},
		},
		{
			Name:        "get_script_info",
			Description: "Get detailed information about a script",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Script name",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "create_script",
			Description: "Create a new script (if allowed by security settings)",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name": {
						Type:        "string",
						Description: "Script name",
					},
					"content": {
						Type:        "string",
						Description: "Script content",
					},
					"interpreter": {
						Type:        "string",
						Description: "Interpreter (bash, python3, perl, etc.)",
					},
				},
				Required: []string{"name", "content", "interpreter"},
			},
		},
	}
}

// CallTool executes a tool with given arguments
func (h *ToolHandler) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	// Script execution
	case "execute_script":
		return h.executeScript(ctx, args)
	case "enqueue_script":
		return h.enqueueScript(args)

	// Alias execution
	case "execute_alias":
		return h.executeAlias(ctx, args)
	case "enqueue_alias":
		return h.enqueueAlias(args)

	// Job management
	case "list_jobs":
		return h.listJobs(args)
	case "get_job":
		return h.getJob(args)
	case "cancel_job":
		return h.cancelJob(args)
	case "cleanup_jobs":
		return h.cleanupJobs(args)

	// Log management
	case "tail_log":
		return h.tailLog(args)
	case "read_log":
		return h.readLog(args)
	case "search_log":
		return h.searchLog(args)

	// Scheduling
	case "schedule_script":
		return h.scheduleScript(args)
	case "schedule_alias":
		return h.scheduleAlias(args)
	case "list_scheduled":
		return h.listScheduled(args)
	case "unschedule_job":
		return h.unscheduleJob(args)

	// Script management
	case "list_scripts":
		return h.listScripts(args)
	case "list_aliases":
		return h.listAliases(args)
	case "get_script_info":
		return h.getScriptInfo(args)
	case "create_script":
		return h.createScript(args)

	default:
		return &ToolResult{
			Content: fmt.Sprintf("Unknown tool: %s", name),
			IsError: true,
		}, nil
	}
}

// Helper functions for tool implementations

func (h *ToolHandler) executeScript(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)
	parameters := make(map[string]string)

	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	job, err := h.server.jobMgr.Execute(ctx, name, false, parameters)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"job_id":    job.ID,
			"status":    job.Status,
			"exit_code": job.ExitCode,
			"duration":  job.Duration.String(),
			"log_path":  job.LogPath,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) enqueueScript(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)
	parameters := make(map[string]string)

	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	job, err := h.server.jobMgr.Enqueue(name, false, parameters)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"job_id":     job.ID,
			"status":     job.Status,
			"created_at": job.CreatedAt,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) executeAlias(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)

	job, err := h.server.jobMgr.Execute(ctx, name, true, nil)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"job_id":    job.ID,
			"status":    job.Status,
			"exit_code": job.ExitCode,
			"duration":  job.Duration.String(),
			"log_path":  job.LogPath,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) enqueueAlias(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)

	job, err := h.server.jobMgr.Enqueue(name, true, nil)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"job_id":     job.ID,
			"status":     job.Status,
			"created_at": job.CreatedAt,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) listJobs(args map[string]interface{}) (*ToolResult, error) {
	status := "running"
	if s, ok := args["status"].(string); ok {
		status = s
	}

	var jobs []*job.Job
	switch status {
	case "running":
		jobs = h.server.jobMgr.ListRunningJobs()
	case "pending":
		jobs = h.server.jobMgr.ListPendingJobs()
	case "completed":
		jobs = h.server.jobMgr.ListCompletedJobs()
	default:
		return &ToolResult{
			Content: map[string]interface{}{
				"error": fmt.Sprintf("invalid status: %s", status),
			},
			IsError: true,
		}, nil
	}

	jobList := make([]map[string]interface{}, len(jobs))
	for i, j := range jobs {
		jobList[i] = map[string]interface{}{
			"id":         j.ID,
			"name":       j.Name,
			"is_alias":   j.IsAlias,
			"status":     j.Status,
			"created_at": j.CreatedAt,
		}
		if j.StartedAt != nil {
			jobList[i]["started_at"] = j.StartedAt
		}
		if j.CompletedAt != nil {
			jobList[i]["completed_at"] = j.CompletedAt
			jobList[i]["duration"] = j.Duration.String()
			jobList[i]["exit_code"] = j.ExitCode
		}
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"jobs":  jobList,
			"count": len(jobs),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) getJob(args map[string]interface{}) (*ToolResult, error) {
	jobID := args["job_id"].(string)

	j, err := h.server.jobMgr.GetJob(jobID)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	result := map[string]interface{}{
		"id":         j.ID,
		"name":       j.Name,
		"is_alias":   j.IsAlias,
		"status":     j.Status,
		"created_at": j.CreatedAt,
		"log_path":   j.LogPath,
	}

	if j.StartedAt != nil {
		result["started_at"] = j.StartedAt
	}
	if j.CompletedAt != nil {
		result["completed_at"] = j.CompletedAt
		result["duration"] = j.Duration.String()
		result["exit_code"] = j.ExitCode
	}
	if j.Error != nil {
		result["error"] = j.Error.Error()
	}
	if len(j.Parameters) > 0 {
		result["parameters"] = j.Parameters
	}

	return &ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

func (h *ToolHandler) cancelJob(args map[string]interface{}) (*ToolResult, error) {
	jobID := args["job_id"].(string)

	err := h.server.jobMgr.CancelJob(jobID)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"message": fmt.Sprintf("Job %s cancelled", jobID),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) cleanupJobs(args map[string]interface{}) (*ToolResult, error) {
	var jobIDs []string

	if ids, ok := args["job_ids"].([]interface{}); ok {
		for _, id := range ids {
			jobIDs = append(jobIDs, id.(string))
		}
	}

	err := h.server.jobMgr.CleanupJobs(jobIDs)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	message := "All completed jobs cleaned up"
	if len(jobIDs) > 0 {
		message = fmt.Sprintf("Cleaned up %d job(s)", len(jobIDs))
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"message": message,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) tailLog(args map[string]interface{}) (*ToolResult, error) {
	jobID := args["job_id"].(string)
	lines := 5
	if l, ok := args["lines"].(float64); ok {
		lines = int(l)
	}

	filter := ""
	if f, ok := args["filter"].(string); ok {
		filter = f
	}

	logLines, err := h.server.jobMgr.TailLog(jobID, job.TailOptions{
		Lines:  lines,
		Filter: filter,
	})

	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"lines": logLines,
			"count": len(logLines),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) readLog(args map[string]interface{}) (*ToolResult, error) {
	jobID := args["job_id"].(string)

	content, err := h.server.jobMgr.ReadFullLog(jobID)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"content": content,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) searchLog(args map[string]interface{}) (*ToolResult, error) {
	jobID := args["job_id"].(string)
	pattern := args["pattern"].(string)

	matches, err := h.server.jobMgr.SearchLog(jobID, pattern)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"matches": matches,
			"count":   len(matches),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) scheduleScript(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)
	schedule := args["schedule"].(string)
	parameters := make(map[string]string)

	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	// Try to parse as timestamp first, then as cron
	var scheduledJob *job.ScheduledJob
	var err error

	if t, parseErr := time.Parse(time.RFC3339, schedule); parseErr == nil {
		scheduledJob, err = h.server.scheduler.ScheduleOnce(name, false, parameters, t)
	} else {
		scheduledJob, err = h.server.scheduler.Schedule(name, false, parameters, schedule)
	}

	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"schedule_id": scheduledJob.ID,
			"name":        scheduledJob.Name,
			"schedule":    scheduledJob.Schedule,
			"next_run":    scheduledJob.NextRun,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) scheduleAlias(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)
	schedule := args["schedule"].(string)

	// Try to parse as timestamp first, then as cron
	var scheduledJob *job.ScheduledJob
	var err error

	if t, parseErr := time.Parse(time.RFC3339, schedule); parseErr == nil {
		scheduledJob, err = h.server.scheduler.ScheduleOnce(name, true, nil, t)
	} else {
		scheduledJob, err = h.server.scheduler.Schedule(name, true, nil, schedule)
	}

	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"schedule_id": scheduledJob.ID,
			"name":        scheduledJob.Name,
			"schedule":    scheduledJob.Schedule,
			"next_run":    scheduledJob.NextRun,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) listScheduled(args map[string]interface{}) (*ToolResult, error) {
	scheduled := h.server.scheduler.ListScheduledJobs()

	scheduleList := make([]map[string]interface{}, len(scheduled))
	for i, s := range scheduled {
		scheduleList[i] = map[string]interface{}{
			"id":       s.ID,
			"name":     s.Name,
			"is_alias": s.IsAlias,
			"schedule": s.Schedule,
			"next_run": s.NextRun,
		}
		if s.LastRun != nil {
			scheduleList[i]["last_run"] = s.LastRun
		}
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"scheduled": scheduleList,
			"count":     len(scheduled),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) unscheduleJob(args map[string]interface{}) (*ToolResult, error) {
	scheduleID := args["schedule_id"].(string)

	err := h.server.scheduler.Unschedule(scheduleID)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"message": fmt.Sprintf("Scheduled job %s removed", scheduleID),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) listScripts(args map[string]interface{}) (*ToolResult, error) {
	scripts := h.server.manager.ListScripts()

	scriptList := make([]map[string]interface{}, len(scripts))
	for i, s := range scripts {
		scriptList[i] = map[string]interface{}{
			"name":        s.Config.Name,
			"description": s.Config.Description,
			"path":        s.AbsolutePath,
			"interpreter": s.Config.Interpreter,
			"source":      s.Source,
			"parameters":  s.Config.Parameters,
		}
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"scripts": scriptList,
			"count":   len(scripts),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) listAliases(args map[string]interface{}) (*ToolResult, error) {
	aliases := h.server.manager.ListAliases()

	aliasList := make([]map[string]interface{}, len(aliases))
	for i, a := range aliases {
		aliasList[i] = map[string]interface{}{
			"name":        a.Name,
			"description": a.Description,
			"command":     a.Command,
		}
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"aliases": aliasList,
			"count":   len(aliases),
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) getScriptInfo(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)

	script, err := h.server.manager.GetScript(name)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: map[string]interface{}{
			"name":        script.Config.Name,
			"description": script.Config.Description,
			"path":        script.AbsolutePath,
			"interpreter": script.Config.Interpreter,
			"source":      script.Source,
			"executable":  script.IsExecutable,
			"parameters":  script.Config.Parameters,
		},
		IsError: false,
	}, nil
}

func (h *ToolHandler) createScript(args map[string]interface{}) (*ToolResult, error) {
	name := args["name"].(string)
	content := args["content"].(string)
	interpreter := args["interpreter"].(string)

	err := h.server.manager.CreateScript(name, content, interpreter)
	if err != nil {
		return &ToolResult{
			Content: map[string]interface{}{
				"error": err.Error(),
			},
			IsError: true,
		}, nil
	}

	// Get the created script info
	script, _ := h.server.manager.GetScript(name)

	return &ToolResult{
		Content: map[string]interface{}{
			"message": fmt.Sprintf("Script %s created", name),
			"path":    script.AbsolutePath,
		},
		IsError: false,
	}, nil
}

// SerializeToolResult serializes a tool result to JSON
func SerializeToolResult(result *ToolResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
