// Package document defines the core types used across goragkit.
package document

// Document represents a source text with optional metadata.
type Document struct {
	ID       string
	Text     string
	Metadata map[string]string
}

// Chunk is a piece of a Document produced by a Chunker.
type Chunk struct {
	ID         string
	DocumentID string
	Text       string
	Index      int
	Metadata   map[string]string
}

// ScoredChunk pairs a Chunk with a retrieval score.
type ScoredChunk struct {
	Chunk Chunk
	Score float64
}