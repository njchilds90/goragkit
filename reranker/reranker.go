// Package reranker provides post-retrieval scoring strategies.
package reranker

import (
	"strings"

	"github.com/njchilds90/goragkit/document"
)

// Reranker rescores a slice of ScoredChunks given the original query.
type Reranker interface {
	Rerank(query string, results []document.ScoredChunk) []document.ScoredChunk
}

// KeywordBoost is a simple reranker that boosts chunks containing query terms.
type KeywordBoost struct {
	boost float64
}

// NewKeywordBoost returns a KeywordBoost reranker.
// boost is added to the score for each query term found in the chunk.
func NewKeywordBoost(boost float64) *KeywordBoost {
	return &KeywordBoost{boost: boost}
}

// Rerank implements Reranker.
func (k *KeywordBoost) Rerank(query string, results []document.ScoredChunk) []document.ScoredChunk {
	terms := strings.Fields(strings.ToLower(query))
	out := make([]document.ScoredChunk, len(results))
	copy(out, results)
	for i, r := range out {
		lower := strings.ToLower(r.Chunk.Text)
		for _, term := range terms {
			if strings.Contains(lower, term) {
				out[i].Score += k.boost
			}
		}
	}
	// sort descending
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Score > out[i].Score {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}