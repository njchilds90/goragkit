package store

import (
	"context"
	"strconv"
	"testing"

	"github.com/njchilds90/goragkit/document"
)

func BenchmarkMemoryQuery(b *testing.B) {
	m := NewMemory()
	chunks := make([]document.Chunk, 5000)
	vecs := make([][]float64, 5000)
	for i := 0; i < 5000; i++ {
		chunks[i] = document.Chunk{ID: strconv.Itoa(i), Text: "chunk"}
		vecs[i] = []float64{float64(i%7) + 1, float64(i%11) + 1, float64(i%13) + 1}
	}
	if err := m.Upsert(context.Background(), chunks, vecs); err != nil {
		b.Fatal(err)
	}
	query := []float64{1, 2, 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Query(context.Background(), query, QueryOptions{TopK: 10})
	}
}
