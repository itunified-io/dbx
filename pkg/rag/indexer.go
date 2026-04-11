package rag

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

// ChunkDocument splits a document into chunks of maxTokens with overlap.
func ChunkDocument(doc Document, maxTokens, overlap int) []Chunk {
	tokens := tokenize(doc.Content)
	if len(tokens) <= maxTokens {
		return []Chunk{{
			ID:         fmt.Sprintf("%s-0", doc.ID),
			DocumentID: doc.ID,
			Source:     doc.Source,
			Content:    doc.Content,
			Metadata:   copyMeta(doc.Metadata),
			TokenCount: len(tokens),
		}}
	}

	var chunks []Chunk
	idx := 0
	chunkNum := 0
	for idx < len(tokens) {
		end := idx + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}
		chunkTokens := tokens[idx:end]
		content := strings.Join(chunkTokens, " ")
		chunks = append(chunks, Chunk{
			ID:         fmt.Sprintf("%s-%d", doc.ID, chunkNum),
			DocumentID: doc.ID,
			Source:     doc.Source,
			Content:    content,
			Metadata:   copyMeta(doc.Metadata),
			TokenCount: len(chunkTokens),
		})
		chunkNum++
		idx = end - overlap
		if idx <= 0 || end == len(tokens) {
			break
		}
	}
	return chunks
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	})
}

func copyMeta(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Indexer manages the full indexing pipeline: crawl -> chunk -> embed -> store.
type Indexer struct {
	embedder Embedder
	store    VectorStore
	sources  map[string]DocumentSource
}

// NewIndexer creates an indexer with the given embedder and store.
func NewIndexer(embedder Embedder, store VectorStore) *Indexer {
	return &Indexer{
		embedder: embedder,
		store:    store,
		sources:  make(map[string]DocumentSource),
	}
}

// RegisterSource adds a document source.
func (idx *Indexer) RegisterSource(src DocumentSource) {
	idx.sources[src.Name()] = src
}

// IndexSource crawls a source, chunks documents, embeds, and stores.
func (idx *Indexer) IndexSource(ctx context.Context, sourceName string) (*IndexStats, error) {
	src, ok := idx.sources[sourceName]
	if !ok {
		return nil, fmt.Errorf("unknown source: %s", sourceName)
	}

	docs, err := src.Crawl(ctx)
	if err != nil {
		return nil, fmt.Errorf("crawl %s: %w", sourceName, err)
	}

	stats := &IndexStats{Source: sourceName}
	for _, doc := range docs {
		chunks := ChunkDocument(doc, 500, 50)
		stats.Documents++
		stats.Chunks += len(chunks)

		texts := make([]string, len(chunks))
		for i, c := range chunks {
			texts[i] = c.Content
		}
		vectors, err := idx.embedder.Embed(ctx, texts)
		if err != nil {
			return nil, fmt.Errorf("embed %s: %w", doc.ID, err)
		}

		for i, chunk := range chunks {
			meta := chunk.Metadata
			if meta == nil {
				meta = make(map[string]string)
			}
			meta["source"] = chunk.Source
			meta["document_id"] = chunk.DocumentID
			meta["content"] = chunk.Content
			if err := idx.store.Upsert(ctx, chunk.ID, vectors[i], meta); err != nil {
				return nil, fmt.Errorf("store %s: %w", chunk.ID, err)
			}
		}
	}
	return stats, nil
}

// IndexStats reports indexing progress.
type IndexStats struct {
	Source    string `json:"source"`
	Documents int   `json:"documents"`
	Chunks    int   `json:"chunks"`
}

// ContentHash returns SHA-256 of content for deduplication.
func ContentHash(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}
