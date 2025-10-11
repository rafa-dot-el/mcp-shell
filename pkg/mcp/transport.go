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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// StdioTransport implements MCP protocol over stdio
type StdioTransport struct {
	server *Server
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(server *Server) *StdioTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &StdioTransport{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Serve starts serving MCP requests over stdio
func (t *StdioTransport) Serve() error {
	defer t.cancel()

	for {
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		default:
			if err := t.handleRequest(); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}

// handleRequest reads and processes a single JSON-RPC request
func (t *StdioTransport) handleRequest() error {
	// Read line
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// Parse request
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.sendError(nil, -32700, "Parse error", err.Error())
		return nil
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		t.sendError(req.ID, -32600, "Invalid Request", "jsonrpc must be '2.0'")
		return nil
	}

	// Route request
	switch req.Method {
	case "initialize":
		t.handleInitialize(req)
	case "initialized":
		// Notification, no response needed
		return nil
	case "resources/list":
		t.handleListResources(req)
	case "resources/read":
		t.handleReadResource(req)
	case "resources/templates/list":
		t.handleListResourceTemplates(req)
	case "prompts/list":
		t.handleListPrompts(req)
	case "prompts/get":
		t.handleGetPrompt(req)
	case "tools/list":
		t.handleListTools(req)
	case "tools/call":
		t.handleCallTool(req)
	case "ping":
		t.sendResponse(req.ID, map[string]interface{}{})
	default:
		t.sendError(req.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", req.Method))
	}

	return nil
}

// handleInitialize handles the initialize request
func (t *StdioTransport) handleInitialize(req JSONRPCRequest) {
	info := t.server.GetServerInfo()
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"resources": map[string]interface{}{
				"subscribe": false,
			},
			"prompts": map[string]interface{}{},
			"tools":   map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    info.Name,
			"version": info.Version,
		},
	}
	t.sendResponse(req.ID, result)
}

// handleListResources handles the resources/list request
func (t *StdioTransport) handleListResources(req JSONRPCRequest) {
	resources, err := t.server.resources.ListResources()
	if err != nil {
		t.sendError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	result := map[string]interface{}{
		"resources": resources,
	}
	t.sendResponse(req.ID, result)
}

// handleReadResource handles the resources/read request
func (t *StdioTransport) handleReadResource(req JSONRPCRequest) {
	uri, ok := req.Params["uri"].(string)
	if !ok {
		t.sendError(req.ID, -32602, "Invalid params", "uri parameter is required")
		return
	}

	content, err := t.server.resources.ReadResource(uri)
	if err != nil {
		t.sendError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	result := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"uri":      content.URI,
				"mimeType": content.MimeType,
				"text":     content.Text,
			},
		},
	}
	t.sendResponse(req.ID, result)
}

// handleListResourceTemplates handles the resources/templates/list request
func (t *StdioTransport) handleListResourceTemplates(req JSONRPCRequest) {
	templates := t.server.resources.ListResourceTemplates()
	result := map[string]interface{}{
		"resourceTemplates": templates,
	}
	t.sendResponse(req.ID, result)
}

// handleListPrompts handles the prompts/list request
func (t *StdioTransport) handleListPrompts(req JSONRPCRequest) {
	prompts := t.server.prompts.ListPrompts()
	result := map[string]interface{}{
		"prompts": prompts,
	}
	t.sendResponse(req.ID, result)
}

// handleGetPrompt handles the prompts/get request
func (t *StdioTransport) handleGetPrompt(req JSONRPCRequest) {
	name, ok := req.Params["name"].(string)
	if !ok {
		t.sendError(req.ID, -32602, "Invalid params", "name parameter is required")
		return
	}

	// Extract arguments
	args := make(map[string]string)
	if argsParam, ok := req.Params["arguments"].(map[string]interface{}); ok {
		for k, v := range argsParam {
			args[k] = fmt.Sprintf("%v", v)
		}
	}

	prompt, err := t.server.prompts.GetPrompt(name, args)
	if err != nil {
		t.sendError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	t.sendResponse(req.ID, prompt)
}

// handleListTools handles the tools/list request
func (t *StdioTransport) handleListTools(req JSONRPCRequest) {
	tools := t.server.tools.ListTools()
	result := map[string]interface{}{
		"tools": tools,
	}
	t.sendResponse(req.ID, result)
}

// handleCallTool handles the tools/call request
func (t *StdioTransport) handleCallTool(req JSONRPCRequest) {
	name, ok := req.Params["name"].(string)
	if !ok {
		t.sendError(req.ID, -32602, "Invalid params", "name parameter is required")
		return
	}

	// Extract arguments
	args := make(map[string]interface{})
	if argsParam, ok := req.Params["arguments"].(map[string]interface{}); ok {
		args = argsParam
	}

	toolResult, err := t.server.tools.CallTool(t.ctx, name, args)
	if err != nil {
		t.sendError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	// Format result according to MCP protocol
	result := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": formatToolResult(toolResult),
			},
		},
	}

	if toolResult.IsError {
		result["isError"] = true
	}

	t.sendResponse(req.ID, result)
}

// formatToolResult formats a tool result as text
func formatToolResult(result *ToolResult) string {
	data, err := json.MarshalIndent(result.Content, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting result: %v", err)
	}
	return string(data)
}

// sendResponse sends a JSON-RPC response
func (t *StdioTransport) sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	t.sendJSON(resp)
}

// sendError sends a JSON-RPC error response
func (t *StdioTransport) sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	t.sendJSON(resp)
}

// sendJSON sends a JSON object to stdout
func (t *StdioTransport) sendJSON(v interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
		return
	}

	t.writer.Write(data)
	t.writer.WriteByte('\n')
	t.writer.Flush()
}

// Shutdown gracefully shuts down the transport
func (t *StdioTransport) Shutdown() error {
	t.cancel()
	return t.server.Shutdown(t.ctx)
}
