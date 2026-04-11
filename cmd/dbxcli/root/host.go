package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewHostCmd creates the "host" parent command for multi-distro host operations.
func NewHostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Host OS monitoring and management (multi-distro)",
		Long: `Monitor and manage Linux hosts across Fedora, Ubuntu, RHEL, SLES, and Oracle Linux.
Distro detection is automatic via /etc/os-release. All operations go through SSH
using the distro abstraction layer — callers never need to know apt vs dnf vs zypper.`,
	}

	cmd.PersistentFlags().String("target", "", "target name (from ~/.dbx/targets/)")

	cmd.AddCommand(newHostInfoCmd())
	cmd.AddCommand(newHostCPUCmd())
	cmd.AddCommand(newHostMemoryCmd())
	cmd.AddCommand(newHostDiskCmd())
	cmd.AddCommand(newHostNetworkCmd())
	cmd.AddCommand(newHostProcessCmd())
	cmd.AddCommand(newHostFilesystemCmd())
	cmd.AddCommand(newHostKernelCmd())
	cmd.AddCommand(newHostServiceCmd())
	cmd.AddCommand(newHostPackageCmd())
	cmd.AddCommand(newHostUserCmd())
	cmd.AddCommand(newHostUptimeCmd())
	cmd.AddCommand(newHostLoadCmd())
	cmd.AddCommand(newHostNTPCmd())
	cmd.AddCommand(newHostDNSCmd())

	return cmd
}

func newHostInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show OS info, distro, kernel, uptime",
		Long:  `Auto-detect distro and show OS release, kernel version, architecture, and uptime.`,
		Example: `  dbxcli host info --target web01
  dbxcli host info entity_name=prod-db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host info (target=%s)\n", target)
			return nil
		},
	}
}

func newHostCPUCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cpu",
		Short: "Show CPU usage, cores, load average",
		Long:  `Collect CPU metrics from /proc/stat (delta-based) and lscpu.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host cpu (target=%s)\n", target)
			return nil
		},
	}
}

func newHostMemoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "memory",
		Short: "Show memory and swap usage",
		Long:  `Parse /proc/meminfo for RAM, swap, buffers, cached, and hugepages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host memory (target=%s)\n", target)
			return nil
		},
	}
}

func newHostDiskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Disk I/O and space usage",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "io",
		Short: "Show disk I/O metrics (IOPS, throughput)",
		Long:  `Delta-based disk I/O from /proc/diskstats — IOPS, throughput (MB/s) per device.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host disk io (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "space",
		Short: "Show disk space usage",
		Long:  `Disk space from df -k, excluding pseudo-filesystems.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host disk space (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newHostNetworkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "network",
		Short: "Show network interfaces, routes, listening ports",
		Long:  `NIC list, IP addresses, routing table, and listening ports via ip addr and ss.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host network (target=%s)\n", target)
			return nil
		},
	}
}

func newHostProcessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process monitoring",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "top",
		Short: "Show top N processes by CPU or memory",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host process top (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host process list (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newHostFilesystemCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "filesystem",
		Short: "Show mounts, LVM layout, inodes",
		Long:  `Mount points from findmnt, LVM topology from pvs/lvs, inode counts from df -i.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host filesystem (target=%s)\n", target)
			return nil
		},
	}
}

func newHostKernelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kernel",
		Short: "Show kernel parameters, modules, hugepages",
		Long:  `Sysctl params, loaded modules via lsmod, hugepage configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host kernel (target=%s)\n", target)
			return nil
		},
	}
}

func newHostServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Systemd service management",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all systemd services",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host service list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show detailed service status",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host service status (target=%s, params=%v)\n", target, params)
			return nil
		},
	})
	return cmd
}

func newHostPackageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Distro-agnostic package management",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List installed packages (rpm/dpkg)",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host package list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "updates",
		Short: "Show available updates (security-only with --security)",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host package updates (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newHostUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User account audit",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List OS users with login shell status",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host user list (target=%s)\n", target)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "sessions",
		Short: "Show active login sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host user sessions (target=%s)\n", target)
			return nil
		},
	})
	return cmd
}

func newHostUptimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uptime",
		Short: "Show system uptime",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host uptime (target=%s)\n", target)
			return nil
		},
	}
}

func newHostLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Show load average history",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host load (target=%s)\n", target)
			return nil
		},
	}
}

func newHostNTPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ntp",
		Short: "Show NTP synchronization status",
		Long:  `Parse chronyc tracking for NTP server, stratum, offset, and leap status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host ntp (target=%s)\n", target)
			return nil
		},
	}
}

func newHostDNSCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dns",
		Short: "Show DNS resolver configuration",
		Long:  `Parse /etc/resolv.conf for nameservers and search domains.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			fmt.Printf("host dns (target=%s)\n", target)
			return nil
		},
	}
}
