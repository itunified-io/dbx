package root

import (
	"context"
	"encoding/json"
	"fmt"

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
