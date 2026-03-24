// Package document defines the core types used across goragkit.
package document

// Document represents a source text with optional metadata.
type Document struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Chunk is a piece of a Document produced by a Chunker.
type Chunk struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id,omitempty"`
	Text       string            `json:"text"`
	Index      int               `json:"index"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ScoredChunk pairs a Chunk with a retrieval score.
type ScoredChunk struct {
	Chunk Chunk   `json:"chunk"`
	Score float64 `json:"score"`
}
