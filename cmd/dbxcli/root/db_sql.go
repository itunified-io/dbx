package root

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/license"
	oraclesql "github.com/itunified-io/dbx/pkg/oracle/sql"
	"github.com/spf13/cobra"
)

// newDBSQLExecReadWriteCmd: dbxcli db sql exec-readwrite --target X
//   --sql "<stmt>" | --sql-file <path> --oracle-sid ORCLPRI
//   --oracle-home /u01/app/oracle/product/19c/dbhome_1 [--format json]
//
// Privileged sibling of `db sql exec` — executes DDL/DML/anonymous PL/SQL
// via sqlplus / as sysdba over SSH. Required for /lab-up Phase E.1
// (FORCE LOGGING, FLASHBACK ON, ADD STANDBY LOGFILE) and Phase E.2
// (RECOVER MANAGED STANDBY DATABASE USING CURRENT LOGFILE DISCONNECT).
//
// License gate: provision bundle (Enterprise tier) — same gate as install
// primitives. Bundle gate implicitly requires Enterprise tier per
// pkg/license/require.go.
func newDBSQLExecReadWriteCmd() *cobra.Command {
	var (
		target     string
		sqlInline  string
		sqlFile    string
		oracleSID  string
		oracleHome string
		formatFlag string
		logTail    int
	)
	cmd := &cobra.Command{
		Use:   "exec-readwrite",
		Short: "Execute privileged DDL/DML/PL-SQL via sqlplus / as sysdba (Enterprise tier)",
		Long: `Execute one or more DDL/DML/anonymous-PL/SQL statements against an Oracle
database via sqlplus connected as sysdba over SSH on the host.

Use this for privileged operations that the read-only ` + "`db sql exec`" + ` cannot run:
  - ALTER DATABASE FORCE LOGGING
  - ALTER DATABASE FLASHBACK ON
  - ALTER DATABASE ADD STANDBY LOGFILE
  - DBMS_DATAGUARD_BROKER package calls
  - GRANT / REVOKE / CREATE USER

Phase E.1 + Phase E.2 of /lab-up depend on this command.

License: requires the provision bundle (Enterprise tier).`,
		Example: `  dbxcli db sql exec-readwrite --target ext3adm1 \
    --oracle-sid ORCLPRI \
    --oracle-home /u01/app/oracle/product/19c/dbhome_1 \
    --sql "ALTER DATABASE FORCE LOGGING;"

  dbxcli db sql exec-readwrite --target ext3adm1 \
    --oracle-sid ORCLPRI \
    --oracle-home /u01/app/oracle/product/19c/dbhome_1 \
    --sql-file /tmp/dg-prepare-primary.sql --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Tier-gate: provision bundle (Enterprise tier).
			if err := license.RequireBundle("provision"); err != nil {
				return fmt.Errorf("dbxcli db sql exec-readwrite: %w", err)
			}

			t, err := resolveTarget(cmd, target)
			if err != nil {
				return err
			}

			if (sqlInline == "" && sqlFile == "") || (sqlInline != "" && sqlFile != "") {
				return fmt.Errorf("exactly one of --sql or --sql-file must be set")
			}
			stmt := sqlInline
			if sqlFile != "" {
				b, err := os.ReadFile(sqlFile) //nolint:gosec
				if err != nil {
					return fmt.Errorf("read --sql-file %s: %w", sqlFile, err)
				}
				stmt = string(b)
			}

			res, err := oraclesql.ExecReadWrite(context.Background(), t, stmt, oraclesql.ExecOptions{
				OracleSID:    oracleSID,
				OracleHome:   oracleHome,
				LogTailLines: logTail,
			})
			// Render best-effort even on error so operators see exit code + stderr.
			switch strings.ToLower(formatFlag) {
			case "json":
				if res != nil {
					out, _ := json.MarshalIndent(res, "", "  ")
					fmt.Println(string(out))
				}
			default:
				if res != nil {
					fmt.Printf("exit_code: %d\n", res.ExitCode)
					if res.Stdout != "" {
						fmt.Printf("--- stdout ---\n%s", res.Stdout)
					}
					if res.Stderr != "" {
						fmt.Printf("--- stderr ---\n%s", res.Stderr)
					}
				}
			}
			return err
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "target name (overrides db --target)")
	cmd.Flags().StringVar(&sqlInline, "sql", "", "SQL to execute (mutually exclusive with --sql-file)")
	cmd.Flags().StringVar(&sqlFile, "sql-file", "", "path to file containing SQL (mutually exclusive with --sql)")
	cmd.Flags().StringVar(&oracleSID, "oracle-sid", "", "$ORACLE_SID for sqlplus session")
	cmd.Flags().StringVar(&oracleHome, "oracle-home", "", "$ORACLE_HOME containing bin/sqlplus")
	cmd.Flags().StringVar(&formatFlag, "format", "table", "output format: table | json")
	cmd.Flags().IntVar(&logTail, "log-tail", 0, "trailing N lines of output to capture in log_tail (0 = full)")
	_ = cmd.MarkFlagRequired("oracle-sid")
	_ = cmd.MarkFlagRequired("oracle-home")
	return cmd
}
