// Package pipeline composes chunking, embedding and retrieval into a high-level workflow.
package pipeline

import (
	"context"
	"fmt"

	"github.com/njchilds90/goragkit/chunker"
	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/embedder"
	"github.com/njchilds90/goragkit/retrieval"
	"github.com/njchilds90/goragkit/store"
	"github.com/njchilds90/goragkit/telemetry"
)

// Pipeline orchestrates embedding and retrieval.
type Pipeline struct {
	chunker   chunker.Chunker
	embedder  embedder.Embedder
	store     store.VectorStore
	retriever *retrieval.Retriever
	tracer    telemetry.Tracer
}

// Config configures a new Pipeline.
type Config struct {
	Chunker  chunker.Chunker
	Embedder embedder.Embedder
	Store    store.VectorStore
	Tracer   telemetry.Tracer
}

// New returns a Pipeline with explicit components.
func New(cfg Config) *Pipeline {
	tracer := cfg.Tracer
	if tracer == nil {
		tracer = telemetry.NewNoop()
	}
	r := retrieval.New(cfg.Embedder, cfg.Store)
	return &Pipeline{chunker: cfg.Chunker, embedder: cfg.Embedder, store: cfg.Store, retriever: r, tracer: tracer}
}

// NewRAGPipeline returns a batteries-included pipeline.
func NewRAGPipeline(emb embedder.Embedder, vs store.VectorStore) *Pipeline {
	return New(Config{
		Chunker:  chunker.NewSliding(700, 120),
		Embedder: embedder.NewCached(emb, 4096),
		Store:    vs,
	})
}

// IndexDocuments chunks, embeds, and stores documents.
func (p *Pipeline) IndexDocuments(ctx context.Context, docs []document.Document) error {
	ctx, span := p.tracer.Start(ctx, "pipeline.index_documents")
	defer span.End()
	chunks := make([]document.Chunk, 0, len(docs)*2)
	for _, d := range docs {
		for _, c := range p.chunker.Chunk(d.Text) {
			c.DocumentID = d.ID
			if c.Metadata == nil {
				c.Metadata = map[string]string{}
			}
			for k, v := range d.Metadata {
				c.Metadata[k] = v
			}
			chunks = append(chunks, c)
		}
	}
	return p.Index(ctx, chunks)
}

// Index embeds each chunk and upserts it into the vector store.
func (p *Pipeline) Index(ctx context.Context, chunks []document.Chunk) error {
	ctx, span := p.tracer.Start(ctx, "pipeline.index")
	defer span.End()
	if len(chunks) == 0 {
		return nil
	}
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	vecs, err := p.embedder.Embed(ctx, texts)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("embed chunks: %w", err)
	}
	if len(vecs) != len(chunks) {
		return fmt.Errorf("unexpected number of vectors: got %d want %d", len(vecs), len(chunks))
	}
	if err := p.store.Upsert(ctx, chunks, vecs); err != nil {
		span.RecordError(err)
		return fmt.Errorf("upsert vectors: %w", err)
	}
	return nil
}

// Query retrieves chunks for a query.
func (p *Pipeline) Query(ctx context.Context, query string, opts retrieval.QueryOptions) ([]document.ScoredChunk, error) {
	ctx, span := p.tracer.Start(ctx, "pipeline.query")
	defer span.End()
	results, err := p.retriever.Query(ctx, query, opts)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("query retrieval: %w", err)
	}
	return results, nil
}
