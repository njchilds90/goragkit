// Package chunker provides strategies for splitting text into Chunks.
package chunker

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/njchilds90/goragkit/document"
)

// Chunker splits text into chunks.
type Chunker interface {
	Chunk(text string) []document.Chunk
}

// Fixed splits text into fixed-size rune windows with overlap.
type Fixed struct {
	size, overlap int
}

// NewFixed returns a Fixed chunker.
func NewFixed(size, overlap int) *Fixed {
	if size <= 0 {
		size = 1
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size - 1
	}
	return &Fixed{size: size, overlap: overlap}
}

// Chunk implements Chunker.
func (f *Fixed) Chunk(text string) []document.Chunk {
	var chunks []document.Chunk
	runes := []rune(text)
	step := f.size - f.overlap
	for i, idx := 0, 0; idx < len(runes); i, idx = i+1, idx+step {
		end := idx + f.size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, document.Chunk{ID: newID(), Text: string(runes[idx:end]), Index: i})
		if end == len(runes) {
			break
		}
	}
	return chunks
}

// Sentence splits text on sentence boundaries.
type Sentence struct{ maxSize int }

// NewSentence returns a Sentence chunker.
func NewSentence(maxSize int) *Sentence {
	if maxSize <= 0 {
		maxSize = 512
	}
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
		chunks = append(chunks, document.Chunk{ID: newID(), Text: strings.TrimSpace(buf.String()), Index: idx})
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

// Sliding splits text by tokens with overlap, preserving word boundaries.
type Sliding struct {
	window, overlap int
}

// NewSliding returns a word-aware sliding-window chunker.
func NewSliding(window, overlap int) *Sliding {
	if window <= 0 {
		window = 256
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= window {
		overlap = window - 1
	}
	return &Sliding{window: window, overlap: overlap}
}

// Chunk implements Chunker.
func (s *Sliding) Chunk(text string) []document.Chunk {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	step := s.window - s.overlap
	out := make([]document.Chunk, 0, len(words)/step+1)
	for i, idx := 0, 0; idx < len(words); i, idx = i+1, idx+step {
		end := idx + s.window
		if end > len(words) {
			end = len(words)
		}
		out = append(out, document.Chunk{ID: newID(), Text: strings.Join(words[idx:end], " "), Index: i})
		if end == len(words) {
			break
		}
	}
	return out
}

func splitOnAny(text string, seps []string) []string {
	result := []string{text}
	for _, sep := range seps {
		var next []string
		for _, part := range result {
			next = append(next, strings.SplitAfter(part, sep)...)
		}
		result = next
	}
	return result
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
