package root

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/itunified-io/dbx/pkg/license"
	"github.com/itunified-io/dbx/pkg/provision/install"
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
	cmd.AddCommand(newOracleApplyCmd())
	return cmd
}

// newOracleApplyCmd: dbxcli provision oracle apply <dbsys.yaml> [--execute]
//
// Default is a DRY-RUN (no infrastructure touched, no license required) that
// prints the resolved per-step install specs. With --execute it runs the
// sequence via the install primitives, in order, stopping at the first error;
// --execute is license-gated (Enterprise, ADR-0094). Secrets, raw disks, and
// response files are operator-supplied via flags — never read from the manifest.
func newOracleApplyCmd() *cobra.Command {
	var (
		execute      bool
		reset        bool
		asmImpl      string
		disks        map[string]string
		asmLabels    []string
		gridRsp      string
		dbhomeRsp    string
		dbcaRsp      string
		sysPwFile    string
		systemPwFile string
		pdbPwFile    string
		listenerName string
		listenerPort int
		softwareRoot string
	)
	cmd := &cobra.Command{
		Use:   "apply <dbsys.yaml>",
		Short: "Run (or dry-run) the provisioning sequence derived from an OracleDatabase manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if execute {
				if err := license.RequireBundle("provision"); err != nil {
					return err
				}
			}
			m, err := oracle.LoadManifest(args[0])
			if err != nil {
				return err
			}
			opts := oracle.ApplyOptions{
				ASMImplementation:    asmImpl,
				DisksByTag:           splitDisks(disks),
				ASMLabels:            parseAsmLabels(asmLabels),
				GridResponseFile:     gridRsp,
				DBHomeResponseFile:   dbhomeRsp,
				DbcaResponseFile:     dbcaRsp,
				SysPasswordFile:      sysPwFile,
				SystemPasswordFile:   systemPwFile,
				PdbAdminPasswordFile: pdbPwFile,
				ListenerName:         listenerName,
				ListenerPort:         listenerPort,
				SoftwareRoot:         softwareRoot,
			}
			results, applyErr := oracle.Apply(context.Background(), m, opts, execute, reset)

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(results)
			case "yaml":
				_ = yaml.NewEncoder(cmd.OutOrStdout()).Encode(results)
			default:
				mode := "DRY-RUN (no changes; pass --execute to run)"
				if execute {
					mode = "EXECUTE"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "apply %s — %s (%d step(s)):\n\n", m.Metadata.Name, mode, len(results))
				tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(tw, "#\tPRIMITIVE\tTARGET\tEXECUTED\tRESULT")
				for _, r := range results {
					status := "-"
					if r.Err != "" {
						status = "ERROR: " + r.Err
					} else if r.Executed {
						status = "ok"
					}
					fmt.Fprintf(tw, "%d\t%s\t%s\t%v\t%s\n", r.Step.Order, r.Step.Primitive, r.Step.Target, r.Executed, status)
				}
				_ = tw.Flush()
			}
			return applyErr
		},
	}
	cmd.Flags().BoolVar(&execute, "execute", false, "actually run the sequence (default: dry-run); Enterprise-gated")
	cmd.Flags().BoolVar(&reset, "reset", false, "pass --reset through to each primitive")
	cmd.Flags().StringVar(&asmImpl, "asm-impl", "", "ASM implementation: asmlib|afd (default asmlib)")
	cmd.Flags().StringToStringVar(&disks, "disks", nil, "ASM devices per diskgroup tag, e.g. asm-data=/dev/sdb,/dev/sdc")
	cmd.Flags().StringArrayVar(&asmLabels, "asm-label", nil, "raw-disk label as name=device (repeatable)")
	cmd.Flags().StringVar(&gridRsp, "grid-response-file", "", "grid response file path on target")
	cmd.Flags().StringVar(&dbhomeRsp, "dbhome-response-file", "", "db home response file path on target")
	cmd.Flags().StringVar(&dbcaRsp, "dbca-response-file", "", "dbca response file path on target")
	cmd.Flags().StringVar(&sysPwFile, "sys-password-file", "", "SYS password file path on target")
	cmd.Flags().StringVar(&systemPwFile, "system-password-file", "", "SYSTEM password file path on target")
	cmd.Flags().StringVar(&pdbPwFile, "pdb-admin-password-file", "", "PDB admin password file path on target (required to execute pdb steps)")
	cmd.Flags().StringVar(&listenerName, "listener-name", "", "listener name (default LISTENER)")
	cmd.Flags().IntVar(&listenerPort, "listener-port", 0, "listener port (default 1521)")
	cmd.Flags().StringVar(&softwareRoot, "software-root", "", "override the manifest software_staging.source")
	return cmd
}

// splitDisks converts cobra StringToString (tag=dev1,dev2) into the
// map[tag][]device shape ApplyOptions expects.
func splitDisks(in map[string]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for tag, csv := range in {
		for _, d := range strings.Split(csv, ",") {
			if d = strings.TrimSpace(d); d != "" {
				out[tag] = append(out[tag], d)
			}
		}
	}
	return out
}

// parseAsmLabels converts "name=device" strings into AsmLabelEntry values.
func parseAsmLabels(in []string) []install.AsmLabelEntry {
	var out []install.AsmLabelEntry
	for _, s := range in {
		if name, dev, ok := strings.Cut(s, "="); ok {
			out = append(out, install.AsmLabelEntry{Name: strings.TrimSpace(name), Device: strings.TrimSpace(dev)})
		}
	}
	return out
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
