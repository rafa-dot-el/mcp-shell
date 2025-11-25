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

package cmd

import (
	"testing"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
)

func TestDefaultConfiguration(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}

	// Test that default config is valid
	if !cfg.IsValid() {
		t.Error("Expected default config to be valid")
	}
}
