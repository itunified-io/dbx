package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPolicyCmd creates the "policy" subcommand group for compliance scanning.
func NewPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Policy compliance scanning (CIS, STIG, custom)",
		Long: `Evaluate security and compliance policies against targets.
Supports CIS benchmarks, DISA STIG, and custom YAML-driven policies
for Linux hosts, Oracle Database, and PostgreSQL.`,
	}

	cmd.AddCommand(newPolicyScanCmd())
	cmd.AddCommand(newPolicyReportCmd())
	cmd.AddCommand(newPolicyDriftCmd())
	cmd.AddCommand(newPolicyStatusCmd())
	cmd.AddCommand(newPolicyReloadCmd())
	cmd.AddCommand(newPolicyFleetScanCmd())
	cmd.AddCommand(newPolicyFleetReportCmd())
	cmd.AddCommand(newPolicyRemediateCmd())

	return cmd
}

func newPolicyScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run policy scan against a target",
		Long:  `Evaluate all matching policies against a target entity and output the scan result.`,
		Example: `  dbxcli policy scan entity_name=db-prod entity_type=oracle_database framework=cis
  dbxcli policy scan entity_name=web01 entity_type=host framework=stig
  dbxcli policy scan entity_name=pg-prod entity_type=pg_database level=1`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			entityName := params["entity_name"]
			entityType := params["entity_type"]
			if entityName == "" || entityType == "" {
				return fmt.Errorf("entity_name and entity_type are required")
			}
			framework := params["framework"]
			level := params["level"]
			fmt.Printf("policy scan: entity=%s type=%s framework=%s level=%s\n", entityName, entityType, framework, level)
			return nil
		},
	}
	return cmd
}

func newPolicyReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate compliance report from last scan",
		Long:  `Generate a compliance report in JSON, HTML, or CSV format from the most recent scan result.`,
		Example: `  dbxcli policy report entity_name=db-prod entity_type=oracle_database format=html output=/tmp/report.html
  dbxcli policy report entity_name=web01 entity_type=host format=csv`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			entityName := params["entity_name"]
			entityType := params["entity_type"]
			format := params["format"]
			if format == "" {
				format = "json"
			}
			output := params["output"]
			fmt.Printf("policy report: entity=%s type=%s format=%s output=%s\n", entityName, entityType, format, output)
			return nil
		},
	}
	return cmd
}

func newPolicyDriftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Compare current state against baseline scan",
		Long:  `Compare the most recent scan result against a baseline to detect compliance drift.`,
		Example: `  dbxcli policy drift entity_name=db-prod entity_type=oracle_database baseline=2026-04-01`,
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			entityName := params["entity_name"]
			entityType := params["entity_type"]
			baseline := params["baseline"]
			fmt.Printf("policy drift: entity=%s type=%s baseline=%s\n", entityName, entityType, baseline)
			return nil
		},
	}
	return cmd
}

func newPolicyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show loaded policies and versions",
		Long:    `List all loaded policy files with their framework, scope, version, and SHA-256 hash.`,
		Example: `  dbxcli policy status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("policy status: listing loaded policies...")
			return nil
		},
	}
}

func newPolicyReloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "reload",
		Short:   "Reload policy files from disk",
		Long:    `Re-read all policy YAML files from built-in and custom directories.`,
		Example: `  dbxcli policy reload`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("policy reload: reloading policies from disk...")
			return nil
		},
	}
}

func newPolicyFleetScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet-scan",
		Short: "Scan all registered targets",
		Long:  `Run a policy scan against all registered targets in the fleet.`,
		Example: `  dbxcli policy fleet-scan framework=cis
  dbxcli policy fleet-scan`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			framework := params["framework"]
			fmt.Printf("policy fleet-scan: framework=%s\n", framework)
			return nil
		},
	}
	return cmd
}

func newPolicyFleetReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet-report",
		Short: "Generate fleet-wide compliance report",
		Long:  `Aggregate scan results from all targets into a single compliance report.`,
		Example: `  dbxcli policy fleet-report format=csv output=/tmp/fleet.csv
  dbxcli policy fleet-report format=json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			format := params["format"]
			if format == "" {
				format = "csv"
			}
			output := params["output"]
			fmt.Printf("policy fleet-report: format=%s output=%s\n", format, output)
			return nil
		},
	}
	return cmd
}

func newPolicyRemediateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remediate",
		Short: "Apply remediation for a failing rule (confirm-gated)",
		Long: `Execute the remediation command for a specific rule that failed during a scan.
This is a destructive operation that requires explicit confirmation.`,
		Example: `  dbxcli policy remediate entity_name=web01 entity_type=host rule=5.2.2 confirm=true`,
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			entityName := params["entity_name"]
			entityType := params["entity_type"]
			ruleID := params["rule"]
			confirm := params["confirm"]
			if confirm != "true" {
				return fmt.Errorf("remediation requires confirm=true (destructive operation)")
			}
			fmt.Printf("policy remediate: entity=%s type=%s rule=%s\n", entityName, entityType, ruleID)
			return nil
		},
	}
	return cmd
}
