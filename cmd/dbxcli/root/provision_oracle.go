package root

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/itunified-io/dbx/pkg/provision/oracle"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewOracleCmd returns the `provision oracle` subtree — the orchestrator that
// turns a single infrastructure `kind: OracleDatabase` (DbSys) manifest into
// the ordered sequence of install primitives (grid/dbhome/root-sh/asmca/
// asm-label/netca/dbca/pdb) across the cluster's nodes.
//
// MVP scope: `plan` (read-only) derives and prints the sequence. It is NOT
// license-gated — it touches no infrastructure and reveals nothing sensitive,
// matching how `--help` works without a license. The future `apply` execution
// subcommand WILL gate via license.RequireBundle("provision") (ADR-0094),
// since that drives the primitives that actually provision.
func NewOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Orchestrate Oracle install primitives from an OracleDatabase (DbSys) manifest",
	}
	cmd.AddCommand(newOraclePlanCmd())
	return cmd
}

func newOraclePlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan <dbsys.yaml>",
		Short: "Derive the ordered provisioning sequence from an OracleDatabase manifest (read-only)",
		Long: `Read a kind: OracleDatabase (DbSys) manifest and print the ordered
install-primitive sequence that would provision it across the cluster's nodes.

This is a dry, read-only preview — it executes nothing. Each step maps 1:1
onto a 'dbxcli provision install <primitive>' call.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := oracle.LoadManifest(args[0])
			if err != nil {
				return err
			}
			steps, err := oracle.Plan(m)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				return json.NewEncoder(cmd.OutOrStdout()).Encode(steps)
			case "yaml":
				return yaml.NewEncoder(cmd.OutOrStdout()).Encode(steps)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Provisioning plan for %s (%s, %d node(s), %d step(s)):\n\n",
					m.Metadata.Name, m.Spec.Topology, len(m.Spec.NodesRef), len(steps))
				tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(tw, "#\tPRIMITIVE\tSCOPE\tTARGET\tDETAIL")
				for _, s := range steps {
					fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", s.Order, s.Primitive, s.Scope, s.Target, s.Detail)
				}
				return tw.Flush()
			}
		},
	}
}
