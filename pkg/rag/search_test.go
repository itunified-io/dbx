package rag_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEmbedder struct {
	vectors map[string][]float32
}

func (m *mockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, t := range texts {
		if v, ok := m.vectors[t]; ok {
			result[i] = v
		} else {
			result[i] = []float32{0.5, 0.5, 0.5}
		}
	}
	return result, nil
}
func (m *mockEmbedder) Dimensions() int { return 3 }
func (m *mockEmbedder) Model() string   { return "mock" }

func setupSearchTest(t *testing.T) (*rag.Searcher, *rag.InMemoryStore) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()
	_ = store.Upsert(ctx, "c1", []float32{1, 0, 0}, map[string]string{
		"source": "oracle-docs", "content": "AWR report generation and snapshot management",
	})
	_ = store.Upsert(ctx, "c2", []float32{0, 1, 0}, map[string]string{
		"source": "pg-docs", "content": "PostgreSQL vacuum tuning and autovacuum settings",
	})
	_ = store.Upsert(ctx, "c3", []float32{0.9, 0.1, 0}, map[string]string{
		"source": "oracle-docs", "content": "AWR snapshot interval configuration",
	})

	embedder := &mockEmbedder{vectors: map[string][]float32{
		"AWR report": {1, 0, 0},
		"vacuum":     {0, 1, 0},
	}}

	return rag.NewSearcher(embedder, store), store
}

func TestSemanticSearch(t *testing.T) {
	searcher, _ := setupSearchTest(t)
	results, err := searcher.SemanticSearch(context.Background(), "AWR report", 2, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "c1", results[0].ChunkID)
}

func TestSemanticSearch_WithFilter(t *testing.T) {
	searcher, _ := setupSearchTest(t)
	results, err := searcher.SemanticSearch(context.Background(), "AWR report", 10, map[string]string{"source": "pg-docs"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "c2", results[0].ChunkID)
}

func TestHybridSearch(t *testing.T) {
	searcher, _ := setupSearchTest(t)
	results, err := searcher.HybridSearch(context.Background(), "AWR report", 2, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	// c1 should still be top since it matches both vector and keyword
	assert.Equal(t, "c1", results[0].ChunkID)
}
