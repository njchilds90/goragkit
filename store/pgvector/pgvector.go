// Package pgvector provides a pgvector-backed VectorStore adapter.
package pgvector

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/store"
)

// Store is a PostgreSQL/pgvector store.
type Store struct {
	db    *sql.DB
	table string
}

// New returns a pgvector adapter.
func New(db *sql.DB, table string) *Store { return &Store{db: db, table: table} }

// Upsert inserts rows with vector and metadata.
func (s *Store) Upsert(ctx context.Context, chunks []document.Chunk, vectors [][]float64) error {
	if len(chunks) != len(vectors) {
		return fmt.Errorf("chunks and vectors mismatch")
	}
	for i := range chunks {
		meta, _ := json.Marshal(chunks[i].Metadata)
		q := fmt.Sprintf(`INSERT INTO %s (id, document_id, idx, text, metadata, embedding) VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (id) DO UPDATE SET text = EXCLUDED.text, metadata = EXCLUDED.metadata, embedding = EXCLUDED.embedding`, s.table)
		if _, err := s.db.ExecContext(ctx, q, chunks[i].ID, chunks[i].DocumentID, chunks[i].Index, chunks[i].Text, string(meta), vectorLiteral(vectors[i])); err != nil {
			return err
		}
	}
	return nil
}

// Query uses cosine distance in PostgreSQL.
func (s *Store) Query(ctx context.Context, vector []float64, opts store.QueryOptions) ([]document.ScoredChunk, error) {
	if opts.TopK <= 0 {
		opts.TopK = 5
	}
	query := fmt.Sprintf(`SELECT id, document_id, idx, text, metadata, 1 - (embedding <=> $1::vector) AS score FROM %s ORDER BY embedding <=> $1::vector LIMIT $2`, s.table)
	rows, err := s.db.QueryContext(ctx, query, vectorLiteral(vector), opts.TopK)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]document.ScoredChunk, 0, opts.TopK)
	for rows.Next() {
		var c document.Chunk
		var meta string
		var sscore float64
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.Index, &c.Text, &meta, &sscore); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(meta), &c.Metadata)
		out = append(out, document.ScoredChunk{Chunk: c, Score: sscore})
	}
	return out, rows.Err()
}

func vectorLiteral(v []float64) string {
	parts := make([]string, len(v))
	for i := range v {
		parts[i] = fmt.Sprintf("%f", v[i])
	}
	return "[" + strings.Join(parts, ",") + "]"
}

var _ store.VectorStore = (*Store)(nil)
