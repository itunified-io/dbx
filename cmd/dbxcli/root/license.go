package root

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/itunified-io/dbx/pkg/license"
	"github.com/spf13/cobra"
)

// newLicenseCmd builds the "dbxcli license" command tree.
//
// Subcommands:
//   - status   show current state of ~/.dbx/license.jwt
//   - activate <path>  install a JWT file as the active license
//   - issue   DEV-MODE: self-sign a license with a local Ed25519 key
func newLicenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Manage dbx license",
		Long: `Manage the dbx license. Licenses are Ed25519-signed JWT tokens stored
at ~/.dbx/license.jwt (mode 0600). Without a license, only Community
(OSS) features are available — Enterprise tier + matching bundles
unlock gated commands such as ` + "`provision install`" + `.`,
	}
	cmd.AddCommand(newLicenseStatusCmd())
	cmd.AddCommand(newLicenseActivateCmd())
	cmd.AddCommand(newLicenseIssueCmd())
	return cmd
}

func newLicenseStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show license status",
		Long: `Show current license status: tier, bundles, expiry, and whether the
license is dev-issued (self-signed locally) or production-signed.`,
		Example: `  dbxcli license status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLicenseStatus(cmd.OutOrStdout())
		},
	}
}

func runLicenseStatus(w io.Writer) error {
	lic, err := license.Load()
	if errors.Is(err, license.ErrMissing) {
		fmt.Fprintln(w, "license: OSS (no license file)")
		fmt.Fprintln(w, "         Community-tier features only.")
		fmt.Fprintln(w, "         To self-sign a dev license: dbxcli license issue --tier enterprise --bundles provision")
		return nil
	}
	if err != nil {
		return fmt.Errorf("load license: %w", err)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "license\t: %s\n", license.Path())
	fmt.Fprintf(tw, "subject\t: %s\n", lic.Subject)
	fmt.Fprintf(tw, "tier\t: %s\n", lic.Tier)
	fmt.Fprintf(tw, "bundles\t: %s\n", strings.Join(lic.Bundles, ", "))
	if lic.IssuedAt > 0 {
		fmt.Fprintf(tw, "issued\t: %s\n", time.Unix(lic.IssuedAt, 0).UTC().Format(time.RFC3339))
	}
	if lic.ExpiresAt > 0 {
		fmt.Fprintf(tw, "expires\t: %s\n", time.Unix(lic.ExpiresAt, 0).UTC().Format(time.RFC3339))
	} else {
		fmt.Fprintf(tw, "expires\t: never\n")
	}
	if lic.Dev {
		fmt.Fprintf(tw, "source\t: dev-issued (self-signed) — NOT for production\n")
	} else {
		fmt.Fprintf(tw, "source\t: production-signed\n")
	}
	if vErr := lic.IsValid(); vErr != nil {
		fmt.Fprintf(tw, "state\t: INVALID — %s\n", vErr)
	} else {
		fmt.Fprintf(tw, "state\t: valid\n")
	}
	return tw.Flush()
}

func newLicenseActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <path>",
		Short: "Activate a license file",
		Long: `Read the JWT at <path>, verify it against trusted keys, and on
success copy it to ~/.dbx/license.jwt (mode 0600).`,
		Example: `  dbxcli license activate /path/to/license.jwt`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read %s: %w", args[0], err)
			}
			if err := license.Save(strings.TrimSpace(string(data))); err != nil {
				return err
			}
			// Re-load to validate signature against the trust list.
			if _, err := license.Load(); err != nil {
				return fmt.Errorf("activated file failed verification: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "license activated: %s\n", license.Path())
			return runLicenseStatus(cmd.OutOrStdout())
		},
	}
	return cmd
}

func newLicenseIssueCmd() *cobra.Command {
	var (
		tier     string
		bundles  []string
		subject  string
		expires  string
		outPath  string
	)
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "DEV-MODE: self-sign a license",
		Long: `DEV-MODE ONLY. Self-sign a license JWT using a locally-generated
Ed25519 key at ~/.dbx/.signing-key.ed25519, with the matching public
key auto-trusted under ~/.dbx/.trust/.

The resulting license carries a "dev" claim that prints a warning
banner whenever it is loaded — DO NOT use dev-issued licenses in
production deployments.`,
		Example: `  dbxcli license issue --tier enterprise \
    --bundles provision,dataguard,audit \
    --subject lab-dev --expires 365d \
    --out ~/.dbx/license.jwt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := license.Tier(strings.ToLower(tier))
			switch t {
			case license.TierCommunity, license.TierBusiness, license.TierEnterprise:
			default:
				return fmt.Errorf("invalid --tier %q (want community|business|enterprise)", tier)
			}
			expSecs, err := parseDuration(expires)
			if err != nil {
				return err
			}
			now := time.Now().Unix()
			tok, err := license.IssueDev(license.License{
				Subject:   subject,
				Tier:      t,
				Bundles:   bundles,
				IssuedAt:  now,
				ExpiresAt: now + expSecs,
			})
			if err != nil {
				return err
			}
			if outPath != "" {
				out := expandHome(outPath)
				if out == license.Path() {
					if err := license.Save(tok); err != nil {
						return err
					}
				} else {
					if err := os.WriteFile(out, []byte(tok), 0o600); err != nil {
						return fmt.Errorf("write %s: %w", out, err)
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "WARNING: dev-issued license written to %s\n", out)
				fmt.Fprintln(cmd.OutOrStdout(), "         This license is self-signed and NOT valid for production.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tok)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tier, "tier", "enterprise", "tier: community|business|enterprise")
	cmd.Flags().StringSliceVar(&bundles, "bundles", nil, "comma-separated bundle list (e.g. provision,dataguard)")
	cmd.Flags().StringVar(&subject, "subject", "lab-dev", "license subject (sub claim)")
	cmd.Flags().StringVar(&expires, "expires", "365d", "duration until expiry (e.g. 30d, 24h, 365d)")
	cmd.Flags().StringVar(&outPath, "out", "", "output path (default: stdout). Use ~/.dbx/license.jwt to install.")
	return cmd
}

// parseDuration accepts NNh / NNd / standard time.ParseDuration values
// and returns the equivalent number of seconds.
func parseDuration(s string) (int64, error) {
	if s == "" {
		return 0, errors.New("empty duration")
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return int64(n) * 86400, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return int64(d.Seconds()), nil
}

// expandHome rewrites a leading ~ to $HOME.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(p, "~", home, 1)
		}
	}
	return p
}
