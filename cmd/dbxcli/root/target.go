package root

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewTargetCmd creates the "target" subcommand group.
//
// Targets are persisted as YAML files under ~/.dbx/targets/<name>.yaml
// (mode 0600). The file format is the same as pkg/core/target.Target.
func NewTargetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage system targets",
		Long: `Manage the target registry — databases, hosts, and services that dbx connects to.
Targets are stored in ~/.dbx/targets/ as YAML files (mode 0600).`,
	}

	cmd.AddCommand(newTargetAddCmd())
	cmd.AddCommand(newTargetListCmd())
	cmd.AddCommand(newTargetTestCmd())
	cmd.AddCommand(newTargetRemoveCmd())
	return cmd
}

// targetFromParams maps emcli-style key=value params onto a Target struct.
// Recognised params:
//
//	entity_name   → Target.Name (required)
//	entity_type   → Target.Type (required)
//	description   → Target.Description
//	host          → SSH.Host AND Primary.Host
//	port          → Primary.Port
//	service       → Primary.Service
//	database      → Primary.Database
//	ssh_user      → SSH.User
//	ssh_key_path  → SSH.KeyPath
//	vault_path    → SSH.VaultPath / Primary.VaultPath
func targetFromParams(p map[string]string) (*target.Target, error) {
	if p["entity_name"] == "" || p["entity_type"] == "" {
		return nil, fmt.Errorf("entity_name and entity_type are required")
	}
	t := &target.Target{
		Name:        p["entity_name"],
		Type:        target.EntityType(p["entity_type"]),
		Description: p["description"],
	}

	// SSH section — populated when host/ssh_* keys present.
	if p["host"] != "" || p["ssh_user"] != "" || p["ssh_key_path"] != "" {
		t.SSH = &target.SSHConfig{
			Host:      p["host"],
			User:      p["ssh_user"],
			KeyPath:   p["ssh_key_path"],
			VaultPath: p["vault_path"],
		}
	}

	// Primary endpoint — populated when port/service/database set.
	if p["port"] != "" || p["service"] != "" || p["database"] != "" {
		ep := &target.Endpoint{
			Host:      p["host"],
			Service:   p["service"],
			Database:  p["database"],
			VaultPath: p["vault_path"],
		}
		if p["port"] != "" {
			n, err := strconv.Atoi(p["port"])
			if err != nil {
				return nil, fmt.Errorf("invalid port %q: %w", p["port"], err)
			}
			ep.Port = n
		}
		t.Primary = ep
	}

	return t, nil
}

func newTargetAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Register a new target",
		Long:  `Register a new target (database, host, or service) in the target registry. Persisted to ~/.dbx/targets/<entity_name>.yaml.`,
		Example: `  dbxcli target add entity_name=prod-db entity_type=oracle_database host=db01.example.com port=1521 service=ORCL
  dbxcli target add entity_name=web01 entity_type=oracle_host host=web01.example.com ssh_user=root ssh_key_path=/home/me/.ssh/id_ed25519`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			t, err := targetFromParams(params)
			if err != nil {
				return err
			}
			if err := target.Save(t); err != nil {
				return err
			}
			path := filepath.Join(target.StoreDir(), t.Name+".yaml")
			cmd.PrintErrf("target added: %s (type=%s) → %s\n", t.Name, t.Type, path)
			return nil
		},
	}
}

func newTargetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered targets",
		Long:  `List all registered targets from ~/.dbx/targets/.`,
		Example: `  dbxcli target list
  dbxcli target list entity_type=oracle_database`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			items, err := target.List()
			if err != nil {
				return err
			}
			// Optional entity_type filter.
			if et := params["entity_type"]; et != "" {
				filtered := items[:0]
				for _, t := range items {
					if string(t.Type) == et {
						filtered = append(filtered, t)
					}
				}
				items = filtered
			}

			format, _ := cmd.Root().PersistentFlags().GetString("format")
			switch format {
			case "json":
				out, _ := json.MarshalIndent(items, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(out))
			case "yaml":
				out, _ := yaml.Marshal(items)
				fmt.Fprint(cmd.OutOrStdout(), string(out))
			default:
				if len(items) == 0 {
					cmd.PrintErrln("(no targets registered)")
					return nil
				}
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tTYPE\tHOST\tDESCRIPTION")
				for _, t := range items {
					host := ""
					if t.SSH != nil {
						host = t.SSH.Host
					} else if t.Primary != nil {
						host = t.Primary.Host
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.Type, host, t.Description)
				}
				_ = w.Flush()
			}
			return nil
		},
	}
}

func newTargetTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "test",
		Short:   "Test SSH connectivity to a target",
		Long:    `Test SSH connectivity to a registered target by running 'whoami'.`,
		Example: `  dbxcli target test entity_name=prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			name := params["entity_name"]
			if name == "" {
				return fmt.Errorf("entity_name is required")
			}
			t, err := target.Load(name)
			if err != nil {
				return err
			}
			if t.SSH == nil || t.SSH.Host == "" {
				return fmt.Errorf("target %q has no SSH config", name)
			}
			user := t.SSH.User
			if user == "" {
				user = "root"
			}
			dest := fmt.Sprintf("%s@%s", user, t.SSH.Host)

			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			sshArgs := []string{
				"-o", "BatchMode=yes",
				"-o", "ConnectTimeout=5",
				"-o", "StrictHostKeyChecking=accept-new",
			}
			if t.SSH.KeyPath != "" {
				sshArgs = append(sshArgs, "-i", t.SSH.KeyPath)
			}
			sshArgs = append(sshArgs, dest, "whoami")

			out, err := exec.CommandContext(ctx, "ssh", sshArgs...).CombinedOutput() //nolint:gosec
			result := strings.TrimSpace(string(out))
			if err != nil {
				cmd.PrintErrf("target test FAILED: %s (%s): %v\noutput: %s\n", name, dest, err, result)
				return fmt.Errorf("ssh failed: %w", err)
			}
			cmd.PrintErrf("target test OK: %s (%s) → whoami=%s\n", name, dest, result)
			return nil
		},
	}
}

func newTargetRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a target from the registry",
		Long:    `Delete the YAML file for a target. Idempotent: succeeds even if the target does not exist.`,
		Example: `  dbxcli target remove prod-db
  dbxcli target remove entity_name=prod-db`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Accept either positional name or entity_name=<name>.
			var name string
			if strings.Contains(args[0], "=") {
				params, err := ParseNamedParams(args)
				if err != nil {
					return err
				}
				name = params["entity_name"]
			} else {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("target name is required")
			}
			if err := target.Remove(name); err != nil {
				return err
			}
			cmd.PrintErrf("target removed: %s\n", name)
			return nil
		},
	}
}
