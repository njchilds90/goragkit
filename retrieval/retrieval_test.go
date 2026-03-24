package retrieval

import (
	"testing"

	"github.com/njchilds90/goragkit/document"
)

func TestBM25(t *testing.T) {
	chunks := []document.Chunk{{ID: "1", Text: "go rag toolkit"}, {ID: "2", Text: "python chains"}}
	s := BM25("go toolkit", chunks)
	if s["1"] <= s["2"] {
		t.Fatalf("expected chunk 1 to score higher: %#v", s)
	}
}
