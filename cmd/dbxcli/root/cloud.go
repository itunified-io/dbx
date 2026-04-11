package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCloudCmd creates the "cloud" subcommand group for cloud infrastructure provisioning.
func NewCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Cloud infrastructure provisioning and management",
		Long:  "Provision, manage, and monitor cloud infrastructure across AWS, Azure, OCI, and on-prem.",
	}

	cmd.AddCommand(
		newCloudProvisionCmd(),
		newCloudListCmd(),
		newCloudShowCmd(),
		newCloudTerminateCmd(),
		newCloudEstimateCmd(),
		newCloudRecommendCmd(),
	)

	return cmd
}

func newCloudProvisionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provision infrastructure from a YAML blueprint",
		Long: `Provision cloud infrastructure from a YAML blueprint file.
The blueprint defines instances, storage, networking, and load balancers.
Use --dry-run to preview the plan without executing.`,
		Example: `  dbxcli cloud provision --blueprint infra/oracle-prod.yaml --dry-run
  dbxcli cloud provision --blueprint infra/oracle-prod.yaml --confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			blueprint, _ := cmd.Flags().GetString("blueprint")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if blueprint == "" {
				return fmt.Errorf("--blueprint is required")
			}

			if dryRun {
				fmt.Println("DRY RUN -- showing what would be provisioned...")
				// Load blueprint, estimate cost, display plan
			}

			if !confirm && !dryRun {
				return fmt.Errorf("--confirm=true required for provisioning (or use --dry-run)")
			}

			fmt.Printf("Provisioning from blueprint: %s\n", blueprint)
			return nil
		},
	}
	cmd.Flags().String("blueprint", "", "Path to YAML blueprint file")
	cmd.Flags().Bool("dry-run", false, "Show plan without executing")
	cmd.Flags().Bool("confirm", false, "Confirm provisioning (required)")
	return cmd
}

func newCloudListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud instances",
		Long:  `List cloud instances across all registered providers. Filterable by provider, tags, and status.`,
		Example: `  dbxcli cloud list
  dbxcli cloud list --provider aws --status running`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing cloud instances...")
			return nil
		},
	}
	cmd.Flags().String("provider", "", "Filter by cloud provider (aws, azure, oci, onprem)")
	cmd.Flags().String("profile", "", "Vault credential profile")
	cmd.Flags().String("status", "", "Filter by status (running, stopped)")
	cmd.Flags().String("tag", "", "Filter by tag (key=value)")
	return cmd
}

func newCloudShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show instance details (IP, type, storage, status, cost)",
		Long:  `Show detailed information about a cloud instance including IPs, storage, cost attribution, and current state.`,
		Example: `  dbxcli cloud show --name ora-prod-01 --provider aws`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Showing instance details...")
			return nil
		},
	}
	cmd.Flags().String("name", "", "Instance name or ID")
	cmd.Flags().String("provider", "", "Cloud provider")
	cmd.Flags().String("profile", "", "Vault credential profile")
	return cmd
}

func newCloudTerminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate an instance (double-confirm -- destructive)",
		Long:  `Terminate a cloud instance. This is a destructive operation that cannot be undone. Requires --confirm flag.`,
		Example: `  dbxcli cloud terminate --name ora-prod-01 --confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			confirm, _ := cmd.Flags().GetBool("confirm")
			if !confirm {
				return fmt.Errorf("--confirm=true required for termination (destructive operation)")
			}
			fmt.Println("Terminating instance...")
			return nil
		},
	}
	cmd.Flags().String("name", "", "Instance name or ID")
	cmd.Flags().String("provider", "", "Cloud provider")
	cmd.Flags().String("profile", "", "Vault credential profile")
	cmd.Flags().Bool("confirm", false, "Double-confirm termination (destructive)")
	return cmd
}

func newCloudEstimateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate monthly cost for a blueprint",
		Long:  `Parse a YAML blueprint and compute the estimated monthly cost across compute, storage, network, and dbx licensing.`,
		Example: `  dbxcli cloud estimate --blueprint infra/oracle-prod.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			blueprint, _ := cmd.Flags().GetString("blueprint")
			if blueprint == "" {
				return fmt.Errorf("--blueprint is required")
			}
			fmt.Printf("Estimating cost for blueprint: %s\n", blueprint)
			return nil
		},
	}
	cmd.Flags().String("blueprint", "", "Path to YAML blueprint file")
	return cmd
}

func newCloudRecommendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommend",
		Short: "Recommend instance types for a workload profile",
		Long: `Recommend cloud instance types based on a workload profile (e.g., oracle_oltp_medium).
Shows recommended instance type, vCPUs, memory, storage, and IOPS per provider.`,
		Example: `  dbxcli cloud recommend --profile oracle_oltp_medium --provider aws
  dbxcli cloud recommend --profile pg_oltp_large --provider oci`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			if profile == "" {
				return fmt.Errorf("--profile is required (e.g., oracle_oltp_medium, pg_oltp_large)")
			}
			fmt.Printf("Recommending instance types for profile: %s\n", profile)
			return nil
		},
	}
	cmd.Flags().String("profile", "", "Workload profile name")
	cmd.Flags().String("provider", "", "Cloud provider (aws, azure, oci)")
	return cmd
}
