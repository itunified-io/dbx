package rag

import (
	"fmt"
	"strings"
)

// ContextBuilder assembles token-limited prompts from search results.
type ContextBuilder struct {
	maxTokens int
}

// NewContextBuilder creates a context builder with the given token limit.
func NewContextBuilder(maxTokens int) *ContextBuilder {
	if maxTokens <= 0 {
		maxTokens = 4000
	}
	return &ContextBuilder{maxTokens: maxTokens}
}

// BuildContext creates an LLM context string from search results.
func (cb *ContextBuilder) BuildContext(results []SearchResult, query string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Relevant Context for: %s\n\n", query))

	tokenCount := len(tokenize(b.String()))
	for i, r := range results {
		section := fmt.Sprintf("### Source: %s (score: %.2f)\n%s\n\n", r.Source, r.Score, r.Content)
		sectionTokens := len(tokenize(section))
		if tokenCount+sectionTokens > cb.maxTokens {
			if i == 0 {
				// Always include at least one result, truncated
				truncated := truncateToTokens(section, cb.maxTokens-tokenCount)
				b.WriteString(truncated)
			}
			break
		}
		b.WriteString(section)
		tokenCount += sectionTokens
	}
	return b.String()
}

func truncateToTokens(text string, maxTokens int) string {
	tokens := tokenize(text)
	if len(tokens) <= maxTokens {
		return text
	}
	return strings.Join(tokens[:maxTokens], " ") + "..."
}
