package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- In-Memory Vector Store (fallback) ---

type inMemoryEntry struct {
	ChunkID  string
	Vector   []float32
	Metadata map[string]string
}

// InMemoryStore is a simple in-memory vector store using brute-force cosine similarity.
type InMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*inMemoryEntry
	dims    int
}

func NewInMemoryStore(dims int) *InMemoryStore {
	return &InMemoryStore{entries: make(map[string]*inMemoryEntry), dims: dims}
}

func (s *InMemoryStore) Upsert(_ context.Context, chunkID string, vector []float32, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[chunkID] = &inMemoryEntry{ChunkID: chunkID, Vector: vector, Metadata: metadata}
	return nil
}

func (s *InMemoryStore) Search(_ context.Context, vector []float32, topK int, filter map[string]string) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		entry *inMemoryEntry
		score float64
	}
	var candidates []scored

	for _, e := range s.entries {
		if !matchesFilter(e.Metadata, filter) {
			continue
		}
		score := cosineSimilarity(vector, e.Vector)
		candidates = append(candidates, scored{entry: e, score: score})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if topK > len(candidates) {
		topK = len(candidates)
	}

	results := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		c := candidates[i]
		results[i] = SearchResult{
			ChunkID:    c.entry.ChunkID,
			DocumentID: c.entry.Metadata["document_id"],
			Source:     c.entry.Metadata["source"],
			Content:    c.entry.Metadata["content"],
			Score:      c.score,
			Metadata:   c.entry.Metadata,
		}
	}
	return results, nil
}

func (s *InMemoryStore) Delete(_ context.Context, chunkIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range chunkIDs {
		delete(s.entries, id)
	}
	return nil
}

func (s *InMemoryStore) Count(_ context.Context, source string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if source == "" {
		return len(s.entries), nil
	}
	count := 0
	for _, e := range s.entries {
		if e.Metadata["source"] == source {
			count++
		}
	}
	return count, nil
}

func matchesFilter(metadata, filter map[string]string) bool {
	if filter == nil {
		return true
	}
	for k, v := range filter {
		if metadata[k] != v {
			return false
		}
	}
	return true
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// --- PgVector Store (primary) ---

// PgVectorStore uses pgvector extension in PostgreSQL.
type PgVectorStore struct {
	pool *pgxpool.Pool
	dims int
}

// NewPgVectorStore creates a pgvector-backed store.
func NewPgVectorStore(pool *pgxpool.Pool, dims int) *PgVectorStore {
	return &PgVectorStore{pool: pool, dims: dims}
}

// EnsureSchema creates the embeddings table and ivfflat index.
func (s *PgVectorStore) EnsureSchema(ctx context.Context) error {
	sqls := []string{
		"CREATE EXTENSION IF NOT EXISTS vector",
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS dbx_rag_embeddings (
			chunk_id TEXT PRIMARY KEY,
			embedding vector(%d) NOT NULL,
			metadata JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`, s.dims),
		`CREATE INDEX IF NOT EXISTS idx_rag_embeddings_ivfflat
		 ON dbx_rag_embeddings USING ivfflat (embedding vector_cosine_ops)
		 WITH (lists = 100)`,
		`CREATE INDEX IF NOT EXISTS idx_rag_embeddings_source
		 ON dbx_rag_embeddings USING gin ((metadata->'source'))`,
	}
	for _, sql := range sqls {
		if _, err := s.pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("pgvector schema: %w", err)
		}
	}
	return nil
}

func (s *PgVectorStore) Upsert(ctx context.Context, chunkID string, vector []float32, metadata map[string]string) error {
	metaJSON := mapToJSON(metadata)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO dbx_rag_embeddings (chunk_id, embedding, metadata, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (chunk_id) DO UPDATE
		 SET embedding = EXCLUDED.embedding, metadata = EXCLUDED.metadata, updated_at = NOW()`,
		chunkID, pgvectorArray(vector), metaJSON,
	)
	return err
}

func (s *PgVectorStore) Search(ctx context.Context, vector []float32, topK int, filter map[string]string) ([]SearchResult, error) {
	query := `SELECT chunk_id, metadata, 1 - (embedding <=> $1) AS score
	          FROM dbx_rag_embeddings`
	args := []any{pgvectorArray(vector)}
	argIdx := 2

	if source, ok := filter["source"]; ok {
		query += fmt.Sprintf(" WHERE metadata->>'source' = $%d", argIdx)
		args = append(args, source)
		argIdx++
	}

	query += " ORDER BY embedding <=> $1 LIMIT $" + fmt.Sprintf("%d", argIdx)
	args = append(args, topK)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("pgvector search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var metaJSON []byte
		if err := rows.Scan(&r.ChunkID, &metaJSON, &r.Score); err != nil {
			return nil, err
		}
		r.Metadata = jsonToMap(metaJSON)
		r.DocumentID = r.Metadata["document_id"]
		r.Source = r.Metadata["source"]
		r.Content = r.Metadata["content"]
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *PgVectorStore) Delete(ctx context.Context, chunkIDs []string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM dbx_rag_embeddings WHERE chunk_id = ANY($1)", chunkIDs)
	return err
}

func (s *PgVectorStore) Count(ctx context.Context, source string) (int, error) {
	var count int
	var err error
	if source == "" {
		err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM dbx_rag_embeddings").Scan(&count)
	} else {
		err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM dbx_rag_embeddings WHERE metadata->>'source' = $1", source).Scan(&count)
	}
	return count, err
}

func pgvectorArray(v []float32) string {
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", f)
	}
	return s + "]"
}

func mapToJSON(m map[string]string) []byte {
	if m == nil {
		return []byte("{}")
	}
	data, _ := json.Marshal(m)
	return data
}

func jsonToMap(data []byte) map[string]string {
	m := make(map[string]string)
	_ = json.Unmarshal(data, &m)
	return m
}
