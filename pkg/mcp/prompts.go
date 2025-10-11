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
	"fmt"
	"strings"
)

// PromptHandler handles MCP prompt requests
type PromptHandler struct {
	server *Server
}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler(server *Server) *PromptHandler {
	return &PromptHandler{
		server: server,
	}
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   []PromptArgument  `json:"arguments,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// PromptMessage represents a message in a prompt response
type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// ListPrompts returns all available prompt templates
func (h *PromptHandler) ListPrompts() []Prompt {
	return []Prompt{
		{
			Name:        "execute-script",
			Description: "Execute a script with parameters",
			Arguments: []PromptArgument{
				{
					Name:        "script_name",
					Description: "Name of the script to execute",
					Required:    true,
				},
				{
					Name:        "parameters",
					Description: "JSON object of parameters (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "execute-alias",
			Description: "Execute an alias command",
			Arguments: []PromptArgument{
				{
					Name:        "alias_name",
					Description: "Name of the alias to execute",
					Required:    true,
				},
			},
		},
		{
			Name:        "schedule-script",
			Description: "Schedule a script to run at a specific time or with cron",
			Arguments: []PromptArgument{
				{
					Name:        "script_name",
					Description: "Name of the script to schedule",
					Required:    true,
				},
				{
					Name:        "schedule",
					Description: "Cron expression or timestamp",
					Required:    true,
				},
				{
					Name:        "parameters",
					Description: "JSON object of parameters (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "monitor-job",
			Description: "Monitor a running job and view its logs",
			Arguments: []PromptArgument{
				{
					Name:        "job_id",
					Description: "ID of the job to monitor",
					Required:    true,
				},
			},
		},
		{
			Name:        "list-available-scripts",
			Description: "List all available scripts and aliases",
			Arguments:   []PromptArgument{},
		},
		{
			Name:        "create-script",
			Description: "Create a new script (if allowed by security settings)",
			Arguments: []PromptArgument{
				{
					Name:        "name",
					Description: "Name for the new script",
					Required:    true,
				},
				{
					Name:        "interpreter",
					Description: "Interpreter to use (bash, python3, etc.)",
					Required:    true,
				},
				{
					Name:        "content",
					Description: "Script content",
					Required:    true,
				},
			},
		},
	}
}

// GetPrompt retrieves a specific prompt with arguments
func (h *PromptHandler) GetPrompt(name string, args map[string]string) (*GetPromptResult, error) {
	switch name {
	case "execute-script":
		return h.getExecuteScriptPrompt(args)
	case "execute-alias":
		return h.getExecuteAliasPrompt(args)
	case "schedule-script":
		return h.getScheduleScriptPrompt(args)
	case "monitor-job":
		return h.getMonitorJobPrompt(args)
	case "list-available-scripts":
		return h.getListScriptsPrompt(args)
	case "create-script":
		return h.getCreateScriptPrompt(args)
	default:
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}
}

func (h *PromptHandler) getExecuteScriptPrompt(args map[string]string) (*GetPromptResult, error) {
	scriptName := args["script_name"]
	if scriptName == "" {
		return nil, fmt.Errorf("script_name argument is required")
	}

	// Get script details
	script, err := h.server.manager.GetScript(scriptName)
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	// Build prompt message
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
	content += "```json\n{\n"
	content += fmt.Sprintf("  \"name\": \"%s\",\n", scriptName)

	parametersStr := args["parameters"]
	if parametersStr != "" {
		content += fmt.Sprintf("  \"parameters\": %s\n", parametersStr)
	} else {
		content += "  \"parameters\": {}\n"
	}
	content += "}\n```"

	return &GetPromptResult{
		Description: fmt.Sprintf("Execute script: %s", scriptName),
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}, nil
}

func (h *PromptHandler) getExecuteAliasPrompt(args map[string]string) (*GetPromptResult, error) {
	aliasName := args["alias_name"]
	if aliasName == "" {
		return nil, fmt.Errorf("alias_name argument is required")
	}

	alias, err := h.server.manager.GetAlias(aliasName)
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

	return &GetPromptResult{
		Description: fmt.Sprintf("Execute alias: %s", aliasName),
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}, nil
}

func (h *PromptHandler) getScheduleScriptPrompt(args map[string]string) (*GetPromptResult, error) {
	scriptName := args["script_name"]
	schedule := args["schedule"]

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

	parametersStr := args["parameters"]
	if parametersStr != "" {
		content += fmt.Sprintf("  \"parameters\": %s\n", parametersStr)
	} else {
		content += "  \"parameters\": {}\n"
	}
	content += "}\n```"

	return &GetPromptResult{
		Description: fmt.Sprintf("Schedule script: %s", scriptName),
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}, nil
}

func (h *PromptHandler) getMonitorJobPrompt(args map[string]string) (*GetPromptResult, error) {
	jobID := args["job_id"]
	if jobID == "" {
		return nil, fmt.Errorf("job_id argument is required")
	}

	content := fmt.Sprintf("# Monitor Job: %s\n\n", jobID)
	content += "## Available Actions:\n\n"
	content += "1. **Get Job Status:**\n```json\n{\"tool\": \"get_job\", \"job_id\": \"" + jobID + "\"}\n```\n\n"
	content += "2. **Tail Job Logs:**\n```json\n{\"tool\": \"tail_log\", \"job_id\": \"" + jobID + "\", \"lines\": 10}\n```\n\n"
	content += "3. **Cancel Job:**\n```json\n{\"tool\": \"cancel_job\", \"job_id\": \"" + jobID + "\"}\n```\n"

	return &GetPromptResult{
		Description: fmt.Sprintf("Monitor job: %s", jobID),
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}, nil
}

func (h *PromptHandler) getListScriptsPrompt(args map[string]string) (*GetPromptResult, error) {
	var content strings.Builder
	content.WriteString("# Available Scripts and Aliases\n\n")

	// List scripts
	scripts := h.server.manager.ListScripts()
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

	// List aliases
	aliases := h.server.manager.ListAliases()
	if len(aliases) > 0 {
		content.WriteString("## Aliases\n\n")
		for _, alias := range aliases {
			content.WriteString(fmt.Sprintf("### %s\n", alias.Name))
			content.WriteString(fmt.Sprintf("- **Description:** %s\n", alias.Description))
			content.WriteString(fmt.Sprintf("- **Command:** `%s`\n\n", alias.Command))
		}
	}

	return &GetPromptResult{
		Description: "List of available scripts and aliases",
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: content.String(),
			},
		},
	}, nil
}

func (h *PromptHandler) getCreateScriptPrompt(args map[string]string) (*GetPromptResult, error) {
	name := args["name"]
	interpreter := args["interpreter"]
	content := args["content"]

	if name == "" || interpreter == "" || content == "" {
		return nil, fmt.Errorf("name, interpreter, and content arguments are required")
	}

	promptContent := fmt.Sprintf("# Create Script: %s\n\n", name)
	promptContent += fmt.Sprintf("**Interpreter:** %s\n\n", interpreter)
	promptContent += "Use the `create_script` tool to create this script:\n\n"
	promptContent += "```json\n{\n"
	promptContent += fmt.Sprintf("  \"name\": \"%s\",\n", name)
	promptContent += fmt.Sprintf("  \"interpreter\": \"%s\",\n", interpreter)
	promptContent += fmt.Sprintf("  \"content\": %q\n", content)
	promptContent += "}\n```"

	return &GetPromptResult{
		Description: fmt.Sprintf("Create script: %s", name),
		Messages: []PromptMessage{
			{
				Role:    "user",
				Content: promptContent,
			},
		},
	}, nil
}
