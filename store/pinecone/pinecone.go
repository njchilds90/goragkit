// Package pinecone provides a Pinecone VectorStore adapter.
package pinecone

import (
	"context"
	"fmt"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/store"
)

// Client is the HTTP client contract used by this adapter.
type Client interface {
	Upsert(ctx context.Context, namespace string, chunks []document.Chunk, vectors [][]float64) error
	Query(ctx context.Context, namespace string, vector []float64, topK int, filters map[string]string) ([]document.ScoredChunk, error)
}

// Store is a Pinecone-backed VectorStore.
type Store struct {
	client    Client
	namespace string
}

// New returns a Pinecone adapter store.
func New(client Client, namespace string) *Store { return &Store{client: client, namespace: namespace} }

// Upsert writes vectors.
func (s *Store) Upsert(ctx context.Context, chunks []document.Chunk, vectors [][]float64) error {
	if s.client == nil {
		return fmt.Errorf("pinecone client is nil")
	}
	return s.client.Upsert(ctx, s.namespace, chunks, vectors)
}

// Query retrieves vectors.
func (s *Store) Query(ctx context.Context, vector []float64, opts store.QueryOptions) ([]document.ScoredChunk, error) {
	if opts.TopK <= 0 {
		opts.TopK = 5
	}
	return s.client.Query(ctx, s.namespace, vector, opts.TopK, opts.MetadataFilters)
}

var _ store.VectorStore = (*Store)(nil)
