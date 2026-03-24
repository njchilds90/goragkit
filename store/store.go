// Package store defines the VectorStore interface and an in-memory implementation.
package store

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/rerrors"
)

// QueryOptions controls retrieval behavior.
type QueryOptions struct {
	TopK            int
	MetadataFilters map[string]string
}

// VectorStore persists and retrieves chunk embeddings.
type VectorStore interface {
	// Upsert stores chunks with their embedding vectors.
	Upsert(ctx context.Context, chunks []document.Chunk, vectors [][]float64) error
	// Query returns the most similar chunks to the query vector.
	Query(ctx context.Context, vector []float64, opts QueryOptions) ([]document.ScoredChunk, error)
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
	if len(chunks) != len(vectors) {
		return rerrors.Wrap(rerrors.InvalidInput, "memory.upsert", fmt.Sprintf("chunks(%d) != vectors(%d)", len(chunks), len(vectors)), nil)
	}
	for i := range chunks {
		m.entries = append(m.entries, entry{chunk: chunks[i], vector: vectors[i]})
	}
	return nil
}

// Query implements VectorStore.
func (m *Memory) Query(_ context.Context, vector []float64, opts QueryOptions) ([]document.ScoredChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if opts.TopK <= 0 {
		opts.TopK = 5
	}
	results := make([]document.ScoredChunk, 0, len(m.entries))
	for _, e := range m.entries {
		if !metadataMatch(e.chunk.Metadata, opts.MetadataFilters) {
			continue
		}
		results = append(results, document.ScoredChunk{Chunk: e.chunk, Score: Cosine(vector, e.vector)})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if opts.TopK > len(results) {
		opts.TopK = len(results)
	}
	return results[:opts.TopK], nil
}

func metadataMatch(meta map[string]string, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	for k, v := range filters {
		if meta[k] != v {
			return false
		}
	}
	return true
}

// Cosine computes cosine similarity between two vectors.
func Cosine(a, b []float64) float64 {
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
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
