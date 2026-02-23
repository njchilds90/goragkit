// Package chunker provides strategies for splitting text into Chunks.
package chunker

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/njchilds90/goragkit/document"
)

// Chunker splits text into a slice of Chunks.
type Chunker interface {
	Chunk(text string) []document.Chunk
}

// Fixed splits text into fixed-size token windows with optional overlap.
type Fixed struct {
	size    int
	overlap int
}

// NewFixed returns a Fixed chunker.
// size is the approximate character window; overlap is the number of
// characters re-included from the previous chunk.
func NewFixed(size, overlap int) *Fixed {
	return &Fixed{size: size, overlap: overlap}
}

// Chunk implements Chunker.
func (f *Fixed) Chunk(text string) []document.Chunk {
	var chunks []document.Chunk
	runes := []rune(text)
	step := f.size - f.overlap
	if step <= 0 {
		step = f.size
	}
	for i, idx := 0, 0; idx < len(runes); i, idx = i+1, idx+step {
		end := idx + f.size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, document.Chunk{
			ID:    newID(),
			Text:  string(runes[idx:end]),
			Index: i,
		})
		if end == len(runes) {
			break
		}
	}
	return chunks
}

// Sentence splits text on sentence boundaries (". ", "? ", "! ").
type Sentence struct {
	maxSize int
}

// NewSentence returns a Sentence chunker. maxSize controls the soft
// maximum character length before a new chunk is started.
func NewSentence(maxSize int) *Sentence {
	return &Sentence{maxSize: maxSize}
}

// Chunk implements Chunker.
func (s *Sentence) Chunk(text string) []document.Chunk {
	separators := []string{". ", "? ", "! "}
	sentences := splitOnAny(text, separators)
	var chunks []document.Chunk
	var buf strings.Builder
	idx := 0
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		chunks = append(chunks, document.Chunk{
			ID:    newID(),
			Text:  strings.TrimSpace(buf.String()),
			Index: idx,
		})
		idx++
		buf.Reset()
	}
	for _, sent := range sentences {
		if buf.Len()+len(sent) > s.maxSize && buf.Len() > 0 {
			flush()
		}
		buf.WriteString(sent)
	}
	flush()
	return chunks
}

func splitOnAny(text string, seps []string) []string {
	result := []string{text}
	for _, sep := range seps {
		var next []string
		for _, part := range result {
			split := strings.SplitAfter(part, sep)
			next = append(next, split...)
		}
		result = next
	}
	return result
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
