package root

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/provision/install"
	"github.com/spf13/cobra"
)

// NewInstallCmd returns the `provision install` parent subcommand.
// Each leaf delegates to a function in dbx/pkg/provision/install/.
//
// License gate: provision domain maps to BundleOps
// (pkg/core/license.DomainToBundle["provision"]). A RequireTier/RequireBundle
// helper does not yet exist in pkg/core/license — tracked in #519.
// When that helper ships, add:  license.RequireBundle(license.BundleOps)
// at the top of each RunE.
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Oracle install primitives (grid, dbhome, root-sh, asmca, netca, asm-label)",
	}
	cmd.AddCommand(newInstallGridCmd())
	cmd.AddCommand(newInstallDbhomeCmd())
	cmd.AddCommand(newInstallRootshCmd())
	cmd.AddCommand(newInstallAsmcaCmd())
	cmd.AddCommand(newInstallNetcaCmd())
	cmd.AddCommand(newInstallAsmLabelCmd())
	cmd.AddCommand(newInstallDbcaCmd())
	return cmd
}

// newInstallDbcaCmd: dbxcli provision install dbca --target X --oracle-home Y --oracle-base Z --response-file W --db-unique-name ORCL [--sys-password-file F] [--system-password-file F] [--reset]
func newInstallDbcaCmd() *cobra.Command {
	var (
		spec  install.DbcaCreateDbSpec
		reset bool
	)
	cmd := &cobra.Command{
		Use:   "dbca",
		Short: "Create CDB via dbca -silent -createDatabase (two-phase sentinel)",
		Long: `Create an Oracle CDB via dbca silent. Phase D.4 of /lab-up — runs after
Grid + DB Home + listener and before PDB creation / Data Guard standby
cloning.

Idempotency: NON-IDEMPOTENT primitive — uses a two-phase sentinel
(<oracle_base>/cfgtoollogs/dbx/dbca.<DB_UNIQUE_NAME>.partial → dbca.<DB_UNIQUE_NAME>.installed).
Detection ALSO probes ` + "`srvctl status database -d <unique>`" + ` so a
pre-existing database is recognised without forcing a sentinel.

The caller (skill) is responsible for rendering and SCPing the dbca .rsp
response file to the target host before invoking this command. The
response file format is multi-version; the skill picks the correct
template per Oracle release (19c / 23ai / 26ai).

Reset (MVP): --reset on installed/partial state prints a manual recovery
runbook to stderr and skips. The destructive ` + "`dbca -silent -deleteDatabase`" + `
step is deferred to a reverter follow-up plan.`,
		Example: `  dbxcli provision install dbca --target ext3adm1 \
    --oracle-home /u01/app/oracle/product/19c/dbhome_1 \
    --oracle-base /u01/app/oracle \
    --response-file /tmp/dbca.rsp \
    --db-unique-name ORCL`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			res, err := install.DbcaCreateDb(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$ORACLE_HOME containing bin/dbca + bin/srvctl")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE (sentinel root)")
	cmd.Flags().StringVar(&spec.ResponseFilePath, "response-file", "", "Absolute path on host to rendered dbca .rsp")
	cmd.Flags().StringVar(&spec.DbUniqueName, "db-unique-name", "", "DB_UNIQUE_NAME (used as sentinel key + srvctl probe target)")
	cmd.Flags().StringVar(&spec.SysPasswordFile, "sys-password-file", "", "Optional: absolute path to file on host containing SYS password (mode 0600)")
	cmd.Flags().StringVar(&spec.SystemPasswordFile, "system-password-file", "", "Optional: absolute path to file on host containing SYSTEM password (mode 0600)")
	cmd.Flags().BoolVar(&reset, "reset", false, "Print manual recovery runbook (NON-DESTRUCTIVE in MVP)")
	_ = cmd.MarkFlagRequired("oracle-home")
	_ = cmd.MarkFlagRequired("oracle-base")
	_ = cmd.MarkFlagRequired("response-file")
	_ = cmd.MarkFlagRequired("db-unique-name")
	return cmd
}

// newInstallAsmLabelCmd: dbxcli provision install asm-label --target X --grid-home Y --oracle-base Z --impl asmlib|afd --labels DATA1:/dev/sdb,DATA2:/dev/sdc [--reset]
func newInstallAsmLabelCmd() *cobra.Command {
	var (
		spec       install.AsmDiskLabelSpec
		labelsFlag string
		reset      bool
	)
	cmd := &cobra.Command{
		Use:   "asm-label",
		Short: "Label raw disks for ASM discovery (asmlib or AFD; per-label two-phase sentinel)",
		Long: `Label raw block devices via ASMlib (oracleasm) or Oracle ASM Filter
Driver (AFD) so the disks become discoverable as ASM disks. This is a
Phase D.1 prerequisite that runs BEFORE asmca (which creates the
diskgroup over labeled devices).

Idempotency: NON-IDEMPOTENT primitive — uses a per-label two-phase
sentinel (<oracle_base>/cfgtoollogs/dbx/asm-label.<NAME>.partial →
asm-label.<NAME>.installed). Detection ALSO probes oracleasm listdisks
(asmlib) or asmcmd afd_lslbl <device> (afd) so a pre-existing label is
recognised without forcing a sentinel.

Reset (MVP): --reset on installed/partial state for any label prints a
manual recovery runbook to stderr and skips that label. The destructive
label-removal step is deferred to a reverter follow-up plan.`,
		Example: `  dbxcli provision install asm-label --target ext3adm1 \
    --grid-home /u01/app/19c/grid \
    --oracle-base /u01/app/grid \
    --impl asmlib \
    --labels DATA1:/dev/sdb,DATA2:/dev/sdc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			if labelsFlag == "" {
				return fmt.Errorf("--labels is required (comma-separated NAME:DEVICE pairs)")
			}
			for _, pair := range strings.Split(labelsFlag, ",") {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("--labels pair %q must be NAME:DEVICE", pair)
				}
				spec.Labels = append(spec.Labels, install.AsmLabelEntry{
					Name:   parts[0],
					Device: parts[1],
				})
			}
			res, err := install.AsmDiskLabel(context.Background(), spec, reset)
			if err != nil {
				// Print partial result before returning so operators see
				// per-label state on failure.
				if res != nil {
					out, _ := json.MarshalIndent(res, "", "  ")
					fmt.Println(string(out))
				}
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.GridHome, "grid-home", "", "$GRID_HOME containing bin/asmcmd (used for AFD; required for both impls)")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE for grid (sentinel root)")
	cmd.Flags().StringVar(&spec.Implementation, "impl", install.AsmDiskLabelImplAFD, "Labeling implementation: asmlib | afd")
	cmd.Flags().StringVar(&labelsFlag, "labels", "", "Comma-separated NAME:DEVICE pairs (e.g. DATA1:/dev/sdb,DATA2:/dev/sdc)")
	cmd.Flags().BoolVar(&reset, "reset", false, "Print manual recovery runbook (NON-DESTRUCTIVE in MVP)")
	_ = cmd.MarkFlagRequired("grid-home")
	_ = cmd.MarkFlagRequired("oracle-base")
	_ = cmd.MarkFlagRequired("labels")
	return cmd
}

// newInstallNetcaCmd: dbxcli provision install netca --target X --oracle-home Y --oracle-base Z --response-file W [--listener-name LISTENER] [--port 1521] [--reset]
func newInstallNetcaCmd() *cobra.Command {
	var (
		spec  install.NetcaSpec
		reset bool
	)
	cmd := &cobra.Command{
		Use:   "netca",
		Short: "Create Oracle listener via netca -silent (two-phase sentinel)",
		Long: `Create an Oracle Net listener via netca silent. Used during Phase D.2
(post-Grid, pre-DBCA) to ensure a LISTENER exists for client connections
AND during Phase E.2 to add static services on a standby for RMAN
DUPLICATE FROM ACTIVE.

Idempotency: NON-IDEMPOTENT primitive — uses a two-phase sentinel
(<oracle_base>/cfgtoollogs/dbx/netca.<LISTENER>.partial → netca.<LISTENER>.installed).
Detection ALSO probes lsnrctl status so a pre-existing listener is
recognised without forcing a sentinel.

Reset (MVP): --reset on installed/partial state prints a manual recovery
runbook to stderr and skips. The destructive listener-drop step is
deferred to a reverter follow-up plan.`,
		Example: `  dbxcli provision install netca --target ext3adm1 \
    --oracle-home /u01/app/oracle/product/19c/dbhome_1 \
    --oracle-base /u01/app/oracle \
    --response-file /tmp/netca.rsp \
    --listener-name LISTENER --port 1521`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			res, err := install.NetcaSilent(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$ORACLE_HOME containing bin/netca + bin/lsnrctl")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE (sentinel root)")
	cmd.Flags().StringVar(&spec.ResponseFilePath, "response-file", "", "Absolute path on host to rendered netca .rsp")
	cmd.Flags().StringVar(&spec.ListenerName, "listener-name", "LISTENER", "Listener name")
	cmd.Flags().IntVar(&spec.Port, "port", 1521, "TCP listening port")
	cmd.Flags().BoolVar(&reset, "reset", false, "Print manual recovery runbook (NON-DESTRUCTIVE in MVP)")
	_ = cmd.MarkFlagRequired("oracle-home")
	_ = cmd.MarkFlagRequired("oracle-base")
	_ = cmd.MarkFlagRequired("response-file")
	return cmd
}

// newInstallAsmcaCmd: dbxcli provision install asmca --target X --oracle-home Y --oracle-base Z --dg-name DATA --disks /dev/sdb,/dev/sdc [--redundancy EXTERNAL] [--au-size-mb 4] [--reset]
func newInstallAsmcaCmd() *cobra.Command {
	var (
		spec       install.AsmcaSpec
		disksFlag  string
		reset      bool
	)
	cmd := &cobra.Command{
		Use:   "asmca",
		Short: "Create initial ASM diskgroup via asmca -silent (two-phase sentinel)",
		Long: `Create the initial ASM diskgroup (DATA, RECO) for an Oracle Grid Infrastructure
installation. Subsequent diskgroup operations should go through the
mcp-oracle-ee-asm tools, which assume ASM is already up.

Idempotency: NON-IDEMPOTENT primitive — uses a two-phase sentinel
(<oracle_base>/cfgtoollogs/dbx/asmca.<DG>.partial → asmca.<DG>.installed).
Detection is version-agnostic.

Reset (MVP): --reset on installed/partial state prints a manual recovery
runbook to stderr and skips. The destructive ` + "`drop diskgroup`" + ` step is
deferred to a reverter follow-up plan.`,
		Example: `  dbxcli provision install asmca --target ext3adm1 \
    --oracle-home /u01/app/19c/grid \
    --oracle-base /u01/app/grid \
    --dg-name DATA --disks /dev/sdb,/dev/sdc \
    --redundancy EXTERNAL --au-size-mb 4`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			if disksFlag == "" {
				return fmt.Errorf("--disks is required (comma-separated)")
			}
			spec.Disks = strings.Split(disksFlag, ",")
			res, err := install.AsmcaSilent(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$GRID_HOME containing bin/asmca")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE for grid (sentinel root)")
	cmd.Flags().StringVar(&spec.DGName, "dg-name", "", "Diskgroup name (e.g. DATA)")
	cmd.Flags().StringVar(&spec.Redundancy, "redundancy", "EXTERNAL", "EXTERNAL | NORMAL | HIGH")
	cmd.Flags().IntVar(&spec.AUSizeMB, "au-size-mb", 4, "Allocation Unit size in MB")
	cmd.Flags().StringVar(&disksFlag, "disks", "", "Comma-separated disk paths or AFD labels")
	cmd.Flags().BoolVar(&reset, "reset", false, "Print manual recovery runbook (NON-DESTRUCTIVE in MVP)")
	_ = cmd.MarkFlagRequired("oracle-home")
	_ = cmd.MarkFlagRequired("oracle-base")
	_ = cmd.MarkFlagRequired("dg-name")
	return cmd
}

// newInstallDbhomeCmd: dbxcli provision install dbhome --target X --oracle-home Y --software-staging Z --response-file W [--reset]
func newInstallDbhomeCmd() *cobra.Command {
	var (
		spec  install.InstallSpec
		reset bool
	)
	cmd := &cobra.Command{
		Use:   "dbhome",
		Short: "Run runInstaller -silent for Oracle DB Home 19c",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			// spec.Target is bound directly via --target flag; inherit from parent
			// persistent flag if not set on this leaf.
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			res, err := install.DBHomeInstall(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$ORACLE_HOME (e.g. /u01/app/oracle/product/19c/dbhome_1)")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE (e.g. /u01/app/oracle)")
	cmd.Flags().StringVar(&spec.SoftwareStaging, "software-staging", "", "Path on host where DB home software is unzipped")
	cmd.Flags().StringVar(&spec.ResponseFilePath, "response-file", "", "Absolute path on host to rendered .rsp file")
	cmd.Flags().BoolVar(&reset, "reset", false, "Reset prior install state (NOT IMPLEMENTED YET)")
	_ = cmd.MarkFlagRequired("oracle-home")
	return cmd
}

// newInstallRootshCmd: dbxcli provision install root-sh --target X --oracle-home Y [--reset]
func newInstallRootshCmd() *cobra.Command {
	var (
		spec  install.InstallSpec
		reset bool
	)
	cmd := &cobra.Command{
		Use:   "root-sh",
		Short: "Run <OracleHome>/root.sh idempotently after a runInstaller",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#519): wire license.RequireBundle("provision") once helper ships
			if spec.Target == "" {
				if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			res, err := install.RootSh(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$ORACLE_HOME or $GRID_HOME containing root.sh")
	cmd.Flags().BoolVar(&reset, "reset", false, "Re-run root.sh even if touchfile exists (root.sh is idempotent)")
	_ = cmd.MarkFlagRequired("oracle-home")
	return cmd
}

// newInstallGridCmd: dbxcli provision install grid --target X --oracle-home Y ...
func newInstallGridCmd() *cobra.Command {
	var (
		spec  install.InstallSpec
		reset bool
	)
	cmd := &cobra.Command{
		Use:   "grid",
		Short: "Run runInstaller -silent for Oracle Grid Infrastructure 19c",
		Long: `Run Oracle Grid Infrastructure 19c runInstaller in silent mode.

The caller (skill) is responsible for rendering and SCPing the response file
to the target host before invoking this command.

Idempotency:
  - /etc/oraInst.loc + $GRID_HOME/inventory both present → skipped (Detected=installed)
  - Only one present                                      → error (partial install)
  - Neither present                                       → runs runInstaller`,
		Example: `  dbxcli provision install grid --target ext3adm1 \
    --oracle-home /u01/app/19c/grid \
    --oracle-base /u01/app/grid \
    --response-file /tmp/grid.rsp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Inherit --target from the provision parent persistent flag if
			// not set directly on this leaf.
			if spec.Target == "" {
				if t, _ := cmd.Flags().GetString("target"); t != "" {
					spec.Target = t
				} else if pt := cmd.InheritedFlags().Lookup("target"); pt != nil {
					spec.Target = pt.Value.String()
				}
			}
			res, err := install.GridInstall(context.Background(), spec, reset)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&spec.Target, "target", "", "target name (overrides provision --target)")
	cmd.Flags().StringVar(&spec.OracleHome, "oracle-home", "", "$GRID_HOME (e.g. /u01/app/19c/grid)")
	cmd.Flags().StringVar(&spec.OracleBase, "oracle-base", "", "$ORACLE_BASE for grid (e.g. /u01/app/grid)")
	cmd.Flags().StringVar(&spec.SoftwareStaging, "software-staging", "", "Path on host where Grid software is unzipped")
	cmd.Flags().StringVar(&spec.ResponseFilePath, "response-file", "", "Absolute path on host to rendered .rsp file")
	cmd.Flags().BoolVar(&reset, "reset", false, "Reset prior install state (NOT IMPLEMENTED YET — see #519 reverter follow-up)")
	_ = cmd.MarkFlagRequired("oracle-home")
	return cmd
}
