package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPgCmd creates the "pg" parent command for PostgreSQL database operations.
func NewPgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pg",
		Short: "PostgreSQL database operations",
		Long: `PostgreSQL database operations covering connections, queries, schema browsing,
CRUD, DBA, performance, security, replication, high availability, backup,
migration, observability, multi-tenancy, WAL, CNPG, disaster recovery,
RAG/pgvector, Vault credentials, and policy checks.

Requires a target with a PostgreSQL database endpoint configured.`,
	}

	cmd.PersistentFlags().String("target", "", "target name (from ~/.dbx/targets/)")

	cmd.AddCommand(newPgConnectCmd())
	cmd.AddCommand(newPgQueryCmd())
	cmd.AddCommand(newPgSchemaCmd())
	cmd.AddCommand(newPgCrudCmd())
	cmd.AddCommand(newPgDbaCmd())
	cmd.AddCommand(newPgDbaAdvCmd())
	cmd.AddCommand(newPgPerfCmd())
	cmd.AddCommand(newPgHealthCmd())
	cmd.AddCommand(newPgSecurityCmd())
	cmd.AddCommand(newPgAuditCmd())
	cmd.AddCommand(newPgComplyCmd())
	cmd.AddCommand(newPgRbacCmd())
	cmd.AddCommand(newPgReplCmd())
	cmd.AddCommand(newPgHaCmd())
	cmd.AddCommand(newPgBackupCmd())
	cmd.AddCommand(newPgMigrateCmd())
	cmd.AddCommand(newPgObserveCmd())
	cmd.AddCommand(newPgTenantCmd())
	cmd.AddCommand(newPgWalCmd())
	cmd.AddCommand(newPgCnpgCmd())
	cmd.AddCommand(newPgDrCmd())
	cmd.AddCommand(newPgRagCmd())
	cmd.AddCommand(newPgVaultCmd())
	cmd.AddCommand(newPgPolicyCmd())

	return cmd
}

// ---------------------------------------------------------------------------
// connect — PostgreSQL connection management (5 tools)
// ---------------------------------------------------------------------------

func newPgConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "PostgreSQL connection management",
		Long:  `Manage PostgreSQL connection profiles — add, remove, switch, and inspect connections.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "add",
		Short:   "Register a new connection profile",
		Long:    `Register a new PostgreSQL connection profile with host, port, database, and user.`,
		Example: `  dbxcli pg connect add name=prod host=db01.example.com port=5432 database=app user=admin --target prod-pg`,
		Args:    cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg connect add (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "remove",
		Short:   "Remove a connection profile",
		Long:    `Remove an existing PostgreSQL connection profile by name.`,
		Example: `  dbxcli pg connect remove name=prod --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg connect remove (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "switch",
		Short:   "Switch active connection",
		Long:    `Switch the active PostgreSQL connection to a different profile.`,
		Example: `  dbxcli pg connect switch name=staging --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg connect switch (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show current connection status",
		Long:    `Show the current active PostgreSQL connection status and details.`,
		Example: `  dbxcli pg connect status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg connect status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List all connection profiles",
		Long:    `List all registered PostgreSQL connection profiles.`,
		Example: `  dbxcli pg connect list --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg connect list (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// query — SQL query execution (3 tools)
// ---------------------------------------------------------------------------

func newPgQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "SQL query execution",
		Long:  `Execute SQL queries against a PostgreSQL database — SELECT, EXPLAIN, and prepared statements.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "exec",
		Short:   "Execute a read-only SELECT",
		Long:    `Execute a read-only SELECT query and return the results.`,
		Example: `  dbxcli pg query exec query="SELECT * FROM users LIMIT 10" --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg query exec (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "explain",
		Short:   "Generate execution plan",
		Long:    `Generate an EXPLAIN ANALYZE execution plan for a query.`,
		Example: `  dbxcli pg query explain query="SELECT * FROM orders WHERE status='pending'" --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg query explain (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "prepared",
		Short:   "Execute a prepared statement",
		Long:    `Execute a prepared statement by name with the given parameters.`,
		Example: `  dbxcli pg query prepared name=get_user params="[1]" --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg query prepared (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// schema — Schema browser (9 tools)
// ---------------------------------------------------------------------------

func newPgSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Schema browser",
		Long:  `Browse PostgreSQL schemas, tables, views, functions, types, sequences, indexes, and extensions.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "tables",
		Short:   "List tables",
		Long:    `List all tables in the specified schema.`,
		Example: `  dbxcli pg schema tables schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema tables (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "table-describe",
		Short:   "Describe a table",
		Long:    `Show column details, constraints, and metadata for a specific table.`,
		Example: `  dbxcli pg schema table-describe schema=public table=users --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema table-describe (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "indexes",
		Short:   "List indexes",
		Long:    `List all indexes for a table in the specified schema.`,
		Example: `  dbxcli pg schema indexes schema=public table=users --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema indexes (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "views",
		Short:   "List views",
		Long:    `List all views in the specified schema.`,
		Example: `  dbxcli pg schema views schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema views (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "functions",
		Short:   "List functions",
		Long:    `List all functions in the specified schema.`,
		Example: `  dbxcli pg schema functions schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema functions (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "types",
		Short:   "List custom types",
		Long:    `List all custom types in the specified schema.`,
		Example: `  dbxcli pg schema types schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema types (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "sequences",
		Short:   "List sequences",
		Long:    `List all sequences in the specified schema.`,
		Example: `  dbxcli pg schema sequences schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema sequences (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "extensions",
		Short:   "List installed extensions",
		Long:    `List all installed PostgreSQL extensions.`,
		Example: `  dbxcli pg schema extensions --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema extensions (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "schemas",
		Short:   "List schemas",
		Long:    `List all schemas in the database.`,
		Example: `  dbxcli pg schema schemas --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg schema schemas (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// crud — Data manipulation (4 tools)
// ---------------------------------------------------------------------------

func newPgCrudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crud",
		Short: "Data manipulation",
		Long:  `Insert, update, delete, and upsert rows in PostgreSQL tables.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "insert",
		Short:   "Insert a row",
		Long:    `Insert a new row into the specified table.`,
		Example: `  dbxcli pg crud insert schema=public table=users data='{"name":"alice"}' --target prod-pg`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg crud insert (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "update",
		Short:   "Update rows",
		Long:    `Update rows matching the WHERE clause in the specified table.`,
		Example: `  dbxcli pg crud update schema=public table=users set='{"active":true}' where="id=1" --target prod-pg`,
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg crud update (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "delete",
		Short:   "Delete rows",
		Long:    `Delete rows matching the WHERE clause from the specified table.`,
		Example: `  dbxcli pg crud delete schema=public table=users where="id=1" --target prod-pg`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg crud delete (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "upsert",
		Short:   "Upsert a row",
		Long:    `Insert a row or update it on conflict using the specified conflict key.`,
		Example: `  dbxcli pg crud upsert schema=public table=users data='{"id":1,"name":"alice"}' conflict=id --target prod-pg`,
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg crud upsert (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// dba — DBA operations (11 tools)
// ---------------------------------------------------------------------------

func newPgDbaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dba",
		Short: "DBA operations",
		Long:  `PostgreSQL DBA operations — server settings, connections, locks, long queries, and management.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "settings",
		Short:   "Show server settings",
		Long:    `Show current PostgreSQL server configuration settings.`,
		Example: `  dbxcli pg dba settings --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba settings (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "version",
		Short:   "Show PostgreSQL version",
		Long:    `Show the PostgreSQL server version string.`,
		Example: `  dbxcli pg dba version --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba version (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "reload",
		Short:   "Reload configuration",
		Long:    `Reload PostgreSQL server configuration without restart (pg_reload_conf).`,
		Example: `  dbxcli pg dba reload --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba reload (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "uptime",
		Short:   "Show server uptime",
		Long:    `Show how long the PostgreSQL server has been running.`,
		Example: `  dbxcli pg dba uptime --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba uptime (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "connections",
		Short:   "Show active connections",
		Long:    `Show all active connections from pg_stat_activity.`,
		Example: `  dbxcli pg dba connections --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba connections (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "locks",
		Short:   "Show lock information",
		Long:    `Show current lock information from pg_locks.`,
		Example: `  dbxcli pg dba locks --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba locks (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "blocker-tree",
		Short:   "Show blocking lock tree",
		Long:    `Show the blocking lock tree — which sessions block which.`,
		Example: `  dbxcli pg dba blocker-tree --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba blocker-tree (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "long-queries",
		Short:   "Show long-running queries",
		Long:    `Show queries that have been running longer than the configured threshold.`,
		Example: `  dbxcli pg dba long-queries --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba long-queries (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "cancel-query",
		Short:   "Cancel a running query",
		Long:    `Cancel a running query by backend PID (pg_cancel_backend).`,
		Example: `  dbxcli pg dba cancel-query pid=12345 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba cancel-query (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "terminate",
		Short:   "Terminate a backend",
		Long:    `Terminate a backend process by PID (pg_terminate_backend).`,
		Example: `  dbxcli pg dba terminate pid=12345 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba terminate (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "database-sizes",
		Short:   "Show database sizes",
		Long:    `Show the size of each database on the server.`,
		Example: `  dbxcli pg dba database-sizes --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba database-sizes (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// dba-adv — Advanced DBA operations (5 tools)
// ---------------------------------------------------------------------------

func newPgDbaAdvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dba-adv",
		Short: "Advanced DBA operations",
		Long:  `Advanced PostgreSQL DBA operations — vacuum status, bloat estimation, cache analysis, and unused indexes.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "vacuum-status",
		Short:   "Show vacuum status",
		Long:    `Show autovacuum and manual vacuum status for all tables.`,
		Example: `  dbxcli pg dba-adv vacuum-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba-adv vacuum-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "table-bloat",
		Short:   "Estimate table bloat",
		Long:    `Estimate table bloat across the database.`,
		Example: `  dbxcli pg dba-adv table-bloat --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba-adv table-bloat (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "index-bloat",
		Short:   "Estimate index bloat",
		Long:    `Estimate index bloat across the database.`,
		Example: `  dbxcli pg dba-adv index-bloat --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba-adv index-bloat (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "cache-hit",
		Short:   "Show buffer cache hit ratio",
		Long:    `Show the buffer cache hit ratio from pg_stat_database.`,
		Example: `  dbxcli pg dba-adv cache-hit --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba-adv cache-hit (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "unused-indexes",
		Short:   "List unused indexes",
		Long:    `List indexes with zero or very low scan counts.`,
		Example: `  dbxcli pg dba-adv unused-indexes --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dba-adv unused-indexes (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// perf — Performance analysis (2 tools)
// ---------------------------------------------------------------------------

func newPgPerfCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf",
		Short: "Performance analysis",
		Long:  `PostgreSQL performance analysis — pg_stat_statements and wait events.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "stat-statements",
		Short:   "Show pg_stat_statements top queries",
		Long:    `Show top queries from pg_stat_statements ordered by total time.`,
		Example: `  dbxcli pg perf stat-statements --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg perf stat-statements (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "wait-events",
		Short:   "Show wait event statistics",
		Long:    `Show wait event statistics from pg_stat_activity.`,
		Example: `  dbxcli pg perf wait-events --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg perf wait-events (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// health — Cluster health (1 tool)
// ---------------------------------------------------------------------------

func newPgHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Cluster health",
		Long:  `Run comprehensive PostgreSQL cluster health checks.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "check",
		Short:   "Run comprehensive health check",
		Long:    `Run a comprehensive health check covering connections, replication, locks, and storage.`,
		Example: `  dbxcli pg health check --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg health check (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// security — Security audit (4 tools)
// ---------------------------------------------------------------------------

func newPgSecurityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Security audit",
		Long:  `PostgreSQL security audit — roles, row-level security, SSL, and pg_hba.conf.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "roles",
		Short:   "List roles and privileges",
		Long:    `List all PostgreSQL roles with their privileges and membership.`,
		Example: `  dbxcli pg security roles --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg security roles (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "row-security",
		Short:   "Show row-level security policies",
		Long:    `Show all row-level security (RLS) policies defined on tables.`,
		Example: `  dbxcli pg security row-security --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg security row-security (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "ssl-status",
		Short:   "Show SSL/TLS connection status",
		Long:    `Show SSL/TLS connection status for all active connections.`,
		Example: `  dbxcli pg security ssl-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg security ssl-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pg-hba",
		Short:   "Parse pg_hba.conf rules",
		Long:    `Parse and display the active pg_hba.conf authentication rules.`,
		Example: `  dbxcli pg security pg-hba --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg security pg-hba (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// audit — Audit logging (3 tools)
// ---------------------------------------------------------------------------

func newPgAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit logging",
		Long:  `PostgreSQL audit logging — pgAudit status, log analysis, and DDL history.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "pgaudit-status",
		Short:   "Show pgAudit configuration",
		Long:    `Show the current pgAudit extension configuration and logging settings.`,
		Example: `  dbxcli pg audit pgaudit-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg audit pgaudit-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "log-analysis",
		Short:   "Analyze PostgreSQL logs",
		Long:    `Analyze PostgreSQL server logs for errors, warnings, and patterns.`,
		Example: `  dbxcli pg audit log-analysis --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg audit log-analysis (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "ddl-history",
		Short:   "Show DDL change history",
		Long:    `Show the history of DDL changes (CREATE, ALTER, DROP) from event triggers or logs.`,
		Example: `  dbxcli pg audit ddl-history --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg audit ddl-history (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// comply — Compliance checks (5 tools)
// ---------------------------------------------------------------------------

func newPgComplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comply",
		Short: "Compliance checks",
		Long:  `PostgreSQL compliance checks — CIS benchmarks, password policies, encryption, and privilege audits.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "cis-benchmark",
		Short:   "Run CIS benchmark checks",
		Long:    `Run CIS PostgreSQL benchmark checks against the database.`,
		Example: `  dbxcli pg comply cis-benchmark --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg comply cis-benchmark (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "password-policy",
		Short:   "Check password policies",
		Long:    `Check password policy enforcement on all database roles.`,
		Example: `  dbxcli pg comply password-policy --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg comply password-policy (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "encryption-status",
		Short:   "Check encryption at rest",
		Long:    `Check the status of data encryption at rest.`,
		Example: `  dbxcli pg comply encryption-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg comply encryption-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "privilege-audit",
		Short:   "Audit excessive privileges",
		Long:    `Audit roles for excessive privileges (superuser, CREATEROLE, etc.).`,
		Example: `  dbxcli pg comply privilege-audit --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg comply privilege-audit (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "report",
		Short:   "Generate compliance report",
		Long:    `Generate a comprehensive compliance report covering all checks.`,
		Example: `  dbxcli pg comply report --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg comply report (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// rbac — Role-based access control (4 tools)
// ---------------------------------------------------------------------------

func newPgRbacCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Role-based access control",
		Long:  `PostgreSQL role-based access control — list, grant, revoke, and inspect effective permissions.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list-roles",
		Short:   "List all roles with members",
		Long:    `List all roles with their member relationships.`,
		Example: `  dbxcli pg rbac list-roles --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rbac list-roles (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "grant",
		Short:   "Grant role or privilege",
		Long:    `Grant a role or privilege to a target role.`,
		Example: `  dbxcli pg rbac grant role=readonly target=app_user --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rbac grant (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "revoke",
		Short:   "Revoke role or privilege",
		Long:    `Revoke a role or privilege from a target role.`,
		Example: `  dbxcli pg rbac revoke role=readonly target=app_user --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rbac revoke (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "effective-perms",
		Short:   "Show effective permissions",
		Long:    `Show effective permissions for a role including inherited grants.`,
		Example: `  dbxcli pg rbac effective-perms role=app_user --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rbac effective-perms (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// repl — Replication management (4 tools)
// ---------------------------------------------------------------------------

func newPgReplCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Replication management",
		Long:  `PostgreSQL replication management — status, slots, lag, and logical subscriptions.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show replication status",
		Long:    `Show the current replication status from pg_stat_replication.`,
		Example: `  dbxcli pg repl status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg repl status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "slots",
		Short:   "List replication slots",
		Long:    `List all replication slots with their status and lag.`,
		Example: `  dbxcli pg repl slots --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg repl slots (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "lag",
		Short:   "Show replication lag",
		Long:    `Show replication lag for all standbys.`,
		Example: `  dbxcli pg repl lag --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg repl lag (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "subscriptions",
		Short:   "List logical subscriptions",
		Long:    `List all logical replication subscriptions.`,
		Example: `  dbxcli pg repl subscriptions --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg repl subscriptions (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// ha — High availability (10 tools)
// ---------------------------------------------------------------------------

func newPgHaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ha",
		Short: "High availability",
		Long:  `PostgreSQL high availability — Patroni cluster management, PgBouncer pools, and VIP failover.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "patroni-status",
		Short:   "Show Patroni cluster status",
		Long:    `Show the current Patroni cluster status including leader, replicas, and timeline.`,
		Example: `  dbxcli pg ha patroni-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha patroni-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "patroni-config",
		Short:   "Show Patroni configuration",
		Long:    `Show the current Patroni DCS configuration.`,
		Example: `  dbxcli pg ha patroni-config --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha patroni-config (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "patroni-switchover",
		Short:   "Initiate Patroni switchover",
		Long:    `Initiate a controlled Patroni switchover to the specified target node.`,
		Example: `  dbxcli pg ha patroni-switchover target_node=replica1 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha patroni-switchover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "patroni-failover",
		Short:   "Initiate Patroni failover",
		Long:    `Initiate a Patroni failover to the specified target node.`,
		Example: `  dbxcli pg ha patroni-failover target_node=replica1 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha patroni-failover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "patroni-restart",
		Short:   "Restart Patroni member",
		Long:    `Restart a specific Patroni cluster member.`,
		Example: `  dbxcli pg ha patroni-restart member=node1 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha patroni-restart (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbouncer-status",
		Short:   "Show PgBouncer pool status",
		Long:    `Show PgBouncer connection pool status and statistics.`,
		Example: `  dbxcli pg ha pgbouncer-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha pgbouncer-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbouncer-pools",
		Short:   "List PgBouncer pools",
		Long:    `List all PgBouncer connection pools with their configuration.`,
		Example: `  dbxcli pg ha pgbouncer-pools --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha pgbouncer-pools (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbouncer-pause",
		Short:   "Pause PgBouncer connections",
		Long:    `Pause all PgBouncer connection pools.`,
		Example: `  dbxcli pg ha pgbouncer-pause --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha pgbouncer-pause (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbouncer-resume",
		Short:   "Resume PgBouncer connections",
		Long:    `Resume all PgBouncer connection pools.`,
		Example: `  dbxcli pg ha pgbouncer-resume --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha pgbouncer-resume (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "vip-status",
		Short:   "Show VIP failover status",
		Long:    `Show the virtual IP failover status.`,
		Example: `  dbxcli pg ha vip-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg ha vip-status (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// backup — Backup operations (4 tools)
// ---------------------------------------------------------------------------

func newPgBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup operations",
		Long:  `PostgreSQL backup operations — pgBackRest and Barman management.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbackrest-status",
		Short:   "Show pgBackRest backup status",
		Long:    `Show the current pgBackRest backup status and history.`,
		Example: `  dbxcli pg backup pgbackrest-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg backup pgbackrest-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbackrest-backup",
		Short:   "Trigger pgBackRest backup",
		Long:    `Trigger a pgBackRest backup of the specified type and stanza.`,
		Example: `  dbxcli pg backup pgbackrest-backup type=full stanza=main --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg backup pgbackrest-backup (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pgbackrest-restore",
		Short:   "Restore from pgBackRest",
		Long:    `Restore from a pgBackRest backup with optional point-in-time target.`,
		Example: `  dbxcli pg backup pgbackrest-restore stanza=main target_time="2026-01-01 12:00:00" --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg backup pgbackrest-restore (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "barman-status",
		Short:   "Show Barman backup status",
		Long:    `Show the current Barman backup status and catalog.`,
		Example: `  dbxcli pg backup barman-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg backup barman-status (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// migrate — Migration operations (4 tools)
// ---------------------------------------------------------------------------

func newPgMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migration operations",
		Long:  `PostgreSQL migration operations — pg_upgrade, logical migration, and foreign data wrappers.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "pg-upgrade-check",
		Short:   "Pre-check pg_upgrade compatibility",
		Long:    `Run pg_upgrade compatibility checks without performing the upgrade.`,
		Example: `  dbxcli pg migrate pg-upgrade-check --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg migrate pg-upgrade-check (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pg-upgrade-run",
		Short:   "Run pg_upgrade",
		Long:    `Run pg_upgrade from the old version to the new version.`,
		Example: `  dbxcli pg migrate pg-upgrade-run old_version=15 new_version=16 --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg migrate pg-upgrade-run (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "logical-migrate",
		Short:   "Logical migration via pub/sub",
		Long:    `Perform a logical migration using publication/subscription between source and target.`,
		Example: `  dbxcli pg migrate logical-migrate source=old-db target=new-db --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg migrate logical-migrate (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "fdw-setup",
		Short:   "Setup foreign data wrapper",
		Long:    `Setup a foreign data wrapper to access a remote server and schema.`,
		Example: `  dbxcli pg migrate fdw-setup server=remote-db schema=public --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg migrate fdw-setup (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// observe — Observability (4 tools)
// ---------------------------------------------------------------------------

func newPgObserveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observe",
		Short: "Observability",
		Long:  `PostgreSQL observability — Prometheus metrics, slow queries, activity, and background writer stats.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "metrics",
		Short:   "Export Prometheus-compatible metrics",
		Long:    `Export PostgreSQL metrics in Prometheus-compatible format.`,
		Example: `  dbxcli pg observe metrics --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg observe metrics (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "slow-log",
		Short:   "Analyze slow query log",
		Long:    `Analyze the PostgreSQL slow query log for patterns.`,
		Example: `  dbxcli pg observe slow-log --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg observe slow-log (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "stat-activity",
		Short:   "Show pg_stat_activity snapshot",
		Long:    `Show a snapshot of pg_stat_activity with all backend details.`,
		Example: `  dbxcli pg observe stat-activity --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg observe stat-activity (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "stat-bgwriter",
		Short:   "Show background writer stats",
		Long:    `Show background writer statistics from pg_stat_bgwriter.`,
		Example: `  dbxcli pg observe stat-bgwriter --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg observe stat-bgwriter (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// tenant — Multi-tenant management (5 tools)
// ---------------------------------------------------------------------------

func newPgTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Multi-tenant management",
		Long:  `PostgreSQL multi-tenant management — provision, deprovision, list, migrate, and check tenant databases.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "provision",
		Short:   "Provision new tenant database",
		Long:    `Provision a new tenant database with schema and role setup.`,
		Example: `  dbxcli pg tenant provision tenant_id=acme-corp --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg tenant provision (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "deprovision",
		Short:   "Remove tenant database",
		Long:    `Remove a tenant database and all associated resources.`,
		Example: `  dbxcli pg tenant deprovision tenant_id=acme-corp --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg tenant deprovision (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List tenant databases",
		Long:    `List all tenant databases with their status and metadata.`,
		Example: `  dbxcli pg tenant list --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg tenant list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "migrate",
		Short:   "Run tenant schema migration",
		Long:    `Run schema migrations on a specific tenant database.`,
		Example: `  dbxcli pg tenant migrate tenant_id=acme-corp --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg tenant migrate (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show tenant health status",
		Long:    `Show health status for a specific tenant database.`,
		Example: `  dbxcli pg tenant status tenant_id=acme-corp --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg tenant status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// wal — WAL management (5 tools)
// ---------------------------------------------------------------------------

func newPgWalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wal",
		Short: "WAL management",
		Long:  `PostgreSQL WAL (Write-Ahead Log) management — status, archiving, size, retention, and replay lag.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show WAL generation stats",
		Long:    `Show WAL generation statistics and current LSN.`,
		Example: `  dbxcli pg wal status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg wal status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "archive-status",
		Short:   "Check WAL archiving status",
		Long:    `Check the WAL archiving status and any failures.`,
		Example: `  dbxcli pg wal archive-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg wal archive-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "size",
		Short:   "Show WAL directory size",
		Long:    `Show the current WAL directory size on disk.`,
		Example: `  dbxcli pg wal size --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg wal size (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "retention",
		Short:   "Show WAL retention policy",
		Long:    `Show the configured WAL retention policy and current usage.`,
		Example: `  dbxcli pg wal retention --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg wal retention (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "replay-lag",
		Short:   "Show WAL replay lag on standbys",
		Long:    `Show WAL replay lag on all standby servers.`,
		Example: `  dbxcli pg wal replay-lag --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg wal replay-lag (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// cnpg — CloudNativePG (6 tools)
// ---------------------------------------------------------------------------

func newPgCnpgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cnpg",
		Short: "CloudNativePG",
		Long:  `CloudNativePG cluster management — list, status, promote, restart, backup, and switchover.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "clusters",
		Short:   "List CNPG clusters",
		Long:    `List all CloudNativePG clusters in the specified namespace.`,
		Example: `  dbxcli pg cnpg clusters namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg clusters (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "cluster-status",
		Short:   "Show CNPG cluster status",
		Long:    `Show detailed status for a specific CloudNativePG cluster.`,
		Example: `  dbxcli pg cnpg cluster-status name=my-cluster namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg cluster-status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "promote",
		Short:   "Promote CNPG replica",
		Long:    `Promote a CloudNativePG replica to primary.`,
		Example: `  dbxcli pg cnpg promote name=my-cluster namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg promote (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "restart",
		Short:   "Rolling restart CNPG cluster",
		Long:    `Perform a rolling restart of a CloudNativePG cluster.`,
		Example: `  dbxcli pg cnpg restart name=my-cluster namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg restart (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "backup-status",
		Short:   "Show CNPG backup status",
		Long:    `Show backup status for a CloudNativePG cluster.`,
		Example: `  dbxcli pg cnpg backup-status name=my-cluster namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg backup-status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "switchover",
		Short:   "Trigger CNPG switchover",
		Long:    `Trigger a switchover on a CloudNativePG cluster.`,
		Example: `  dbxcli pg cnpg switchover name=my-cluster namespace=default --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg cnpg switchover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// dr — Disaster recovery (18 tools)
// ---------------------------------------------------------------------------

func newPgDrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dr",
		Short: "Disaster recovery",
		Long: `PostgreSQL disaster recovery — configuration, status, switchover, failover,
WAL shipping, MinIO archive, PITR, runbooks, and compliance reporting.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "config-list",
		Short:   "List DR configurations",
		Long:    `List all disaster recovery configurations.`,
		Example: `  dbxcli pg dr config-list --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr config-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "config-get",
		Short:   "Get DR configuration",
		Long:    `Get a specific disaster recovery configuration by name.`,
		Example: `  dbxcli pg dr config-get name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr config-get (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "config-set",
		Short:   "Set DR configuration",
		Long:    `Create or update a disaster recovery configuration.`,
		Example: `  dbxcli pg dr config-set name=prod-dr primary=dc1-pg standby=dc2-pg --target prod-pg`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr config-set (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "config-delete",
		Short:   "Delete DR configuration",
		Long:    `Delete a disaster recovery configuration by name.`,
		Example: `  dbxcli pg dr config-delete name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr config-delete (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show DR status",
		Long:    `Show the current disaster recovery status for a configuration.`,
		Example: `  dbxcli pg dr status name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "validate",
		Short:   "Validate DR setup",
		Long:    `Validate that the DR setup is correctly configured and operational.`,
		Example: `  dbxcli pg dr validate name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr validate (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "switchover",
		Short:   "Execute DR switchover",
		Long:    `Execute a controlled DR switchover from primary to standby.`,
		Example: `  dbxcli pg dr switchover name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr switchover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "failover",
		Short:   "Execute DR failover",
		Long:    `Execute an emergency DR failover to the standby.`,
		Example: `  dbxcli pg dr failover name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr failover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "test-failover",
		Short:   "Dry-run failover test",
		Long:    `Run a dry-run failover test without actually switching over.`,
		Example: `  dbxcli pg dr test-failover name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr test-failover (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "sync-status",
		Short:   "Show replication sync status",
		Long:    `Show the replication synchronization status for a DR configuration.`,
		Example: `  dbxcli pg dr sync-status name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr sync-status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "wal-shipping",
		Short:   "Configure WAL shipping",
		Long:    `Configure WAL shipping for a DR configuration.`,
		Example: `  dbxcli pg dr wal-shipping name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr wal-shipping (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "minio-status",
		Short:   "Show MinIO WAL archive status",
		Long:    `Show the MinIO WAL archive status and bucket health.`,
		Example: `  dbxcli pg dr minio-status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr minio-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "minio-verify",
		Short:   "Verify MinIO archive integrity",
		Long:    `Verify the integrity of the MinIO WAL archive.`,
		Example: `  dbxcli pg dr minio-verify --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr minio-verify (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pitr-restore",
		Short:   "Point-in-time recovery",
		Long:    `Perform a point-in-time recovery to a specific timestamp.`,
		Example: `  dbxcli pg dr pitr-restore name=prod-dr target_time="2026-01-01 12:00:00" --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr pitr-restore (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "pitr-timeline",
		Short:   "Show PITR timeline",
		Long:    `Show the PITR timeline and available recovery points.`,
		Example: `  dbxcli pg dr pitr-timeline name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr pitr-timeline (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "runbook",
		Short:   "Generate DR runbook",
		Long:    `Generate a disaster recovery runbook for a configuration.`,
		Example: `  dbxcli pg dr runbook name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr runbook (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "report",
		Short:   "Generate DR compliance report",
		Long:    `Generate a disaster recovery compliance report.`,
		Example: `  dbxcli pg dr report --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr report (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "monitor",
		Short:   "Start continuous DR monitoring",
		Long:    `Start continuous disaster recovery monitoring for a configuration.`,
		Example: `  dbxcli pg dr monitor name=prod-dr --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg dr monitor (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// rag — RAG/pgvector (7 tools)
// ---------------------------------------------------------------------------

func newPgRagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rag",
		Short: "RAG/pgvector",
		Long:  `PostgreSQL RAG and pgvector operations — vector collections, similarity search, ingestion, and configuration.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "collections",
		Short:   "List vector collections",
		Long:    `List all vector collections (tables with vector columns).`,
		Example: `  dbxcli pg rag collections --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag collections (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "collection-create",
		Short:   "Create vector collection",
		Long:    `Create a new vector collection with the specified dimensions.`,
		Example: `  dbxcli pg rag collection-create name=embeddings dimensions=1536 --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag collection-create (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "collection-drop",
		Short:   "Drop vector collection",
		Long:    `Drop a vector collection and all its data.`,
		Example: `  dbxcli pg rag collection-drop name=embeddings --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag collection-drop (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "search",
		Short:   "Similarity search",
		Long:    `Perform a similarity search against a vector collection.`,
		Example: `  dbxcli pg rag search collection=embeddings query="machine learning" limit=10 --target prod-pg`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag search (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "ingest",
		Short:   "Ingest documents",
		Long:    `Ingest documents into a vector collection from the specified source.`,
		Example: `  dbxcli pg rag ingest collection=embeddings source=/path/to/docs --target prod-pg`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag ingest (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "index-status",
		Short:   "Show vector index stats",
		Long:    `Show vector index statistics for a collection.`,
		Example: `  dbxcli pg rag index-status collection=embeddings --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag index-status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "config",
		Short:   "Show pgvector configuration",
		Long:    `Show the current pgvector extension configuration.`,
		Example: `  dbxcli pg rag config --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg rag config (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// vault — Vault credential management (3 tools)
// ---------------------------------------------------------------------------

func newPgVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault credential management",
		Long:  `HashiCorp Vault integration for PostgreSQL credential rotation, leases, and revocation.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "rotate",
		Short:   "Rotate database credentials",
		Long:    `Rotate database credentials for the specified Vault role.`,
		Example: `  dbxcli pg vault rotate role=app-readonly --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg vault rotate (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show Vault lease status",
		Long:    `Show the current Vault lease status for database credentials.`,
		Example: `  dbxcli pg vault status --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg vault status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "revoke",
		Short:   "Revoke database credentials",
		Long:    `Revoke a Vault database credential lease by ID.`,
		Example: `  dbxcli pg vault revoke lease_id=database/creds/app-readonly/abc123 --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg vault revoke (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

// ---------------------------------------------------------------------------
// policy — Policy engine (2 tools)
// ---------------------------------------------------------------------------

func newPgPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Policy engine",
		Long:  `PostgreSQL policy engine — run policy checks and list available policies.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "check",
		Short:   "Run policy check against database",
		Long:    `Run a specific policy check against the target database.`,
		Example: `  dbxcli pg policy check policy_name=no-superuser-apps --target prod-pg`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg policy check (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List available policies",
		Long:    `List all available policies that can be checked against databases.`,
		Example: `  dbxcli pg policy list --target prod-pg`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("pg policy list (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}
