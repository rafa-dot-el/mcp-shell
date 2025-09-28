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

func TestConfigValidation(t *testing.T) {
	// Test configuration validation integration
	cfg := &config.Config{
		Verbose:  false,
		Debug:    false,
		LogLevel: "info",
	}

	if !cfg.IsValid() {
		t.Error("Expected valid config to pass validation")
	}

	// Test invalid config
	invalidCfg := &config.Config{
		Verbose:  true,
		Debug:    true,
		LogLevel: "invalid",
	}

	if invalidCfg.IsValid() {
		t.Error("Expected invalid config to fail validation")
	}

	errors := invalidCfg.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid config")
	}
}

func TestDefaultConfiguration(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.Verbose {
		t.Error("Expected default verbose to be false")
	}

	if cfg.Debug {
		t.Error("Expected default debug to be false")
	}

	// Test that default config is valid
	if !cfg.IsValid() {
		t.Error("Expected default config to be valid")
	}
}