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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active user sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db session list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe a session by SID",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tablespaces with usage metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db tablespace list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe tablespace datafiles",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List database users",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db user list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe user with roles and privileges",
		Args:  cobra.MinimumNArgs(1),
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
		Use:   "profiles",
		Short: "List database profiles",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List schemas with object counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db schema list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "objects",
		Short: "List objects in a schema",
		Args:  cobra.MinimumNArgs(1),
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
		Use:   "describe",
		Short: "Describe a specific object",
		Args:  cobra.MinimumNArgs(1),
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "exec",
		Short: "Execute a read-only SELECT statement",
		Args:  cobra.MinimumNArgs(1),
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
		Args:  cobra.MinimumNArgs(1),
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List redo log groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db redo list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "switch-history",
		Short: "Show log switch frequency",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List undo tablespace usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db undo list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "segments",
		Short: "Show undo segment details",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all visible parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe a specific parameter",
		Args:  cobra.MinimumNArgs(1),
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
		Use:   "modified",
		Short: "List non-default parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db parameter modified (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "hidden",
		Short: "List hidden underscore parameters",
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
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "segment",
		Short: "Show segment advisor recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db advisor segment (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "sql-tuning",
		Short: "List SQL tuning advisor tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("db advisor sql-tuning (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}
