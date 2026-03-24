package pipeline

import (
	"context"
	"testing"

	"github.com/njchilds90/goragkit/chunker"
	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/retrieval"
	"github.com/njchilds90/goragkit/store"
)

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(_ context.Context, texts []string) ([][]float64, error) {
	v := make([][]float64, len(texts))
	for i := range texts {
		v[i] = []float64{float64(len(texts[i])), 1}
	}
	return v, nil
}

func TestPipelineIndexQuery(t *testing.T) {
	p := New(Config{Chunker: chunker.NewFixed(10, 0), Embedder: fakeEmbedder{}, Store: store.NewMemory()})
	if err := p.IndexDocuments(context.Background(), []document.Document{{ID: "d1", Text: "go is great for services"}}); err != nil {
		t.Fatal(err)
	}
	out, err := p.Query(context.Background(), "go", retrieval.QueryOptions{TopK: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one result got %d", len(out))
	}
}
