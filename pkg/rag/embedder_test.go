package rag_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/rag"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIEmbedder_Interface(t *testing.T) {
	var _ rag.Embedder = &rag.OpenAIEmbedder{}
	e := rag.NewOpenAIEmbedder("test-key", "text-embedding-3-small")
	assert.Equal(t, 1536, e.Dimensions())
	assert.Equal(t, "text-embedding-3-small", e.Model())
}

func TestOpenAIEmbedder_LargeModel(t *testing.T) {
	e := rag.NewOpenAIEmbedder("test-key", "text-embedding-3-large")
	assert.Equal(t, 3072, e.Dimensions())
}

func TestOllamaEmbedder_Interface(t *testing.T) {
	var _ rag.Embedder = &rag.OllamaEmbedder{}
	e := rag.NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	assert.Equal(t, 768, e.Dimensions())
	assert.Equal(t, "nomic-embed-text", e.Model())
}

func TestLocalEmbedder_Interface(t *testing.T) {
	var _ rag.Embedder = &rag.LocalEmbedder{}
	e := rag.NewLocalEmbedder()
	assert.Equal(t, 384, e.Dimensions())
	assert.Equal(t, "all-MiniLM-L6-v2", e.Model())
}

func TestNewEmbedder_OpenAI(t *testing.T) {
	e := rag.NewEmbedder("openai", "key", "", "")
	assert.Equal(t, "text-embedding-3-small", e.Model())
}

func TestNewEmbedder_Ollama(t *testing.T) {
	e := rag.NewEmbedder("ollama", "", "", "")
	assert.Equal(t, "nomic-embed-text", e.Model())
}

func TestNewEmbedder_Local(t *testing.T) {
	e := rag.NewEmbedder("local", "", "", "")
	assert.Equal(t, "all-MiniLM-L6-v2", e.Model())
}

func TestNewEmbedder_Unknown(t *testing.T) {
	e := rag.NewEmbedder("unknown", "", "", "")
	assert.Equal(t, "all-MiniLM-L6-v2", e.Model())
}
