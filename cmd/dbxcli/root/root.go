// Package root provides the exported Cobra root command constructor.
// Downstream repos (dbx-ee, dbx-agent) import root.New(version) to extend
// the command tree before calling Execute().
package root

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// New returns the Cobra root command initialised with the given version string.
func New(version string) *cobra.Command {
	var formatFlag string

	cmd := &cobra.Command{
		Use:     "dbxcli",
		Short:   "dbx — multi-database management platform",
		Version: version,
		Long: `dbx is a multi-database management platform supporting Oracle and PostgreSQL.

Commands use named parameters (emcli-style):
  dbxcli <domain> <action> entity_name=<name> entity_type=<type> [key=value ...]`,
	}

	cmd.PersistentFlags().StringVar(&formatFlag, "format", "table", "output format: table, json, yaml")

	cmd.AddCommand(NewTargetCmd())
	cmd.AddCommand(NewDBCmd())
	cmd.AddCommand(NewLinuxCmd())
	cmd.AddCommand(NewHostCmd())
	cmd.AddCommand(NewPgCmd())
	cmd.AddCommand(NewServeCmd())
	cmd.AddCommand(NewMCPCmd())
	cmd.AddCommand(NewLicenseCmd())
	cmd.AddCommand(NewPolicyCmd())

	return cmd
}

// ParseNamedParams parses emcli-style key=value arguments.
func ParseNamedParams(args []string) (map[string]string, error) {
	params := make(map[string]string, len(args))
	for _, arg := range args {
		idx := strings.IndexByte(arg, '=')
		if idx < 1 {
			return nil, fmt.Errorf("invalid parameter %q: expected key=value format", arg)
		}
		params[arg[:idx]] = arg[idx+1:]
	}
	return params, nil
}

// NewTargetCmd creates the "target" subcommand group.
func NewTargetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage system targets",
		Long: `Manage the target registry — databases, hosts, and services that dbx connects to.
Targets are stored in ~/.dbx/targets/ as YAML files.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List registered targets",
		Long:  `List all registered targets with their type, host, and connection status.`,
		Example: `  dbxcli target list
  dbxcli target list entity_type=oracle_db`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			fmt.Printf("target list (filter: %v)\n", params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Register a new target",
		Long:  `Register a new target (database, host, or service) in the target registry.`,
		Example: `  dbxcli target add entity_name=prod-db entity_type=oracle_db host=db01.example.com port=1521 service=ORCL
  dbxcli target add entity_name=web01 entity_type=oracle_host host=web01.example.com`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			if params["entity_name"] == "" || params["entity_type"] == "" {
				return fmt.Errorf("entity_name and entity_type are required")
			}
			fmt.Printf("target add: %v\n", params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Test connectivity to a target",
		Long:  `Test network and authentication connectivity to a registered target.`,
		Example: `  dbxcli target test entity_name=prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			fmt.Printf("target test: %v\n", params)
			return nil
		},
	})
	return cmd
}

// NewServeCmd creates the "serve" subcommand for the REST API.
func NewServeCmd() *cobra.Command {
	var port int
	var authMode string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start REST API server",
		Long: `Start the dbx REST API server. Routes are auto-generated from the Cobra command tree.
The API exposes all CLI operations as HTTP endpoints with JWT authentication.`,
		Example: `  dbxcli serve
  dbxcli serve --port 9090 --auth-mode basic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("REST API server starting on :%d (auth: %s)\n", port, authMode)
			return nil
		},
	}
	cmd.Flags().IntVar(&port, "port", 8080, "listen port")
	cmd.Flags().StringVar(&authMode, "auth-mode", "jwt", "auth mode: jwt, basic, none")
	return cmd
}

// NewMCPCmd creates the "mcp" subcommand for the MCP server.
func NewMCPCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server",
		Long: `Start the dbx MCP (Model Context Protocol) server for AI integration.
Supports stdio transport (default) and SSE transport for remote connections.`,
		Example: `  dbxcli mcp                  # stdio transport (for Claude Code, IDE extensions)
  dbxcli mcp --port 3001      # SSE transport (for remote MCP clients)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if port > 0 {
				fmt.Printf("MCP SSE server starting on :%d\n", port)
			} else {
				fmt.Println("MCP stdio server starting")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&port, "port", 0, "SSE transport port (0 = stdio)")
	return cmd
}

// NewLicenseCmd creates the "license" subcommand group.
func NewLicenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Manage dbx license",
		Long: `Manage the dbx license. Licenses are Ed25519-signed JWT tokens stored at ~/.dbx/license.jwt.
Without a license, only OSS (Community) features are available.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show license status",
		Long:    `Show current license status including tier, entity count, expiry date, and grace period.`,
		Example: `  dbxcli license status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("license: OSS (no license file)")
			return nil
		},
	})
	activateCmd := &cobra.Command{
		Use:     "activate",
		Short:   "Activate a license file",
		Long:    `Activate a license JWT file. The file is validated, copied to ~/.dbx/license.jwt, and EE features are unlocked.`,
		Example: `  dbxcli license activate --file /path/to/license.jwt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			fmt.Printf("activating license from %s\n", file)
			return nil
		},
	}
	activateCmd.Flags().String("file", "", "path to license JWT file")
	cmd.AddCommand(activateCmd)
	return cmd
}
