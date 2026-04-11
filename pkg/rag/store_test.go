package rag_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStore_UpsertAndSearch(t *testing.T) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()

	err := store.Upsert(ctx, "chunk-1", []float32{1.0, 0.0, 0.0}, map[string]string{
		"source": "oracle-docs", "content": "AWR report generation",
	})
	require.NoError(t, err)

	err = store.Upsert(ctx, "chunk-2", []float32{0.0, 1.0, 0.0}, map[string]string{
		"source": "pg-docs", "content": "PostgreSQL vacuum tuning",
	})
	require.NoError(t, err)

	err = store.Upsert(ctx, "chunk-3", []float32{0.9, 0.1, 0.0}, map[string]string{
		"source": "oracle-docs", "content": "AWR snapshot interval",
	})
	require.NoError(t, err)

	results, err := store.Search(ctx, []float32{1.0, 0.0, 0.0}, 2, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "chunk-1", results[0].ChunkID)
	assert.Equal(t, "chunk-3", results[1].ChunkID)
}

func TestInMemoryStore_SearchWithFilter(t *testing.T) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()

	_ = store.Upsert(ctx, "c1", []float32{1, 0, 0}, map[string]string{"source": "oracle-docs"})
	_ = store.Upsert(ctx, "c2", []float32{0.9, 0.1, 0}, map[string]string{"source": "pg-docs"})
	_ = store.Upsert(ctx, "c3", []float32{0.8, 0.2, 0}, map[string]string{"source": "oracle-docs"})

	results, err := store.Search(ctx, []float32{1, 0, 0}, 10, map[string]string{"source": "oracle-docs"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()
	_ = store.Upsert(ctx, "c1", []float32{1, 0, 0}, nil)
	_ = store.Upsert(ctx, "c2", []float32{0, 1, 0}, nil)

	err := store.Delete(ctx, []string{"c1"})
	require.NoError(t, err)

	count, _ := store.Count(ctx, "")
	assert.Equal(t, 1, count)
}

func TestInMemoryStore_Count(t *testing.T) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()
	_ = store.Upsert(ctx, "c1", []float32{1, 0, 0}, map[string]string{"source": "oracle-docs"})
	_ = store.Upsert(ctx, "c2", []float32{0, 1, 0}, map[string]string{"source": "pg-docs"})

	total, _ := store.Count(ctx, "")
	assert.Equal(t, 2, total)

	oracleCount, _ := store.Count(ctx, "oracle-docs")
	assert.Equal(t, 1, oracleCount)
}

func TestInMemoryStore_Upsert_Overwrites(t *testing.T) {
	store := rag.NewInMemoryStore(3)
	ctx := context.Background()
	_ = store.Upsert(ctx, "c1", []float32{1, 0, 0}, map[string]string{"version": "1"})
	_ = store.Upsert(ctx, "c1", []float32{0, 1, 0}, map[string]string{"version": "2"})

	count, _ := store.Count(ctx, "")
	assert.Equal(t, 1, count)

	results, _ := store.Search(ctx, []float32{0, 1, 0}, 1, nil)
	assert.Equal(t, "c1", results[0].ChunkID)
}

func TestPgVectorStore_Interface(t *testing.T) {
	var _ rag.VectorStore = &rag.PgVectorStore{}
}
