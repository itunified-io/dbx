// Package sources provides document source implementations for the RAG indexer.
package sources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/itunified-io/dbx/pkg/rag"
)

// OracleDocsSource crawls Oracle documentation files.
type OracleDocsSource struct {
	docsDir  string
	versions []string
}

func NewOracleDocsSource(docsDir string, versions []string) *OracleDocsSource {
	return &OracleDocsSource{docsDir: docsDir, versions: versions}
}

func (s *OracleDocsSource) Name() string { return "oracle-docs" }

func (s *OracleDocsSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.docsDir, "oracle-docs", s.versions)
}

// PgDocsSource crawls PostgreSQL documentation files.
type PgDocsSource struct {
	docsDir  string
	versions []string
}

func NewPgDocsSource(docsDir string, versions []string) *PgDocsSource {
	return &PgDocsSource{docsDir: docsDir, versions: versions}
}

func (s *PgDocsSource) Name() string { return "pg-docs" }

func (s *PgDocsSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.docsDir, "pg-docs", s.versions)
}

// OSDocsSource crawls Linux/OS documentation files.
type OSDocsSource struct {
	docsDir string
}

func NewOSDocsSource(docsDir string) *OSDocsSource {
	return &OSDocsSource{docsDir: docsDir}
}

func (s *OSDocsSource) Name() string { return "os-docs" }

func (s *OSDocsSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.docsDir, "os-docs", nil)
}

// MOSNotesSource reads MOS knowledge base articles from mos-doc-sync downloads.
type MOSNotesSource struct {
	notesDir string
}

func NewMOSNotesSource(notesDir string) *MOSNotesSource {
	return &MOSNotesSource{notesDir: notesDir}
}

func (s *MOSNotesSource) Name() string { return "mos-notes" }

func (s *MOSNotesSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.notesDir, "mos-notes", nil)
}

// RunbooksSource reads customer runbooks (markdown files).
type RunbooksSource struct {
	runbooksDir string
}

func NewRunbooksSource(runbooksDir string) *RunbooksSource {
	return &RunbooksSource{runbooksDir: runbooksDir}
}

func (s *RunbooksSource) Name() string { return "runbooks" }

func (s *RunbooksSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.runbooksDir, "runbooks", nil)
}

// CustomSource reads documents from a user-specified directory.
type CustomSource struct {
	dir  string
	name string
}

func NewCustomSource(name, dir string) *CustomSource {
	return &CustomSource{name: name, dir: dir}
}

func (s *CustomSource) Name() string { return s.name }

func (s *CustomSource) Crawl(ctx context.Context) ([]rag.Document, error) {
	return crawlDirectory(ctx, s.dir, s.name, nil)
}

func crawlDirectory(_ context.Context, dir, source string, versions []string) ([]rag.Document, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // empty is OK
	}

	var docs []rag.Document
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" && ext != ".html" && ext != ".rst" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		meta := map[string]string{
			"path": relPath,
		}
		if len(versions) > 0 {
			meta["versions"] = strings.Join(versions, ",")
		}

		docs = append(docs, rag.Document{
			ID:       fmt.Sprintf("%s:%s", source, relPath),
			Source:   source,
			Title:    strings.TrimSuffix(filepath.Base(path), ext),
			Content:  string(data),
			Metadata: meta,
			Updated:  info.ModTime(),
		})
		return nil
	})
	return docs, err
}

// RegisterAllSources registers all built-in document sources with an indexer.
func RegisterAllSources(idx *rag.Indexer, baseDir string) {
	idx.RegisterSource(NewOracleDocsSource(filepath.Join(baseDir, "oracle"), []string{"19c", "21c", "23ai"}))
	idx.RegisterSource(NewPgDocsSource(filepath.Join(baseDir, "pg"), []string{"14", "15", "16", "17", "18"}))
	idx.RegisterSource(NewOSDocsSource(filepath.Join(baseDir, "os")))
	idx.RegisterSource(NewMOSNotesSource(filepath.Join(baseDir, "mos")))
	idx.RegisterSource(NewRunbooksSource(filepath.Join(baseDir, "runbooks")))
}

// SourceStatus holds status information about a document source.
type SourceStatus struct {
	Name      string    `json:"name"`
	DocCount  int       `json:"doc_count"`
	ChunkCount int      `json:"chunk_count"`
	LastIndex time.Time `json:"last_index"`
}
