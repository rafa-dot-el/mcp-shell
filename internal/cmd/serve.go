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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server in stdio mode",
	Long: `Start the MCP (Model Context Protocol) server in stdio mode.

The server will read JSON-RPC requests from stdin and write responses to stdout.
This mode is designed to be used as a subprocess by MCP clients like Claude Desktop.

Configuration:
  The server reads configuration from:
  - Command line flags (highest priority)
  - Environment variables (MCP_SHELL_*)
  - Configuration file (--config flag or default locations)
  - Default values (lowest priority)

Examples:
  # Start server with default configuration
  mcp-shell serve

  # Start server with custom config file
  mcp-shell serve --config /path/to/config.yaml

  # Start server with verbose output (to stderr)
  mcp-shell serve --verbose`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create MCP server
		server, err := mcp.NewServer(cfg)
		if err != nil {
			return fmt.Errorf("failed to create MCP server: %w", err)
		}

		// Create stdio transport
		transport := mcp.NewStdioTransport(server)

		// Setup signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start server in goroutine
		errChan := make(chan error, 1)
		go func() {
			if viper.GetBool("verbose") {
				fmt.Fprintln(os.Stderr, "[*] MCP server starting in stdio mode")
			}
			errChan <- transport.Serve()
		}()

		// Wait for shutdown signal or error
		select {
		case sig := <-sigChan:
			if viper.GetBool("verbose") {
				fmt.Fprintf(os.Stderr, "[*] Received signal: %v\n", sig)
				fmt.Fprintln(os.Stderr, "[*] Shutting down gracefully...")
			}
			if err := transport.Shutdown(); err != nil {
				return fmt.Errorf("shutdown error: %w", err)
			}
			return nil
		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Add serve-specific flags if needed
	serveCmd.Flags().String("log-dir", "", "Directory for job logs (overrides config)")
	serveCmd.Flags().Int("max-jobs", 0, "Maximum parallel jobs (overrides config, 0 uses config default)")

	// Bind flags to viper
	viper.BindPFlag("execution.log_directory", serveCmd.Flags().Lookup("log-dir"))
	viper.BindPFlag("execution.max_parallel_jobs", serveCmd.Flags().Lookup("max-jobs"))
}

// loadConfig loads the configuration from viper
func loadConfig() (*config.Config, error) {
	// Start with default configuration
	cfg := config.DefaultConfig()

	// Override version from build
	cfg.MCP.Version = version

	// Unmarshal viper config into our struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate configuration
	errors := cfg.Validate()
	if len(errors) > 0 {
		// Return first validation error
		return nil, errors[0]
	}

	return cfg, nil
}
