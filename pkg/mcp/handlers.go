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
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rafa-dot-el/mcp-shell/pkg/job"
)

// Tool Handlers

func (s *Server) handleExecuteScript(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	parameters := make(map[string]string)
	args := request.GetArguments()
	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	j, err := s.jobMgr.Execute(ctx, name, false, parameters)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"job_id":    j.ID,
		"status":    j.Status,
		"exit_code": j.ExitCode,
		"duration":  j.Duration.String(),
		"log_path":  j.LogPath,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleEnqueueScript(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	parameters := make(map[string]string)
	args := request.GetArguments()
	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	j, err := s.jobMgr.Enqueue(name, false, parameters)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("enqueue failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"job_id":     j.ID,
		"status":     j.Status,
		"created_at": j.CreatedAt,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleExecuteAlias(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	j, err := s.jobMgr.Execute(ctx, name, true, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"job_id":    j.ID,
		"status":    j.Status,
		"exit_code": j.ExitCode,
		"duration":  j.Duration.String(),
		"log_path":  j.LogPath,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleEnqueueAlias(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	j, err := s.jobMgr.Enqueue(name, true, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("enqueue failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"job_id":     j.ID,
		"status":     j.Status,
		"created_at": j.CreatedAt,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleListJobs(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := request.GetString("status", "running")

	var jobs []*job.Job
	switch status {
	case "running":
		jobs = s.jobMgr.ListRunningJobs()
	case "pending":
		jobs = s.jobMgr.ListPendingJobs()
	case "completed":
		jobs = s.jobMgr.ListCompletedJobs()
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s", status)), nil
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

	result := map[string]interface{}{
		"jobs":  jobList,
		"count": len(jobs),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleGetJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := request.GetString("job_id", "")
	if jobID == "" {
		return mcp.NewToolResultError("job_id parameter is required"), nil
	}

	j, err := s.jobMgr.GetJob(jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("job not found: %v", err)), nil
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

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleCancelJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := request.GetString("job_id", "")
	if jobID == "" {
		return mcp.NewToolResultError("job_id parameter is required"), nil
	}

	err := s.jobMgr.CancelJob(jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("cancel failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"message": fmt.Sprintf("Job %s cancelled", jobID),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleCleanupJobs(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var jobIDs []string

	args := request.GetArguments()
	if ids, ok := args["job_ids"].([]interface{}); ok {
		for _, id := range ids {
			if idStr, ok := id.(string); ok {
				jobIDs = append(jobIDs, idStr)
			}
		}
	}

	err := s.jobMgr.CleanupJobs(jobIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("cleanup failed: %v", err)), nil
	}

	message := "All completed jobs cleaned up"
	if len(jobIDs) > 0 {
		message = fmt.Sprintf("Cleaned up %d job(s)", len(jobIDs))
	}

	result := map[string]interface{}{
		"message": message,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleTailLog(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := request.GetString("job_id", "")
	if jobID == "" {
		return mcp.NewToolResultError("job_id parameter is required"), nil
	}

	lines := request.GetInt("lines", 5)
	filter := request.GetString("filter", "")

	logLines, err := s.jobMgr.TailLog(jobID, job.TailOptions{
		Lines:  lines,
		Filter: filter,
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("tail failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"lines": logLines,
		"count": len(logLines),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleReadLog(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := request.GetString("job_id", "")
	if jobID == "" {
		return mcp.NewToolResultError("job_id parameter is required"), nil
	}

	content, err := s.jobMgr.ReadFullLog(jobID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"content": content,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleSearchLog(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := request.GetString("job_id", "")
	pattern := request.GetString("pattern", "")

	if jobID == "" {
		return mcp.NewToolResultError("job_id parameter is required"), nil
	}
	if pattern == "" {
		return mcp.NewToolResultError("pattern parameter is required"), nil
	}

	matches, err := s.jobMgr.SearchLog(jobID, pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"matches": matches,
		"count":   len(matches),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleScheduleScript(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	schedule := request.GetString("schedule", "")

	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}
	if schedule == "" {
		return mcp.NewToolResultError("schedule parameter is required"), nil
	}

	parameters := make(map[string]string)
	args := request.GetArguments()
	if params, ok := args["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			parameters[k] = fmt.Sprintf("%v", v)
		}
	}

	var scheduledJob *job.ScheduledJob
	var err error

	if t, parseErr := time.Parse(time.RFC3339, schedule); parseErr == nil {
		scheduledJob, err = s.scheduler.ScheduleOnce(name, false, parameters, t)
	} else {
		scheduledJob, err = s.scheduler.Schedule(name, false, parameters, schedule)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("schedule failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"schedule_id": scheduledJob.ID,
		"name":        scheduledJob.Name,
		"schedule":    scheduledJob.Schedule,
		"next_run":    scheduledJob.NextRun,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleScheduleAlias(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	schedule := request.GetString("schedule", "")

	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}
	if schedule == "" {
		return mcp.NewToolResultError("schedule parameter is required"), nil
	}

	var scheduledJob *job.ScheduledJob
	var err error

	if t, parseErr := time.Parse(time.RFC3339, schedule); parseErr == nil {
		scheduledJob, err = s.scheduler.ScheduleOnce(name, true, nil, t)
	} else {
		scheduledJob, err = s.scheduler.Schedule(name, true, nil, schedule)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("schedule failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"schedule_id": scheduledJob.ID,
		"name":        scheduledJob.Name,
		"schedule":    scheduledJob.Schedule,
		"next_run":    scheduledJob.NextRun,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleListScheduled(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scheduled := s.scheduler.ListScheduledJobs()

	scheduleList := make([]map[string]interface{}, len(scheduled))
	for i, sj := range scheduled {
		scheduleList[i] = map[string]interface{}{
			"id":       sj.ID,
			"name":     sj.Name,
			"is_alias": sj.IsAlias,
			"schedule": sj.Schedule,
			"next_run": sj.NextRun,
		}
		if sj.LastRun != nil {
			scheduleList[i]["last_run"] = sj.LastRun
		}
	}

	result := map[string]interface{}{
		"scheduled": scheduleList,
		"count":     len(scheduled),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleUnscheduleJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scheduleID := request.GetString("schedule_id", "")
	if scheduleID == "" {
		return mcp.NewToolResultError("schedule_id parameter is required"), nil
	}

	err := s.scheduler.Unschedule(scheduleID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("unschedule failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"message": fmt.Sprintf("Scheduled job %s removed", scheduleID),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleListScripts(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scripts := s.manager.ListScripts()

	scriptList := make([]map[string]interface{}, len(scripts))
	for i, script := range scripts {
		scriptList[i] = map[string]interface{}{
			"name":        script.Config.Name,
			"description": script.Config.Description,
			"path":        script.AbsolutePath,
			"interpreter": script.Config.Interpreter,
			"source":      script.Source,
			"parameters":  script.Config.Parameters,
		}
	}

	result := map[string]interface{}{
		"scripts": scriptList,
		"count":   len(scripts),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleListAliases(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	aliases := s.manager.ListAliases()

	aliasList := make([]map[string]interface{}, len(aliases))
	for i, a := range aliases {
		aliasList[i] = map[string]interface{}{
			"name":        a.Name,
			"description": a.Description,
			"command":     a.Command,
		}
	}

	result := map[string]interface{}{
		"aliases": aliasList,
		"count":   len(aliases),
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleGetScriptInfo(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	script, err := s.manager.GetScript(name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("script not found: %v", err)), nil
	}

	result := map[string]interface{}{
		"name":        script.Config.Name,
		"description": script.Config.Description,
		"path":        script.AbsolutePath,
		"interpreter": script.Config.Interpreter,
		"source":      script.Source,
		"executable":  script.IsExecutable,
		"parameters":  script.Config.Parameters,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

func (s *Server) handleCreateScript(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := request.GetString("name", "")
	content := request.GetString("content", "")
	interpreter := request.GetString("interpreter", "")

	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}
	if content == "" {
		return mcp.NewToolResultError("content parameter is required"), nil
	}
	if interpreter == "" {
		return mcp.NewToolResultError("interpreter parameter is required"), nil
	}

	err := s.manager.CreateScript(name, content, interpreter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create failed: %v", err)), nil
	}

	script, _ := s.manager.GetScript(name)

	result := map[string]interface{}{
		"message": fmt.Sprintf("Script %s created", name),
		"path":    script.AbsolutePath,
	}

	return mcp.NewToolResultText(formatJSON(result)), nil
}

// Resource Handlers

func (s *Server) handleReadScriptResource(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	name := extractNameFromURI(uri, "script://")

	script, err := s.manager.GetScript(name)
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	content, err := os.ReadFile(script.AbsolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script: %w", err)
	}

	metadata := map[string]interface{}{
		"name":        script.Config.Name,
		"description": script.Config.Description,
		"path":        script.AbsolutePath,
		"interpreter": script.Config.Interpreter,
		"parameters":  script.Config.Parameters,
		"source":      script.Source,
		"executable":  script.IsExecutable,
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")

	text := fmt.Sprintf("# Script: %s\n\n", script.Config.Name)
	text += fmt.Sprintf("## Metadata\n```json\n%s\n```\n\n", string(metadataJSON))
	text += fmt.Sprintf("## Content\n```%s\n%s\n```\n", script.Config.Interpreter, string(content))

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     text,
		},
	}, nil
}

func (s *Server) handleReadAliasResource(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	name := extractNameFromURI(uri, "alias://")

	alias, err := s.manager.GetAlias(name)
	if err != nil {
		return nil, fmt.Errorf("alias not found: %w", err)
	}

	text := fmt.Sprintf("# Alias: %s\n\n", alias.Name)
	text += fmt.Sprintf("**Description:** %s\n\n", alias.Description)
	text += fmt.Sprintf("**Command:**\n```bash\n%s\n```\n", alias.Command)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     text,
		},
	}, nil
}

// Prompt Handlers

func (s *Server) handleExecuteScriptPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	scriptName := request.Params.Arguments["script_name"]
	if scriptName == "" {
		return nil, fmt.Errorf("script_name argument is required")
	}

	script, err := s.manager.GetScript(scriptName)
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	content := fmt.Sprintf("# Execute Script: %s\n\n", script.Config.Name)
	content += fmt.Sprintf("**Description:** %s\n\n", script.Config.Description)

	if len(script.Config.Parameters) > 0 {
		content += "## Required Parameters:\n\n"
		for paramName, param := range script.Config.Parameters {
			required := ""
			if param.Required {
				required = " (required)"
			}
			content += fmt.Sprintf("- **%s**%s: %s\n", paramName, required, param.Description)
			if param.Default != "" {
				content += fmt.Sprintf("  - Default: `%s`\n", param.Default)
			}
			if len(param.ValidValues) > 0 {
				content += fmt.Sprintf("  - Valid values: %v\n", param.ValidValues)
			}
		}
		content += "\n"
	}

	content += "## Execution\n\n"
	content += "To execute this script, use the `execute_script` tool with the following parameters:\n\n"
	content += "```json\n{\n" //nolint:goconst // JSON template formatting
	content += fmt.Sprintf("  \"name\": \"%s\",\n", scriptName)

	parametersStr := request.Params.Arguments["parameters"]
	if parametersStr != "" {
		content += fmt.Sprintf("  \"parameters\": %s\n", parametersStr)
	} else {
		content += "  \"parameters\": {}\n"
	}
	content += "}\n```" //nolint:goconst // JSON template formatting

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Execute script: %s", scriptName),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content),
			},
		},
	}, nil
}

func (s *Server) handleExecuteAliasPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	aliasName := request.Params.Arguments["alias_name"]
	if aliasName == "" {
		return nil, fmt.Errorf("alias_name argument is required")
	}

	alias, err := s.manager.GetAlias(aliasName)
	if err != nil {
		return nil, fmt.Errorf("alias not found: %w", err)
	}

	content := fmt.Sprintf("# Execute Alias: %s\n\n", alias.Name)
	content += fmt.Sprintf("**Description:** %s\n\n", alias.Description)
	content += fmt.Sprintf("**Command:** `%s`\n\n", alias.Command)
	content += "## Execution\n\n"
	content += "To execute this alias, use the `execute_alias` tool:\n\n"
	content += "```json\n{\n"
	content += fmt.Sprintf("  \"name\": \"%s\"\n", aliasName)
	content += "}\n```"

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Execute alias: %s", aliasName),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content),
			},
		},
	}, nil
}

func (s *Server) handleScheduleScriptPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	scriptName := request.Params.Arguments["script_name"]
	schedule := request.Params.Arguments["schedule"]

	if scriptName == "" {
		return nil, fmt.Errorf("script_name argument is required")
	}
	if schedule == "" {
		return nil, fmt.Errorf("schedule argument is required")
	}

	content := fmt.Sprintf("# Schedule Script: %s\n\n", scriptName)
	content += fmt.Sprintf("**Schedule:** %s\n\n", schedule)
	content += "Use the `schedule_script` tool to schedule this execution:\n\n"
	content += "```json\n{\n"
	content += fmt.Sprintf("  \"name\": \"%s\",\n", scriptName)
	content += fmt.Sprintf("  \"schedule\": \"%s\",\n", schedule)

	parametersStr := request.Params.Arguments["parameters"]
	if parametersStr != "" {
		content += fmt.Sprintf("  \"parameters\": %s\n", parametersStr)
	} else {
		content += "  \"parameters\": {}\n"
	}
	content += "}\n```"

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Schedule script: %s", scriptName),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content),
			},
		},
	}, nil
}

func (s *Server) handleMonitorJobPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	jobID := request.Params.Arguments["job_id"]
	if jobID == "" {
		return nil, fmt.Errorf("job_id argument is required")
	}

	content := fmt.Sprintf("# Monitor Job: %s\n\n", jobID)
	content += "## Available Actions:\n\n"
	content += "1. **Get Job Status:**\n```json\n{\"tool\": \"get_job\", \"job_id\": \"" + jobID + "\"}\n```\n\n"
	content += "2. **Tail Job Logs:**\n```json\n{\"tool\": \"tail_log\", \"job_id\": \"" + jobID + "\", \"lines\": 10}\n```\n\n"
	content += "3. **Cancel Job:**\n```json\n{\"tool\": \"cancel_job\", \"job_id\": \"" + jobID + "\"}\n```\n"

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Monitor job: %s", jobID),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content),
			},
		},
	}, nil
}

func (s *Server) handleListScriptsPrompt(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var content strings.Builder
	content.WriteString("# Available Scripts and Aliases\n\n")

	scripts := s.manager.ListScripts()
	if len(scripts) > 0 {
		content.WriteString("## Scripts\n\n")
		for _, script := range scripts {
			content.WriteString(fmt.Sprintf("### %s\n", script.Config.Name))
			content.WriteString(fmt.Sprintf("- **Description:** %s\n", script.Config.Description))
			content.WriteString(fmt.Sprintf("- **Interpreter:** %s\n", script.Config.Interpreter))
			content.WriteString(fmt.Sprintf("- **Source:** %s\n", script.Source))
			if len(script.Config.Parameters) > 0 {
				content.WriteString("- **Parameters:**\n")
				for paramName, param := range script.Config.Parameters {
					content.WriteString(fmt.Sprintf("  - `%s`: %s\n", paramName, param.Description))
				}
			}
			content.WriteString("\n")
		}
	}

	aliases := s.manager.ListAliases()
	if len(aliases) > 0 {
		content.WriteString("## Aliases\n\n")
		for _, alias := range aliases {
			content.WriteString(fmt.Sprintf("### %s\n", alias.Name))
			content.WriteString(fmt.Sprintf("- **Description:** %s\n", alias.Description))
			content.WriteString(fmt.Sprintf("- **Command:** `%s`\n\n", alias.Command))
		}
	}

	return &mcp.GetPromptResult{
		Description: "List of available scripts and aliases",
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content.String()),
			},
		},
	}, nil
}

func (s *Server) handleCreateScriptPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	name := request.Params.Arguments["name"]
	interpreter := request.Params.Arguments["interpreter"]
	scriptContent := request.Params.Arguments["content"]

	if name == "" || interpreter == "" || scriptContent == "" {
		return nil, fmt.Errorf("name, interpreter, and content arguments are required")
	}

	promptContent := fmt.Sprintf("# Create Script: %s\n\n", name)
	promptContent += fmt.Sprintf("**Interpreter:** %s\n\n", interpreter)
	promptContent += "Use the `create_script` tool to create this script:\n\n"
	promptContent += "```json\n{\n"
	promptContent += fmt.Sprintf("  \"name\": \"%s\",\n", name)
	promptContent += fmt.Sprintf("  \"interpreter\": \"%s\",\n", interpreter)
	promptContent += fmt.Sprintf("  \"content\": %q\n", scriptContent)
	promptContent += "}\n```"

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Create script: %s", name),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(promptContent),
			},
		},
	}, nil
}

// Helper functions

func formatJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting result: %v", err)
	}
	return string(data)
}

func extractNameFromURI(uri, prefix string) string {
	if strings.HasPrefix(uri, prefix) {
		return strings.TrimPrefix(uri, prefix)
	}
	return uri
}
