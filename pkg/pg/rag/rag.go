// Package rag provides PostgreSQL RAG (Retrieval-Augmented Generation) tools.
package rag

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// VectorStoreStatusResult represents pgvector store status.
type VectorStoreStatusResult struct {
	PgvectorVersion string `json:"pgvector_version"`
	DocumentCount   int64  `json:"document_count"`
	TotalChunks     int64  `json:"total_chunks"`
	IndexSize       string `json:"index_size"`
}

// ExceptionMatchResult represents a matched error pattern.
type ExceptionMatchResult struct {
	Category    string `json:"category"`
	Pattern     string `json:"pattern"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

// FileIndexResult represents file indexing status.
type FileIndexResult struct {
	FilesIndexed int64 `json:"files_indexed"`
	ChunksStored int64 `json:"chunks_stored"`
}

// VectorStoreStatus checks pgvector extension and document store status.
func VectorStoreStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) (*VectorStoreStatusResult, error) {
	var version string
	row := pginternal.QueryRow(ctx, q, "SELECT extversion FROM pg_extension WHERE extname = 'vector'")
	if err := row.Scan(&version); err != nil {
		return nil, fmt.Errorf("pgvector not installed: %w", err)
	}
	var result VectorStoreStatusResult
	result.PgvectorVersion = version
	row2 := pginternal.QueryRow(ctx, q, "SELECT count(DISTINCT source_file), count(*) FROM dbx_rag_chunks")
	if err := row2.Scan(&result.DocumentCount, &result.TotalChunks); err != nil {
		return &result, nil // table may not exist yet
	}
	return &result, nil
}

// ExceptionMatch finds the best matching error pattern.
func ExceptionMatch(ctx context.Context, q pginternal.Querier, params map[string]any) (*ExceptionMatchResult, error) {
	errorText, _ := params["error_text"].(string)
	row := pginternal.QueryRow(ctx, q,
		"SELECT category, pattern, action, description FROM dbx_rag_exceptions WHERE $1 ILIKE '%' || pattern || '%' ORDER BY length(pattern) DESC LIMIT 1",
		errorText)
	var r ExceptionMatchResult
	if err := row.Scan(&r.Category, &r.Pattern, &r.Action, &r.Description); err != nil {
		return nil, fmt.Errorf("no matching exception pattern: %w", err)
	}
	return &r, nil
}

// FileIndex returns indexing status.
func FileIndex(ctx context.Context, q pginternal.Querier, _ map[string]any) (*FileIndexResult, error) {
	row := pginternal.QueryRow(ctx, q,
		"SELECT count(DISTINCT source_file), count(*) FROM dbx_rag_chunks")
	var r FileIndexResult
	if err := row.Scan(&r.FilesIndexed, &r.ChunksStored); err != nil {
		return nil, fmt.Errorf("pg file index: %w", err)
	}
	return &r, nil
}

// FileSearch searches indexed documents.
func FileSearch(ctx context.Context, q pginternal.Querier, params map[string]any) ([]map[string]any, error) {
	query, _ := params["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}
	limit := 10
	if v, ok := params["limit"].(int); ok {
		limit = v
	}

	rows, err := pginternal.QueryRows(ctx, q,
		`SELECT source_file, chunk_text, chunk_index
		 FROM dbx_rag_chunks
		 WHERE chunk_text ILIKE '%' || $1 || '%'
		 LIMIT $2`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("pg file search: %w", err)
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var sourceFile, chunkText string
		var chunkIndex int
		if err := rows.Scan(&sourceFile, &chunkText, &chunkIndex); err != nil {
			return nil, fmt.Errorf("pg file search scan: %w", err)
		}
		results = append(results, map[string]any{
			"source_file": sourceFile,
			"chunk_text":  chunkText,
			"chunk_index": chunkIndex,
		})
	}
	return results, rows.Err()
}
