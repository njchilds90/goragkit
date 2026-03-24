package store

import (
	"context"
	"testing"

	"github.com/njchilds90/goragkit/document"
)

func TestMemoryQueryFilters(t *testing.T) {
	m := NewMemory()
	chunks := []document.Chunk{{ID: "1", Text: "alpha", Metadata: map[string]string{"tenant": "a"}}, {ID: "2", Text: "beta", Metadata: map[string]string{"tenant": "b"}}}
	vecs := [][]float64{{1, 0}, {0, 1}}
	if err := m.Upsert(context.Background(), chunks, vecs); err != nil {
		t.Fatal(err)
	}
	out, err := m.Query(context.Background(), []float64{1, 0}, QueryOptions{TopK: 5, MetadataFilters: map[string]string{"tenant": "a"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Chunk.ID != "1" {
		t.Fatalf("unexpected result %#v", out)
	}
}
