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

	"github.com/rafa-dot-el/mcp-shell/pkg/script"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate the MCP Shell configuration file for correctness.

This command checks:
  - Configuration file syntax and structure
  - Script file existence and permissions
  - Parameter definitions
  - Alias command syntax
  - Log directory accessibility
  - MCP server configuration

The command exits with code 0 if validation succeeds, or code 1 if errors are found.

Examples:
  # Validate default configuration
  mcp-shell validate

  # Validate specific config file
  mcp-shell validate --config /path/to/config.yaml

  # Validate with verbose output
  mcp-shell validate --verbose`,

	RunE: func(cmd *cobra.Command, _ []string) error {
		verbose := cmd.Flag("verbose").Value.String() == "true" //nolint:goconst // Standard flag value comparison

		if verbose {
			fmt.Fprintln(os.Stderr, "[*] Loading configuration...")
		}

		// Load configuration
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Configuration validation failed: %v\n", err)
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[+] Configuration loaded successfully\n")
			fmt.Fprintf(os.Stderr, "[*] Validating %d scripts and %d aliases...\n", len(cfg.Scripts), len(cfg.Aliases))
		}

		// Create script manager to validate scripts
		scriptMgr, err := script.NewManager(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Script validation failed: %v\n", err)
			return fmt.Errorf("script validation failed: %w", err)
		}

		// Validate each script
		scripts := scriptMgr.ListScripts()
		scriptErrors := 0
		for _, s := range scripts {
			if verbose {
				fmt.Fprintf(os.Stderr, "[*] Validating script: %s\n", s.Config.Name)
			}

			// Check if script file exists
			if _, err := os.Stat(s.AbsolutePath); err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "[-] Script '%s': file not found: %s\n", s.Config.Name, s.AbsolutePath)
					scriptErrors++
					continue
				}
				fmt.Fprintf(os.Stderr, "[-] Script '%s': cannot access file: %v\n", s.Config.Name, err)
				scriptErrors++
				continue
			}

			// Check if script is executable
			if !s.IsExecutable {
				fmt.Fprintf(os.Stderr, "[-] Script '%s': file is not executable: %s\n", s.Config.Name, s.AbsolutePath)
				scriptErrors++
				continue
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "[+] Script '%s' is valid\n", s.Config.Name)
			}
		}

		// Validate aliases
		aliases := scriptMgr.ListAliases()
		aliasErrors := 0
		for _, a := range aliases {
			if verbose {
				fmt.Fprintf(os.Stderr, "[*] Validating alias: %s\n", a.Name)
			}

			// Basic validation - ensure command is not empty
			if a.Command == "" {
				fmt.Fprintf(os.Stderr, "[-] Alias '%s': command is empty\n", a.Name)
				aliasErrors++
				continue
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "[+] Alias '%s' is valid\n", a.Name)
			}
		}

		// Check log directory
		if verbose {
			fmt.Fprintf(os.Stderr, "[*] Checking log directory: %s\n", cfg.Execution.LogDirectory)
		}

		logDirInfo, err := os.Stat(cfg.Execution.LogDirectory)
		if err != nil {
			if os.IsNotExist(err) {
				// Try to create it
				if err := os.MkdirAll(cfg.Execution.LogDirectory, 0750); err != nil {
					fmt.Fprintf(os.Stderr, "[-] Cannot create log directory: %v\n", err)
					return fmt.Errorf("log directory validation failed: %w", err)
				}
				if verbose {
					fmt.Fprintf(os.Stderr, "[+] Created log directory: %s\n", cfg.Execution.LogDirectory)
				}
			} else {
				fmt.Fprintf(os.Stderr, "[-] Cannot access log directory: %v\n", err)
				return fmt.Errorf("log directory validation failed: %w", err)
			}
		} else if !logDirInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "[-] Log path exists but is not a directory: %s\n", cfg.Execution.LogDirectory)
			return fmt.Errorf("log directory validation failed: not a directory")
		}

		// Print summary
		totalErrors := scriptErrors + aliasErrors
		if totalErrors > 0 {
			fmt.Fprintf(os.Stderr, "\n[-] Validation failed with %d errors\n", totalErrors)
			fmt.Fprintf(os.Stderr, "    Scripts: %d errors\n", scriptErrors)
			fmt.Fprintf(os.Stderr, "    Aliases: %d errors\n", aliasErrors)
			return fmt.Errorf("validation failed with %d errors", totalErrors)
		}

		fmt.Println("[+] Configuration validation successful")
		fmt.Printf("    Scripts: %d valid\n", len(scripts))
		fmt.Printf("    Aliases: %d valid\n", len(aliases))
		fmt.Printf("    Log directory: %s\n", cfg.Execution.LogDirectory)
		fmt.Printf("    Max parallel jobs: %d\n", cfg.Execution.MaxParallelJobs)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
