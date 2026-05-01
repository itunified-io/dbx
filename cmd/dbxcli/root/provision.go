package root

import (
	"github.com/spf13/cobra"
)

// NewProvisionCmd returns the parent cobra.Command for all
// `dbxcli provision …` subcommands.
//
// Subcommands:
//   - install    Phase D install primitives (grid, dbhome, root-sh,
//                asmca, netca, asm-label)
//   - dbca       Database creation/configuration (create-db, …)
//   - pdb        PDB lifecycle (create, …)
//
// Per ADR-0094, all provision actions require Enterprise tier; the gate
// is enforced at each leaf subcommand's RunE (not here, so `--help`
// works without a license).
func NewProvisionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Database/host provisioning primitives (Enterprise)",
		Long: `Provisioning primitives invoked by /lab-up Phase D skills
(infrastructure repo). All actions require Enterprise license.

Use 'dbxcli provision install --help' for install primitives,
'dbxcli provision dbca --help' for database creation,
'dbxcli provision pdb --help' for PDB lifecycle.`,
	}
	cmd.PersistentFlags().String("target", "", "target name (from ~/.dbx/targets/)")
	cmd.AddCommand(NewInstallCmd())
	return cmd
}
