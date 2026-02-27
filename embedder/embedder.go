// Package embedder defines the Embedder interface and built-in adapters.
package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Embedder converts text into a float64 vector.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// OpenAI implements Embedder using the OpenAI embeddings API.
type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAI returns an OpenAI embedder.
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{apiKey: apiKey, model: model, client: &http.Client{}}
}

type openAIRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed implements Embedder.
func (o *OpenAI) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	body, _ := json.Marshal(openAIRequest{Input: texts, Model: o.model})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to surface any API-provided error message.
		if out.Error != nil {
			return nil, fmt.Errorf("openai: %s", out.Error.Message)
		}
		return nil, fmt.Errorf("openai: unexpected status %d", resp.StatusCode)
	}
	if out.Error != nil {
		return nil, fmt.Errorf("openai: %s", out.Error.Message)
	}

	vecs := make([][]float64, len(out.Data))
	for i, d := range out.Data {
		vecs[i] = d.Embedding
	}
	return vecs, nil
}

// Ollama implements Embedder using a local Ollama instance.
type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama returns an Ollama embedder. baseURL defaults to http://localhost:11434.
func NewOllama(baseURL, model string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{baseURL: baseURL, model: model, client: &http.Client{}}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Embed implements Embedder. Ollama embeds one text at a time.
func (o *Ollama) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	vecs := make([][]float64, 0, len(texts))
	for _, t := range texts {
		body, _ := json.Marshal(ollamaRequest{Model: o.model, Prompt: t})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			o.baseURL+"/api/embeddings", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := o.client.Do(req)
		if err != nil {
			return nil, err
		}
		// Ensure the response body is closed to avoid leaks.
		defer resp.Body.Close()
		// Check HTTP status before decoding.
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("ollama: unexpected status %d", resp.StatusCode)
		}
		var out ollamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
		vecs = append(vecs, out.Embedding)
	}
	return vecs, nil
}