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
	cmd.AddCommand(NewServeCmd())
	cmd.AddCommand(NewMCPCmd())
	cmd.AddCommand(NewLicenseCmd())

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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List registered targets",
		Args:  cobra.ArbitraryArgs,
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
		Args:  cobra.MinimumNArgs(2),
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
		Args:  cobra.MinimumNArgs(1),
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show license status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("license: OSS (no license file)")
			return nil
		},
	})
	activateCmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a license file",
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
