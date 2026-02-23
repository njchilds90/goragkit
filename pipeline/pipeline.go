// Package pipeline composes an Embedder and VectorStore into a high-level
// index/query workflow.
package pipeline

import (
	"context"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/embedder"
	"github.com/njchilds90/goragkit/store"
)

// Pipeline orchestrates embedding and retrieval.
type Pipeline struct {
	embedder embedder.Embedder
	store    store.VectorStore
}

// New returns a Pipeline backed by the given Embedder and VectorStore.
func New(emb embedder.Embedder, vs store.VectorStore) *Pipeline {
	return &Pipeline{embedder: emb, store: vs}
}

// Index embeds each chunk and upserts it into the vector store.
func (p *Pipeline) Index(ctx context.Context, chunks []document.Chunk) error {
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	vecs, err := p.embedder.Embed(ctx, texts)
	if err != nil {
		return err
	}
	return p.store.Upsert(ctx, chunks, vecs)
}

// Query embeds the query text and retrieves the topK most relevant chunks.
func (p *Pipeline) Query(ctx context.Context, query string, topK int) ([]document.ScoredChunk, error) {
	vecs, err := p.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	return p.store.Query(ctx, vecs[0], topK)
}
