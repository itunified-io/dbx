// Command docgen generates LLM-friendly documentation from the Cobra command tree.
//
// It produces:
//   - docs/cli/*.md   — one Markdown file per command (Cobra doc format)
//   - llms.txt        — single flat file with all commands for LLM ingestion
//
// Usage:
//
//	go run ./cmd/docgen                      # default output dirs
//	go run ./cmd/docgen -out docs/cli -llms llms.txt
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

func main() {
	var outDir, llmsFile string

	genCmd := &cobra.Command{
		Use:   "docgen",
		Short: "Generate CLI documentation for LLMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build the full command tree (same as dbxcli main).
			rootCmd := root.New("dev")
			rootCmd.DisableAutoGenTag = true

			// 1. Generate Markdown tree.
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}
			if err := doc.GenMarkdownTree(rootCmd, outDir); err != nil {
				return fmt.Errorf("generate markdown: %w", err)
			}
			count := countFiles(outDir, ".md")
			fmt.Printf("Generated %d Markdown files in %s\n", count, outDir)

			// 2. Generate llms.txt — single flat file.
			if err := generateLLMsTxt(rootCmd, llmsFile); err != nil {
				return fmt.Errorf("generate llms.txt: %w", err)
			}
			fmt.Printf("Generated %s\n", llmsFile)

			return nil
		},
	}

	genCmd.Flags().StringVar(&outDir, "out", "docs/cli", "output directory for Markdown files")
	genCmd.Flags().StringVar(&llmsFile, "llms", "llms.txt", "output path for llms.txt")

	if err := genCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// generateLLMsTxt writes a single flat file with all commands, flags, and examples.
func generateLLMsTxt(rootCmd *cobra.Command, path string) error {
	var buf bytes.Buffer

	buf.WriteString("# dbx CLI Reference\n")
	buf.WriteString("# Auto-generated — do not edit manually.\n")
	buf.WriteString("# Source: go run ./cmd/docgen\n\n")

	writeCommand(&buf, rootCmd, "")

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func writeCommand(w io.Writer, cmd *cobra.Command, prefix string) {
	fullName := prefix + cmd.Name()
	if prefix != "" {
		fullName = prefix + " " + cmd.Name()
	}

	// Skip help commands.
	if cmd.Name() == "help" || cmd.Name() == "completion" {
		return
	}

	fmt.Fprintf(w, "## %s\n", fullName)
	fmt.Fprintf(w, "  %s\n", cmd.Short)

	if cmd.Long != "" {
		// Indent Long description.
		for _, line := range strings.Split(strings.TrimSpace(cmd.Long), "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
	}

	if hasNonHelpLocalFlags(cmd) {
		fmt.Fprintln(w, "  Flags:")
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}
			fmt.Fprintf(w, "    --%s", f.Name)
			if f.Shorthand != "" {
				fmt.Fprintf(w, ", -%s", f.Shorthand)
			}
			fmt.Fprintf(w, "  %s", f.Usage)
			if f.DefValue != "" && f.DefValue != "false" {
				fmt.Fprintf(w, " (default: %s)", f.DefValue)
			}
			fmt.Fprintln(w)
		})
	}

	if hasNonHelpInheritedFlags(cmd) && len(cmd.Commands()) > 0 {
		fmt.Fprintln(w, "  Inherited Flags:")
		cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}
			fmt.Fprintf(w, "    --%s  %s\n", f.Name, f.Usage)
		})
	}

	if cmd.Example != "" {
		fmt.Fprintln(w, "  Examples:")
		for _, line := range strings.Split(strings.TrimSpace(cmd.Example), "\n") {
			fmt.Fprintf(w, "    %s\n", line)
		}
	}

	if cmd.Aliases != nil && len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "  Aliases: %s\n", strings.Join(cmd.Aliases, ", "))
	}

	fmt.Fprintln(w)

	for _, sub := range cmd.Commands() {
		writeCommand(w, sub, fullName)
	}
}

func hasNonHelpLocalFlags(cmd *cobra.Command) bool {
	has := false
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			has = true
		}
	})
	return has
}

func hasNonHelpInheritedFlags(cmd *cobra.Command) bool {
	has := false
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			has = true
		}
	})
	return has
}

func countFiles(dir, ext string) int {
	entries, _ := os.ReadDir(dir)
	count := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ext {
			count++
		}
	}
	return count
}
