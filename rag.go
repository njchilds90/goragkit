package goragkit

import (
	"github.com/njchilds90/goragkit/embedder"
	"github.com/njchilds90/goragkit/pipeline"
	"github.com/njchilds90/goragkit/store"
)

// NewRAGPipeline returns the recommended default pipeline for most applications.
func NewRAGPipeline(emb embedder.Embedder, vs store.VectorStore) *pipeline.Pipeline {
	return pipeline.NewRAGPipeline(emb, vs)
}
