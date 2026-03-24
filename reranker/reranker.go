// Package reranker provides post-retrieval scoring strategies.
package reranker

import (
	"sort"
	"strings"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/retrieval"
)

// Reranker rescored retrieval results given the query.
type Reranker interface {
	Rerank(query string, results []document.ScoredChunk) []document.ScoredChunk
}

// KeywordBoost boosts chunks containing query terms.
type KeywordBoost struct{ boost float64 }

// NewKeywordBoost returns a KeywordBoost reranker.
func NewKeywordBoost(boost float64) *KeywordBoost { return &KeywordBoost{boost: boost} }

// Rerank implements Reranker.
func (k *KeywordBoost) Rerank(query string, results []document.ScoredChunk) []document.ScoredChunk {
	terms := strings.Fields(strings.ToLower(query))
	out := append([]document.ScoredChunk(nil), results...)
	for i, r := range out {
		lower := strings.ToLower(r.Chunk.Text)
		for _, term := range terms {
			if strings.Contains(lower, term) {
				out[i].Score += k.boost
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

// BM25 reranks based on lexical relevance while preserving vector score signal.
type BM25 struct{ weight float64 }

// NewBM25 returns BM25 reranker.
func NewBM25(weight float64) *BM25 {
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}
	return &BM25{weight: weight}
}

// Rerank applies weighted BM25 score fusion.
func (b *BM25) Rerank(query string, results []document.ScoredChunk) []document.ScoredChunk {
	chunks := make([]document.Chunk, len(results))
	for i, r := range results {
		chunks[i] = r.Chunk
	}
	bm := retrieval.BM25(query, chunks)
	out := append([]document.ScoredChunk(nil), results...)
	for i := range out {
		out[i].Score = (1-b.weight)*out[i].Score + b.weight*bm[out[i].Chunk.ID]
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}
