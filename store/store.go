// Package store defines the VectorStore interface and an in-memory implementation.
package store

import (
	"context"
	"sync"

	"github.com/YOUR_USERNAME/goragkit/document"
)

// VectorStore persists and retrieves chunk embeddings.
type VectorStore interface {
	// Upsert stores chunks with their embedding vectors.
	Upsert(ctx context.Context, chunks []document.Chunk, vectors [][]float64) error
	// Query returns the topK most similar chunks to the query vector.
	Query(ctx context.Context, vector []float64, topK int) ([]document.ScoredChunk, error)
}

type entry struct {
	chunk  document.Chunk
	vector []float64
}

// Memory is a thread-safe in-memory VectorStore using cosine similarity.
type Memory struct {
	mu      sync.RWMutex
	entries []entry
}

// NewMemory returns an initialised Memory store.
func NewMemory() *Memory { return &Memory{} }

// Upsert implements VectorStore.
func (m *Memory) Upsert(_ context.Context, chunks []document.Chunk, vectors [][]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, c := range chunks {
		m.entries = append(m.entries, entry{chunk: c, vector: vectors[i]})
	}
	return nil
}

// Query implements VectorStore.
func (m *Memory) Query(_ context.Context, vector []float64, topK int) ([]document.ScoredChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type scored struct {
		sc    document.ScoredChunk
		score float64
	}
	results := make([]scored, 0, len(m.entries))
	for _, e := range m.entries {
		s := cosine(vector, e.vector)
		results = append(results, scored{
			sc:    document.ScoredChunk{Chunk: e.chunk, Score: s},
			score: s,
		})
	}
	// partial sort: bubble top-k to front
	for i := 0; i < topK && i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if topK > len(results) {
		topK = len(results)
	}
	out := make([]document.ScoredChunk, topK)
	for i := range out {
		out[i] = results[i].sc
	}
	return out, nil
}

func cosine(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
