package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRagCmd creates the "rag" subcommand group for AI retrieval-augmented generation.
func NewRagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rag",
		Short: "RAG — retrieval-augmented generation for AI assistance",
		Long: `Manage the RAG knowledge base for AI-powered database administration.
Index documentation from Oracle, PostgreSQL, Linux, MOS notes, and custom sources.
Search and build token-limited context for LLM prompts.`,
	}

	cmd.AddCommand(newRagSearchCmd())
	cmd.AddCommand(newRagContextCmd())
	cmd.AddCommand(newRagIndexStatusCmd())
	cmd.AddCommand(newRagIndexRefreshCmd())
	cmd.AddCommand(newRagSourcesCmd())

	return cmd
}

func newRagSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Semantic search over indexed documents",
		Long:  `Search the RAG knowledge base using semantic similarity and optional keyword matching.`,
		Example: `  dbxcli rag search query="how to generate AWR reports" top_k=5
  dbxcli rag search query="vacuum tuning" source=pg-docs mode=hybrid`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			query := params["query"]
			if query == "" {
				return fmt.Errorf("query is required")
			}
			source := params["source"]
			mode := params["mode"]
			if mode == "" {
				mode = "semantic"
			}
			topK := params["top_k"]
			if topK == "" {
				topK = "5"
			}
			fmt.Printf("rag search: query=%q source=%s mode=%s top_k=%s\n", query, source, mode, topK)
			return nil
		},
	}
}

func newRagContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "Build token-limited LLM context from search results",
		Long:  `Search the knowledge base and assemble a token-limited context string for LLM prompts.`,
		Example: `  dbxcli rag context query="Oracle performance tuning" max_tokens=4000
  dbxcli rag context query="PostgreSQL replication setup" source=pg-docs`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			query := params["query"]
			if query == "" {
				return fmt.Errorf("query is required")
			}
			maxTokens := params["max_tokens"]
			if maxTokens == "" {
				maxTokens = "4000"
			}
			fmt.Printf("rag context: query=%q max_tokens=%s\n", query, maxTokens)
			return nil
		},
	}
}

func newRagIndexStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "index-status",
		Short:   "Show RAG index status for all sources",
		Long:    `Display the indexing status for each document source including document count, chunk count, and last index time.`,
		Example: `  dbxcli rag index-status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("rag index-status: showing index status...")
			return nil
		},
	}
}

func newRagIndexRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index-refresh",
		Short: "Re-index documents from a source",
		Long:  `Crawl a document source, chunk new/updated documents, generate embeddings, and store in the vector database.`,
		Example: `  dbxcli rag index-refresh source=oracle-docs
  dbxcli rag index-refresh source=runbooks`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := ParseNamedParams(args)
			if err != nil {
				return err
			}
			source := params["source"]
			if source == "" {
				return fmt.Errorf("source is required (oracle-docs, pg-docs, os-docs, mos-notes, runbooks)")
			}
			fmt.Printf("rag index-refresh: source=%s\n", source)
			return nil
		},
	}
}

func newRagSourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "sources",
		Short:   "List registered document sources",
		Long:    `List all registered document sources and their configuration.`,
		Example: `  dbxcli rag sources`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("rag sources: listing registered sources...")
			return nil
		},
	}
}
