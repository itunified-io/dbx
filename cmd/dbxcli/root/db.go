package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewDBCmd creates the "db" parent command for Oracle database operations.
func NewDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Oracle database read-only operations",
		Long: `Read-only Oracle database operations covering sessions, tablespaces,
users, schemas, SQL execution, redo logs, undo, parameters, and advisors.

Requires a target with an Oracle database endpoint configured.`,
	}

	cmd.PersistentFlags().String("target", "", "target name (from ~/.dbx/targets/)")

	cmd.AddCommand(newDBSessionCmd())
	cmd.AddCommand(newDBTablespaceCmd())
	cmd.AddCommand(newDBUserCmd())
	cmd.AddCommand(newDBSchemaCmd())
	cmd.AddCommand(newDBSQLCmd())
	cmd.AddCommand(newDBRedoCmd())
	cmd.AddCommand(newDBUndoCmd())
	cmd.AddCommand(newDBParameterCmd())
	cmd.AddCommand(newDBAdvisorCmd())

	return cmd
}

func newDBSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Oracle session operations",
		Long:  `Query and inspect Oracle database sessions (V$SESSION). Read-only — no session modification.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active user sessions",
		Long:  `List all active user sessions from V$SESSION excluding background processes.`,
		Example: `  dbxcli db session list --target prod-db
  dbxcli db session list --target prod-db --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db session list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe a session by SID",
		Long:  `Show detailed information for a specific session including SQL, wait event, and program.`,
		Example: `  dbxcli db session describe sid=142 --target prod-db`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db session describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "top-waiters",
		Short: "Show top sessions by wait time",
		Long:  `Show sessions with the highest cumulative wait time, useful for identifying performance bottlenecks.`,
		Example: `  dbxcli db session top-waiters --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db session top-waiters (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBTablespaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tablespace",
		Short: "Oracle tablespace operations",
		Long:  `Query Oracle tablespace metadata, usage metrics, and datafile details (DBA_TABLESPACES, DBA_DATA_FILES). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tablespaces with usage metrics",
		Long:  `List all tablespaces with total size, used space, free space, and percentage utilization.`,
		Example: `  dbxcli db tablespace list --target prod-db
  dbxcli db tablespace list --target prod-db --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db tablespace list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe tablespace datafiles",
		Long:  `Show datafiles belonging to a tablespace with sizes, autoextend status, and paths.`,
		Example: `  dbxcli db tablespace describe name=USERS --target prod-db`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db tablespace describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "usage",
		Short: "Show aggregated tablespace usage summary",
		Long:  `Show aggregated storage utilization across all tablespaces with alert thresholds.`,
		Example: `  dbxcli db tablespace usage --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db tablespace usage (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Oracle user operations",
		Long:  `Query Oracle database users, roles, privileges, and profiles (DBA_USERS, DBA_ROLE_PRIVS). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List database users",
		Long:    `List all database users with account status, default tablespace, and profile.`,
		Example: `  dbxcli db user list --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db user list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "describe",
		Short:   "Describe user with roles and privileges",
		Long:    `Show detailed user information including granted roles, system/object privileges, and quotas.`,
		Example: `  dbxcli db user describe username=SCOTT --target prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db user describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "profiles",
		Short:   "List database profiles",
		Long:    `List all database profiles with their resource limits and password parameters.`,
		Example: `  dbxcli db user profiles --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db user profiles (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Oracle schema browser",
		Long:  `Browse Oracle schemas and their objects (DBA_OBJECTS, DBA_TABLES, DBA_INDEXES). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List schemas with object counts",
		Long:    `List all schemas with counts of tables, indexes, views, packages, and other objects.`,
		Example: `  dbxcli db schema list --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db schema list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "objects",
		Short:   "List objects in a schema",
		Long:    `List all objects in a given schema filtered by object type.`,
		Example: `  dbxcli db schema objects owner=HR --target prod-db
  dbxcli db schema objects owner=HR object_type=TABLE --target prod-db`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db schema objects (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "describe",
		Short:   "Describe a specific object",
		Long:    `Show DDL and metadata for a specific database object (table, index, view, package, etc.).`,
		Example: `  dbxcli db schema describe owner=HR object_name=EMPLOYEES --target prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db schema describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

func newDBSQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql",
		Short: "Read-only SQL execution",
		Long: `Execute read-only SQL statements against an Oracle database.
Only SELECT statements are permitted — DML/DDL is blocked by the ReadOnlyGuard.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "exec",
		Short: "Execute a read-only SELECT statement",
		Long:  `Execute a SELECT query and return results. DML/DDL statements are rejected.`,
		Example: `  dbxcli db sql exec query="SELECT sysdate FROM dual" --target prod-db
  dbxcli db sql exec query="SELECT username, account_status FROM dba_users" --target prod-db --format json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db sql exec (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "explain",
		Short: "Generate execution plan for a SELECT",
		Long:  `Generate EXPLAIN PLAN output for a SELECT statement without executing it.`,
		Example: `  dbxcli db sql explain query="SELECT * FROM hr.employees WHERE department_id = 10" --target prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db sql explain (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

func newDBRedoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redo",
		Short: "Oracle redo log operations",
		Long:  `Query Oracle redo log groups and switch history (V$LOG, V$LOG_HISTORY). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List redo log groups",
		Long:    `List all redo log groups with status, size, member count, and sequence number.`,
		Example: `  dbxcli db redo list --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db redo list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "switch-history",
		Short:   "Show log switch frequency",
		Long:    `Show redo log switch frequency per hour — useful for sizing redo logs.`,
		Example: `  dbxcli db redo switch-history --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db redo switch-history (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBUndoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Oracle undo/rollback operations",
		Long:  `Query Oracle undo tablespace usage and rollback segment details (V$UNDOSTAT, DBA_UNDO_EXTENTS). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List undo tablespace usage",
		Long:    `Show undo tablespace utilization including active, unexpired, and expired extents.`,
		Example: `  dbxcli db undo list --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db undo list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "segments",
		Short:   "Show undo segment details",
		Long:    `Show individual rollback segment status, size, and transaction counts.`,
		Example: `  dbxcli db undo segments --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db undo segments (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBParameterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parameter",
		Short: "Oracle init parameter operations",
		Long:  `Query Oracle initialization parameters (V$PARAMETER, V$SYSTEM_PARAMETER). Read-only — use EE for SET/RESET.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List all visible parameters",
		Long:    `List all non-hidden initialization parameters with current values and whether they are modified.`,
		Example: `  dbxcli db parameter list --target prod-db
  dbxcli db parameter list --target prod-db --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "describe",
		Short:   "Describe a specific parameter",
		Long:    `Show full details for a parameter including description, default, range, and whether it requires restart.`,
		Example: `  dbxcli db parameter describe name=sga_target --target prod-db`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "modified",
		Short:   "List non-default parameters",
		Long:    `List only parameters that have been changed from their default values.`,
		Example: `  dbxcli db parameter modified --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter modified (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "hidden",
		Short:   "List hidden underscore parameters",
		Long:    `List hidden (underscore-prefixed) parameters. Use with caution — these are unsupported by Oracle.`,
		Example: `  dbxcli db parameter hidden --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter hidden (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newDBAdvisorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advisor",
		Short: "Oracle advisor operations",
		Long:  `Query Oracle built-in advisor recommendations (DBA_ADVISOR_RECOMMENDATIONS). Read-only.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "segment",
		Short:   "Show segment advisor recommendations",
		Long:    `Show segment advisor recommendations for table/index shrink, compression, and space reclamation.`,
		Example: `  dbxcli db advisor segment --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db advisor segment (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "sql-tuning",
		Short:   "List SQL tuning advisor tasks",
		Long:    `List SQL tuning advisor tasks and their recommendations (DBA_ADVISOR_TASKS).`,
		Example: `  dbxcli db advisor sql-tuning --target prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db advisor sql-tuning (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}
