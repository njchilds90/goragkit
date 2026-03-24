// Package embedder defines the Embedder interface and built-in adapters.
package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/njchilds90/goragkit/cache"
	"github.com/njchilds90/goragkit/rerrors"
)

// Embedder converts text into float64 vectors.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// HTTPClient is the subset of *http.Client used by embedders.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// OpenAI implements Embedder using the OpenAI embeddings API.
type OpenAI struct {
	apiKey string
	model  string
	client HTTPClient
}

// NewOpenAI returns an OpenAI embedder.
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{apiKey: apiKey, model: model, client: &http.Client{Timeout: 30 * time.Second}}
}

// WithHTTPClient sets a custom HTTP client.
func (o *OpenAI) WithHTTPClient(client HTTPClient) *OpenAI {
	o.client = client
	return o
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
	return embedJSON(ctx, o.client, "POST", "https://api.openai.com/v1/embeddings", openAIRequest{Input: texts, Model: o.model}, map[string]string{
		"Authorization": "Bearer " + o.apiKey,
		"Content-Type":  "application/json",
	}, func(raw []byte) ([][]float64, error) {
		var out openAIResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		if out.Error != nil {
			return nil, fmt.Errorf("openai: %s", out.Error.Message)
		}
		vecs := make([][]float64, len(out.Data))
		for i, d := range out.Data {
			vecs[i] = d.Embedding
		}
		return vecs, nil
	})
}

// Ollama implements Embedder using a local Ollama instance.
type Ollama struct {
	baseURL string
	model   string
	client  HTTPClient
}

// NewOllama returns an Ollama embedder. baseURL defaults to http://localhost:11434.
func NewOllama(baseURL, model string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{baseURL: strings.TrimSuffix(baseURL, "/"), model: model, client: &http.Client{Timeout: 30 * time.Second}}
}

// WithHTTPClient sets a custom HTTP client.
func (o *Ollama) WithHTTPClient(client HTTPClient) *Ollama {
	o.client = client
	return o
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
		v, err := embedJSON(ctx, o.client, "POST", o.baseURL+"/api/embeddings", ollamaRequest{Model: o.model, Prompt: t}, map[string]string{"Content-Type": "application/json"}, func(raw []byte) ([][]float64, error) {
			var out ollamaResponse
			if err := json.Unmarshal(raw, &out); err != nil {
				return nil, err
			}
			return [][]float64{out.Embedding}, nil
		})
		if err != nil {
			return nil, err
		}
		vecs = append(vecs, v[0])
	}
	return vecs, nil
}

// Cohere implements Embedder using the Cohere API.
type Cohere struct {
	apiKey string
	model  string
	client HTTPClient
}

// NewCohere returns a Cohere embedder.
func NewCohere(apiKey, model string) *Cohere {
	return &Cohere{apiKey: apiKey, model: model, client: &http.Client{Timeout: 30 * time.Second}}
}

type cohereRequest struct {
	Texts []string `json:"texts"`
	Model string   `json:"model"`
}

type cohereResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Message    string      `json:"message,omitempty"`
}

// Embed implements Embedder.
func (c *Cohere) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	return embedJSON(ctx, c.client, "POST", "https://api.cohere.com/v1/embed", cohereRequest{Texts: texts, Model: c.model}, map[string]string{
		"Authorization": "Bearer " + c.apiKey,
		"Content-Type":  "application/json",
	}, func(raw []byte) ([][]float64, error) {
		var out cohereResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		if out.Message != "" {
			return nil, fmt.Errorf("cohere: %s", out.Message)
		}
		return out.Embeddings, nil
	})
}

// Cached wraps an Embedder with an in-memory cache.
type Cached struct {
	next  Embedder
	cache *cache.LRU[string, []float64]
}

// NewCached returns an Embedder that caches by exact text match.
func NewCached(next Embedder, size int) *Cached {
	return &Cached{next: next, cache: cache.NewLRU[string, []float64](size)}
}

// Embed implements Embedder.
func (c *Cached) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	missing := make([]string, 0)
	missingIndex := make([]int, 0)
	for i, t := range texts {
		if v, ok := c.cache.Get(t); ok {
			out[i] = v
			continue
		}
		missing = append(missing, t)
		missingIndex = append(missingIndex, i)
	}
	if len(missing) == 0 {
		return out, nil
	}
	vecs, err := c.next.Embed(ctx, missing)
	if err != nil {
		return nil, err
	}
	if len(vecs) != len(missing) {
		return nil, rerrors.Wrap(rerrors.External, "cached_embedder.embed", "unexpected vector count", nil)
	}
	for i, idx := range missingIndex {
		out[idx] = vecs[i]
		c.cache.Put(missing[i], vecs[i])
	}
	return out, nil
}

func embedJSON(ctx context.Context, client HTTPClient, method, url string, payload any, headers map[string]string, decode func([]byte) ([][]float64, error)) ([][]float64, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, rerrors.Wrap(rerrors.InvalidInput, "embed_json.marshal", "marshal payload", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, rerrors.Wrap(rerrors.InvalidInput, "embed_json.request", "build request", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, rerrors.Wrap(rerrors.External, "embed_json.do", "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var raw bytes.Buffer
	if _, err := raw.ReadFrom(resp.Body); err != nil {
		return nil, rerrors.Wrap(rerrors.External, "embed_json.read", "read response", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, rerrors.Wrap(rerrors.External, "embed_json.status", fmt.Sprintf("status %d", resp.StatusCode), nil)
	}
	vecs, err := decode(raw.Bytes())
	if err != nil {
		return nil, rerrors.Wrap(rerrors.External, "embed_json.decode", "decode response", err)
	}
	return vecs, nil
}
