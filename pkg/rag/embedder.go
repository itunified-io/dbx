package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIEmbedder uses OpenAI's embedding API.
type OpenAIEmbedder struct {
	apiKey string
	model  string
	dims   int
}

func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	dims := 1536
	if model == "text-embedding-3-large" {
		dims = 3072
	}
	return &OpenAIEmbedder{apiKey: apiKey, model: model, dims: dims}
}

func (e *OpenAIEmbedder) Dimensions() int { return e.dims }
func (e *OpenAIEmbedder) Model() string   { return e.model }

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body := map[string]any{"input": texts, "model": e.model}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embed: status %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai embed decode: %w", err)
	}

	vectors := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}
	return vectors, nil
}

// OllamaEmbedder uses a local Ollama instance.
type OllamaEmbedder struct {
	baseURL string
	model   string
	dims    int
}

func NewOllamaEmbedder(baseURL, model string) *OllamaEmbedder {
	dims := 768
	if model == "all-minilm" {
		dims = 384
	}
	return &OllamaEmbedder{baseURL: baseURL, model: model, dims: dims}
}

func (e *OllamaEmbedder) Dimensions() int { return e.dims }
func (e *OllamaEmbedder) Model() string   { return e.model }

func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		body := map[string]any{"model": e.model, "prompt": text}
		data, _ := json.Marshal(body)
		req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/embeddings", bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ollama embed: %w", err)
		}

		var result struct {
			Embedding []float32 `json:"embedding"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("ollama embed decode: %w", err)
		}
		resp.Body.Close()
		vectors[i] = result.Embedding
	}
	return vectors, nil
}

// LocalEmbedder uses all-MiniLM-L6-v2 via a local Ollama instance.
type LocalEmbedder struct{}

func NewLocalEmbedder() *LocalEmbedder { return &LocalEmbedder{} }

func (e *LocalEmbedder) Dimensions() int { return 384 }
func (e *LocalEmbedder) Model() string   { return "all-MiniLM-L6-v2" }

func (e *LocalEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	ollama := NewOllamaEmbedder("http://localhost:11434", "all-minilm")
	return ollama.Embed(ctx, texts)
}

// NewEmbedder creates an embedder based on provider config.
func NewEmbedder(provider, apiKey, baseURL, model string) Embedder {
	switch provider {
	case "openai":
		if model == "" {
			model = "text-embedding-3-small"
		}
		return NewOpenAIEmbedder(apiKey, model)
	case "ollama":
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		if model == "" {
			model = "nomic-embed-text"
		}
		return NewOllamaEmbedder(baseURL, model)
	default:
		return NewLocalEmbedder()
	}
}
