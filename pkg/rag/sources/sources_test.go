package sources_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/rag/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOracleDocsSource_CrawlEmptyDir(t *testing.T) {
	src := sources.NewOracleDocsSource("/nonexistent", []string{"19c"})
	assert.Equal(t, "oracle-docs", src.Name())
	docs, err := src.Crawl(context.Background())
	require.NoError(t, err)
	assert.Empty(t, docs)
}

func TestPgDocsSource_CrawlWithFiles(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "vacuum.md"), []byte("# Vacuum Tuning\nContent here"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "ignore.go"), []byte("package x"), 0644)
	require.NoError(t, err)

	src := sources.NewPgDocsSource(dir, []string{"17", "18"})
	assert.Equal(t, "pg-docs", src.Name())
	docs, err := src.Crawl(context.Background())
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "vacuum", docs[0].Title)
	assert.Contains(t, docs[0].Content, "Vacuum Tuning")
	assert.Equal(t, "17,18", docs[0].Metadata["versions"])
}

func TestRunbooksSource_Crawl(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "backup.md"), []byte("# Backup Procedure"), 0644)
	require.NoError(t, err)

	src := sources.NewRunbooksSource(dir)
	assert.Equal(t, "runbooks", src.Name())
	docs, err := src.Crawl(context.Background())
	require.NoError(t, err)
	assert.Len(t, docs, 1)
}

func TestCustomSource(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("Custom notes"), 0644)
	require.NoError(t, err)

	src := sources.NewCustomSource("my-notes", dir)
	assert.Equal(t, "my-notes", src.Name())
	docs, err := src.Crawl(context.Background())
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "my-notes", docs[0].Source)
}

func TestOSDocsSource(t *testing.T) {
	src := sources.NewOSDocsSource("/nonexistent")
	assert.Equal(t, "os-docs", src.Name())
}

func TestMOSNotesSource(t *testing.T) {
	src := sources.NewMOSNotesSource("/nonexistent")
	assert.Equal(t, "mos-notes", src.Name())
}
