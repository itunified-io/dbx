package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Searcher provides semantic and hybrid search over the vector store.
type Searcher struct {
	embedder Embedder
	store    VectorStore
}

// NewSearcher creates a searcher with the given embedder and store.
func NewSearcher(embedder Embedder, store VectorStore) *Searcher {
	return &Searcher{embedder: embedder, store: store}
}

// SemanticSearch embeds the query and searches the vector store.
func (s *Searcher) SemanticSearch(ctx context.Context, query string, topK int, filter map[string]string) ([]SearchResult, error) {
	vectors, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("semantic search embed: %w", err)
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("semantic search: no embedding returned")
	}
	return s.store.Search(ctx, vectors[0], topK, filter)
}

// HybridSearch combines vector similarity with keyword matching.
func (s *Searcher) HybridSearch(ctx context.Context, query string, topK int, filter map[string]string) ([]SearchResult, error) {
	// Get vector results
	vectorResults, err := s.SemanticSearch(ctx, query, topK*2, filter)
	if err != nil {
		return nil, err
	}

	// Re-rank with keyword boost
	keywords := strings.Fields(strings.ToLower(query))
	for i := range vectorResults {
		keywordScore := keywordMatchScore(vectorResults[i].Content, keywords)
		// Blend: 70% vector + 30% keyword
		vectorResults[i].Score = vectorResults[i].Score*0.7 + keywordScore*0.3
	}

	sort.Slice(vectorResults, func(i, j int) bool {
		return vectorResults[i].Score > vectorResults[j].Score
	})

	if topK > len(vectorResults) {
		topK = len(vectorResults)
	}
	return vectorResults[:topK], nil
}

func keywordMatchScore(content string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0
	}
	lower := strings.ToLower(content)
	matches := 0
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			matches++
		}
	}
	return float64(matches) / float64(len(keywords))
}
