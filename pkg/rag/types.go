// Package rag provides retrieval-augmented generation for AI-assisted database management.
package rag

import (
	"context"
	"time"
)

// Document represents a source document before chunking.
type Document struct {
	ID       string            `json:"id"`
	Source   string            `json:"source"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Updated  time.Time         `json:"updated"`
}

// Chunk is a document fragment ready for embedding.
type Chunk struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id"`
	Source     string            `json:"source"`
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata"`
	TokenCount int               `json:"token_count"`
}

// Embedding is a vector representation of a chunk.
type Embedding struct {
	ChunkID    string    `json:"chunk_id"`
	Vector     []float32 `json:"vector"`
	Dimensions int       `json:"dimensions"`
	Model      string    `json:"model"`
}

// SearchResult is a single search hit with relevance score.
type SearchResult struct {
	ChunkID    string            `json:"chunk_id"`
	DocumentID string            `json:"document_id"`
	Source     string            `json:"source"`
	Content    string            `json:"content"`
	Score      float64           `json:"score"`
	Metadata   map[string]string `json:"metadata"`
}

// Embedder generates vector embeddings from text.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
	Model() string
}

// VectorStore persists and queries embeddings.
type VectorStore interface {
	Upsert(ctx context.Context, chunkID string, vector []float32, metadata map[string]string) error
	Search(ctx context.Context, vector []float32, topK int, filter map[string]string) ([]SearchResult, error)
	Delete(ctx context.Context, chunkIDs []string) error
	Count(ctx context.Context, source string) (int, error)
}

// DocumentSource crawls and yields documents for indexing.
type DocumentSource interface {
	Name() string
	Crawl(ctx context.Context) ([]Document, error)
}
