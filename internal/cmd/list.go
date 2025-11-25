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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rafa-dot-el/mcp-shell/pkg/config"
	"github.com/rafa-dot-el/mcp-shell/pkg/script"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scripts and aliases",
	Long: `List all configured scripts and aliases available for execution.

This command displays:
  - Script names, descriptions, and file paths
  - Alias names, descriptions, and commands
  - Parameter information for scripts

Output formats:
  - table (default): Human-readable table format
  - json: JSON format for programmatic use
  - simple: Simple list of names only

Examples:
  # List all scripts and aliases in table format
  mcp-shell list

  # List in JSON format
  mcp-shell list --format json

  # List only scripts
  mcp-shell list --scripts

  # List only aliases
  mcp-shell list --aliases

  # List with detailed parameter information
  mcp-shell list --details`,

	RunE: func(cmd *cobra.Command, _ []string) error {
		format := cmd.Flag("format").Value.String()
		//nolint:goconst // Standard flag value comparisons
		scriptsOnly := cmd.Flag("scripts").Value.String() == "true"
		aliasesOnly := cmd.Flag("aliases").Value.String() == "true"
		details := cmd.Flag("details").Value.String() == "true"

		// Load configuration
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create script manager
		scriptMgr, err := script.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("failed to create script manager: %w", err)
		}

		scripts := scriptMgr.ListScripts()
		aliases := scriptMgr.ListAliases()

		// Handle format
		switch format {
		case "json":
			return printJSON(scripts, aliases, scriptsOnly, aliasesOnly)
		case "simple":
			return printSimple(scripts, aliases, scriptsOnly, aliasesOnly)
		default:
			return printTable(scripts, aliases, scriptsOnly, aliasesOnly, details)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().String("format", "table", "Output format (table, json, simple)")
	listCmd.Flags().Bool("scripts", false, "List only scripts")
	listCmd.Flags().Bool("aliases", false, "List only aliases")
	listCmd.Flags().Bool("details", false, "Show detailed parameter information")
}

func printJSON(scripts []*script.LoadedScript, aliases []*config.Alias, scriptsOnly, aliasesOnly bool) error {
	output := make(map[string]interface{})

	if !aliasesOnly {
		// Convert to simpler structure for JSON output
		scriptList := make([]map[string]interface{}, len(scripts))
		for i, s := range scripts {
			scriptList[i] = map[string]interface{}{
				"name":        s.Config.Name,
				"description": s.Config.Description,
				"path":        s.AbsolutePath,
				"interpreter": s.Config.Interpreter,
				"parameters":  s.Config.Parameters,
			}
		}
		output["scripts"] = scriptList
	}
	if !scriptsOnly {
		output["aliases"] = aliases
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printSimple(scripts []*script.LoadedScript, aliases []*config.Alias, scriptsOnly, aliasesOnly bool) error {
	if !aliasesOnly {
		for _, s := range scripts {
			fmt.Println(s.Config.Name)
		}
	}
	if !scriptsOnly {
		for _, a := range aliases {
			fmt.Println(a.Name)
		}
	}
	return nil
}

func printTable(scripts []*script.LoadedScript, aliases []*config.Alias, scriptsOnly, aliasesOnly bool, details bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !aliasesOnly && len(scripts) > 0 {
		_, _ = fmt.Fprintln(w, "SCRIPTS")
		_, _ = fmt.Fprintln(w, strings.Repeat("-", 80))
		_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tPATH")

		for _, s := range scripts {
			desc := s.Config.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			path := s.AbsolutePath
			if len(path) > 40 {
				path = "..." + path[len(path)-37:]
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", s.Config.Name, desc, path)

			if details && len(s.Config.Parameters) > 0 {
				_, _ = fmt.Fprintln(w, "\tParameters:")
				for name, param := range s.Config.Parameters {
					required := ""
					if param.Required {
						required = " (required)"
					}
					defaultVal := ""
					if param.Default != "" {
						defaultVal = fmt.Sprintf(" [default: %s]", param.Default)
					}
					_, _ = fmt.Fprintf(w, "\t  %s%s%s - %s\n", name, required, defaultVal, param.Description)
				}
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	if !scriptsOnly && len(aliases) > 0 {
		_, _ = fmt.Fprintln(w, "ALIASES")
		_, _ = fmt.Fprintln(w, strings.Repeat("-", 80))
		_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tCOMMAND")

		for _, a := range aliases {
			desc := a.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			command := a.Command
			if len(command) > 40 {
				command = command[:37] + "..."
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", a.Name, desc, command)
		}
		_, _ = fmt.Fprintln(w)
	}

	return w.Flush()
}
