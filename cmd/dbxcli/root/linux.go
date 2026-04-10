package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewLinuxCmd creates the "linux" parent command for Oracle Linux host operations.
func NewLinuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linux",
		Short: "Oracle Linux host management over SSH",
		Long: `Oracle Linux system management operations executed over SSH.
Covers package management, kernel parameters, storage/LVM, network, and security.

Requires a target with SSH endpoint configured (oracle_host entity_type).`,
	}

	cmd.PersistentFlags().String("target", "", "target name (from ~/.dbx/targets/)")

	cmd.AddCommand(newLinuxPackageCmd())
	cmd.AddCommand(newLinuxKernelCmd())
	cmd.AddCommand(newLinuxStorageCmd())
	cmd.AddCommand(newLinuxNetworkCmd())
	cmd.AddCommand(newLinuxSecurityCmd())

	return cmd
}

func newLinuxPackageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "RPM/DNF package management",
		Long:  `RPM package queries and DNF install/update operations via SSH (rpm, dnf commands).`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List installed RPM packages",
		Long:    `List all installed RPM packages with version and architecture.`,
		Example: `  dbxcli linux package list --target web01`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux package list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "info",
		Short:   "Show package details",
		Long:    `Show detailed information for a specific RPM package.`,
		Example: `  dbxcli linux package info name=oracle-database-ee-19c --target db01`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux package info (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "install",
		Short:   "Install a package via DNF (confirm-gated)",
		Long:    `Install a package using dnf. Requires confirmation before execution.`,
		Example: `  dbxcli linux package install name=oracle-database-preinstall-19c --target db01`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux package install (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "update",
		Short:   "Update a package via DNF (confirm-gated)",
		Long:    `Update a package using dnf. Requires confirmation before execution.`,
		Example: `  dbxcli linux package update name=oracle-database-preinstall-19c --target db01`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux package update (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

func newLinuxKernelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kernel",
		Short: "Kernel parameter management",
		Long:  `Kernel parameter management via sysctl and /proc filesystem over SSH.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "param-list",
		Short: "List all sysctl parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux kernel param-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "param-set",
		Short: "Set a sysctl parameter (confirm-gated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux kernel param-set (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "hugepages",
		Short: "Set hugepages count (confirm-gated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux kernel hugepages (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show OS/kernel info (uname)",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux kernel info (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newLinuxStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Storage and LVM management",
		Long:  `Linux storage and LVM management via pvs/vgs/lvs, lvcreate, df, lsblk over SSH.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "pv-list",
		Short: "List physical volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux storage pv-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "vg-list",
		Short: "List volume groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux storage vg-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "lv-list",
		Short: "List logical volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux storage lv-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "lv-create",
		Short: "Create a logical volume (confirm-gated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux storage lv-create (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "disk-usage",
		Short: "Show disk usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux storage disk-usage (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newLinuxNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Network diagnostics",
		Long:  `Network diagnostics via ip, nmcli, chronyc, and ss over SSH.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "nic-list",
		Short: "List NICs with addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux network nic-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "bond-status",
		Short: "Show network bond status",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux network bond-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "dns-check",
		Short: "Check DNS resolver configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux network dns-check (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "ntp-status",
		Short: "Check NTP synchronization",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux network ntp-status (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newLinuxSecurityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Security status checks",
		Long:  `Security status checks: SELinux, firewall rules, and running services over SSH.`,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "selinux-status",
		Short: "Check SELinux status",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux security selinux-status (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "firewall-list",
		Short: "List firewall rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux security firewall-list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "service-status",
		Short: "List running services",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("linux security service-status (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}
