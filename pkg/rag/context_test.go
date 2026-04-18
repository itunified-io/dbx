package rag_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/rag"
	"github.com/stretchr/testify/assert"
)

func TestContextBuilder_BuildContext(t *testing.T) {
	results := []rag.SearchResult{
		{ChunkID: "c1", Source: "oracle-docs", Content: "AWR report generation guide", Score: 0.95},
		{ChunkID: "c2", Source: "pg-docs", Content: "PostgreSQL vacuum tuning", Score: 0.80},
	}

	cb := rag.NewContextBuilder(4000)
	ctx := cb.BuildContext(results, "How to generate AWR reports?")
	assert.Contains(t, ctx, "AWR report generation guide")
	assert.Contains(t, ctx, "oracle-docs")
	assert.Contains(t, ctx, "0.95")
}

func TestContextBuilder_TokenLimit(t *testing.T) {
	// Create a result with lots of content
	longContent := ""
	for i := 0; i < 500; i++ {
		longContent += "word "
	}
	results := []rag.SearchResult{
		{ChunkID: "c1", Source: "test", Content: longContent, Score: 0.9},
		{ChunkID: "c2", Source: "test", Content: "short content", Score: 0.8},
	}

	cb := rag.NewContextBuilder(100)
	ctx := cb.BuildContext(results, "test query")
	// Should include at least first result (truncated) but not second
	assert.Contains(t, ctx, "word")
	assert.NotContains(t, ctx, "short content")
}

func TestContextBuilder_DefaultTokens(t *testing.T) {
	cb := rag.NewContextBuilder(0) // should default to 4000
	results := []rag.SearchResult{
		{ChunkID: "c1", Source: "test", Content: "hello", Score: 0.9},
	}
	ctx := cb.BuildContext(results, "query")
	assert.Contains(t, ctx, "hello")
}
