// Package retrieval implements advanced retrieval primitives.
package retrieval

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/embedder"
	"github.com/njchilds90/goragkit/store"
)

// Retriever performs vector and optional lexical retrieval.
type Retriever struct {
	embedder      embedder.Embedder
	store         store.VectorStore
	lexicalWeight float64
	chunkText     func(ctx context.Context, filters map[string]string) ([]document.Chunk, error)
}

// New returns a Retriever.
func New(emb embedder.Embedder, vs store.VectorStore) *Retriever {
	return &Retriever{embedder: emb, store: vs, lexicalWeight: 0.2}
}

// WithLexicalSource sets a source used for BM25 fusion.
func (r *Retriever) WithLexicalSource(fn func(ctx context.Context, filters map[string]string) ([]document.Chunk, error)) *Retriever {
	r.chunkText = fn
	return r
}

// WithLexicalWeight configures BM25 blend ratio [0,1].
func (r *Retriever) WithLexicalWeight(weight float64) *Retriever {
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}
	r.lexicalWeight = weight
	return r
}

// QueryOptions controls query behavior.
type QueryOptions struct {
	TopK            int
	MetadataFilters map[string]string
}

// Query executes retrieval.
func (r *Retriever) Query(ctx context.Context, query string, opts QueryOptions) ([]document.ScoredChunk, error) {
	vecs, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	results, err := r.store.Query(ctx, vecs[0], store.QueryOptions{TopK: opts.TopK, MetadataFilters: opts.MetadataFilters})
	if err != nil {
		return nil, err
	}
	if r.chunkText == nil || r.lexicalWeight == 0 {
		return results, nil
	}
	chunks, err := r.chunkText(ctx, opts.MetadataFilters)
	if err != nil {
		return nil, err
	}
	bm25 := BM25(query, chunks)
	scores := make(map[string]float64, len(results)+len(bm25))
	chunkByID := map[string]document.Chunk{}
	for _, c := range chunks {
		chunkByID[c.ID] = c
	}
	for _, res := range results {
		scores[res.Chunk.ID] += (1 - r.lexicalWeight) * res.Score
		chunkByID[res.Chunk.ID] = res.Chunk
	}
	for id, score := range bm25 {
		scores[id] += r.lexicalWeight * score
	}
	out := make([]document.ScoredChunk, 0, len(scores))
	for id, score := range scores {
		out = append(out, document.ScoredChunk{Chunk: chunkByID[id], Score: score})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if opts.TopK > 0 && len(out) > opts.TopK {
		out = out[:opts.TopK]
	}
	return out, nil
}

// BM25 returns chunk-id to bm25 score.
func BM25(query string, chunks []document.Chunk) map[string]float64 {
	qTerms := tokenize(query)
	if len(qTerms) == 0 || len(chunks) == 0 {
		return map[string]float64{}
	}
	k1, b := 1.5, 0.75
	df := map[string]float64{}
	tfByChunk := make(map[string]map[string]float64, len(chunks))
	avgdl := 0.0
	for _, ch := range chunks {
		tokens := tokenize(ch.Text)
		avgdl += float64(len(tokens))
		tf := map[string]float64{}
		seen := map[string]struct{}{}
		for _, t := range tokens {
			tf[t]++
			if _, ok := seen[t]; !ok {
				df[t]++
				seen[t] = struct{}{}
			}
		}
		tfByChunk[ch.ID] = tf
	}
	avgdl /= float64(len(chunks))
	scores := map[string]float64{}
	N := float64(len(chunks))
	for _, ch := range chunks {
		docLen := 0.0
		for _, n := range tfByChunk[ch.ID] {
			docLen += n
		}
		s := 0.0
		for _, term := range qTerms {
			tf := tfByChunk[ch.ID][term]
			if tf == 0 {
				continue
			}
			idf := math.Log(1 + (N-df[term]+0.5)/(df[term]+0.5))
			s += idf * ((tf * (k1 + 1)) / (tf + k1*(1-b+b*(docLen/avgdl))))
		}
		scores[ch.ID] = s
	}
	return scores
}

func tokenize(s string) []string {
	fields := strings.Fields(strings.ToLower(s))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, ".,!?:;()[]{}\"'")
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}
