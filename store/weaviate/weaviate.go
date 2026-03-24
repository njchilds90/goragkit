// Package weaviate provides a Weaviate VectorStore adapter.
package weaviate

import (
	"context"
	"fmt"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/store"
)

// Client is the contract used by this adapter.
type Client interface {
	Upsert(ctx context.Context, className string, chunks []document.Chunk, vectors [][]float64) error
	NearVector(ctx context.Context, className string, vector []float64, topK int, filters map[string]string) ([]document.ScoredChunk, error)
}

// Store is a Weaviate-backed store.
type Store struct {
	client    Client
	className string
}

// New returns a Weaviate adapter.
func New(client Client, className string) *Store { return &Store{client: client, className: className} }

// Upsert writes vectors.
func (s *Store) Upsert(ctx context.Context, chunks []document.Chunk, vectors [][]float64) error {
	if s.client == nil {
		return fmt.Errorf("weaviate client is nil")
	}
	return s.client.Upsert(ctx, s.className, chunks, vectors)
}

// Query executes a nearest-neighbor search.
func (s *Store) Query(ctx context.Context, vector []float64, opts store.QueryOptions) ([]document.ScoredChunk, error) {
	if opts.TopK <= 0 {
		opts.TopK = 5
	}
	return s.client.NearVector(ctx, s.className, vector, opts.TopK, opts.MetadataFilters)
}

var _ store.VectorStore = (*Store)(nil)
