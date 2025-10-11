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
	"encoding/json"
	"fmt"
	"os"
)

// ResourceHandler handles MCP resource requests
type ResourceHandler struct {
	server *Server
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(server *Server) *ResourceHandler {
	return &ResourceHandler{
		server: server,
	}
}

// Resource represents an MCP resource
type Resource struct {
	URI         string      `json:"uri"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	MimeType    string      `json:"mimeType"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ListResources returns all available resources
func (h *ResourceHandler) ListResources() ([]Resource, error) {
	var resources []Resource

	// Add script resources
	scripts := h.server.manager.ListScripts()
	for _, script := range scripts {
		resources = append(resources, Resource{
			URI:         fmt.Sprintf("script://%s", script.Config.Name),
			Name:        script.Config.Name,
			Description: script.Config.Description,
			MimeType:    "application/x-sh",
			Metadata: map[string]interface{}{
				"path":        script.AbsolutePath,
				"interpreter": script.Config.Interpreter,
				"source":      script.Source,
				"parameters":  script.Config.Parameters,
			},
		})
	}

	// Add alias resources
	aliases := h.server.manager.ListAliases()
	for _, alias := range aliases {
		resources = append(resources, Resource{
			URI:         fmt.Sprintf("alias://%s", alias.Name),
			Name:        alias.Name,
			Description: alias.Description,
			MimeType:    "text/plain",
			Metadata: map[string]interface{}{
				"command": alias.Command,
			},
		})
	}

	return resources, nil
}

// ReadResource retrieves the content of a specific resource
func (h *ResourceHandler) ReadResource(uri string) (*ResourceContent, error) {
	// Parse URI to determine resource type
	resourceType, name, err := parseResourceURI(uri)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "script":
		return h.readScriptResource(name)
	case "alias":
		return h.readAliasResource(name)
	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// readScriptResource reads a script resource
func (h *ResourceHandler) readScriptResource(name string) (*ResourceContent, error) {
	script, err := h.server.manager.GetScript(name)
	if err != nil {
		return nil, fmt.Errorf("script not found: %w", err)
	}

	// Read script content
	content, err := os.ReadFile(script.AbsolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script: %w", err)
	}

	// Build resource content with metadata
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

	return &ResourceContent{
		URI:      fmt.Sprintf("script://%s", name),
		MimeType: "text/markdown",
		Text:     text,
	}, nil
}

// readAliasResource reads an alias resource
func (h *ResourceHandler) readAliasResource(name string) (*ResourceContent, error) {
	alias, err := h.server.manager.GetAlias(name)
	if err != nil {
		return nil, fmt.Errorf("alias not found: %w", err)
	}

	text := fmt.Sprintf("# Alias: %s\n\n", alias.Name)
	text += fmt.Sprintf("**Description:** %s\n\n", alias.Description)
	text += fmt.Sprintf("**Command:**\n```bash\n%s\n```\n", alias.Command)

	return &ResourceContent{
		URI:      fmt.Sprintf("alias://%s", name),
		MimeType: "text/markdown",
		Text:     text,
	}, nil
}

// parseResourceURI parses a resource URI into type and name
func parseResourceURI(uri string) (string, string, error) {
	// URI format: <type>://<name>
	// Examples: script://backup.sh, alias://git-status

	var resourceType, name string
	_, err := fmt.Sscanf(uri, "%[^:]://%s", &resourceType, &name)
	if err != nil {
		return "", "", fmt.Errorf("invalid resource URI format: %s", uri)
	}

	return resourceType, name, nil
}

// ListResourceTemplates returns templates for resource URIs
func (h *ResourceHandler) ListResourceTemplates() []ResourceTemplate {
	return []ResourceTemplate{
		{
			URITemplate: "script://{name}",
			Name:        "Script Resource",
			Description: "Access a specific script by name",
			MimeType:    "text/markdown",
		},
		{
			URITemplate: "alias://{name}",
			Name:        "Alias Resource",
			Description: "Access a specific alias by name",
			MimeType:    "text/markdown",
		},
	}
}

// ResourceTemplate represents a URI template for resources
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}
