package rag_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/rag"
	"github.com/stretchr/testify/assert"
)

func TestChunkDocument_SmallDoc(t *testing.T) {
	doc := rag.Document{
		ID: "doc-1", Source: "test", Title: "Test",
		Content: "This is a short document.",
	}
	chunks := rag.ChunkDocument(doc, 500, 50)
	assert.Len(t, chunks, 1)
	assert.Equal(t, "doc-1", chunks[0].DocumentID)
	assert.Contains(t, chunks[0].Content, "short document")
}

func TestChunkDocument_LargeDoc(t *testing.T) {
	content := ""
	for i := 0; i < 2000; i++ {
		content += "word "
	}
	doc := rag.Document{ID: "doc-2", Source: "test", Title: "Large", Content: content}
	chunks := rag.ChunkDocument(doc, 500, 50)
	assert.Greater(t, len(chunks), 1)
	for _, c := range chunks {
		assert.NotEmpty(t, c.Content)
		assert.LessOrEqual(t, c.TokenCount, 500)
	}
}

func TestChunkDocument_PreservesMetadata(t *testing.T) {
	doc := rag.Document{
		ID: "doc-3", Source: "oracle-docs", Title: "AWR Guide",
		Content:  "Some content about AWR reports.",
		Metadata: map[string]string{"version": "23ai", "section": "performance"},
	}
	chunks := rag.ChunkDocument(doc, 500, 50)
	assert.Equal(t, "oracle-docs", chunks[0].Source)
	assert.Equal(t, "23ai", chunks[0].Metadata["version"])
}

func TestChunkDocument_IDFormat(t *testing.T) {
	doc := rag.Document{ID: "doc-4", Source: "test", Content: "Hello world"}
	chunks := rag.ChunkDocument(doc, 500, 50)
	assert.Equal(t, "doc-4-0", chunks[0].ID)
}

func TestContentHash(t *testing.T) {
	h1 := rag.ContentHash("hello")
	h2 := rag.ContentHash("hello")
	h3 := rag.ContentHash("world")
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.Len(t, h1, 64) // SHA-256 hex
}
