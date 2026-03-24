package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/njchilds90/goragkit/chunker"
	"github.com/njchilds90/goragkit/document"
	"github.com/njchilds90/goragkit/loader"
	"github.com/njchilds90/goragkit/pipeline"
	"github.com/njchilds90/goragkit/retrieval"
	"github.com/njchilds90/goragkit/store"
)

type deterministicEmbedder struct{}

func (deterministicEmbedder) Embed(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i, t := range texts {
		h := sha256.Sum256([]byte(t))
		vec := make([]float64, 32)
		for j := range vec {
			vec[j] = float64(h[j])/255*2 - 1
		}
		norm := 0.0
		for _, v := range vec {
			norm += v * v
		}
		norm = math.Sqrt(norm)
		for j := range vec {
			vec[j] /= norm
		}
		out[i] = vec
	}
	return out, nil
}

type persisted struct {
	Chunks  []document.Chunk `json:"chunks"`
	Vectors [][]float64      `json:"vectors"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "index":
		err = runIndex(os.Args[2:])
	case "query":
		err = runQuery(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Println("goragkit <index|query|serve>")
}

func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	out := fs.String("out", ".goragkit/index.json", "output file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("index requires <path>")
	}
	path := fs.Arg(0)
	ldr := loader.NewDirectory(path)
	docs, err := ldr.Load(context.Background())
	if err != nil {
		return err
	}
	pl := pipeline.New(pipeline.Config{Chunker: chunker.NewSliding(200, 40), Embedder: deterministicEmbedder{}, Store: store.NewMemory()})
	chunks := make([]document.Chunk, 0)
	for _, d := range docs {
		for _, ch := range chunker.NewSliding(200, 40).Chunk(d.Text) {
			ch.DocumentID = d.ID
			ch.Metadata = d.Metadata
			chunks = append(chunks, ch)
		}
	}
	if err := pl.Index(context.Background(), chunks); err != nil {
		return err
	}
	vecs, _ := deterministicEmbedder{}.Embed(context.Background(), chunkTexts(chunks))
	blob, _ := json.MarshalIndent(persisted{Chunks: chunks, Vectors: vecs}, "", "  ")
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*out, blob, 0o644); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "ok", "documents": len(docs), "chunks": len(chunks), "index": *out})
}

func runQuery(args []string) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	index := fs.String("index", ".goragkit/index.json", "index path")
	q := fs.String("q", "", "query")
	topk := fs.Int("topk", 5, "top K")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *q == "" {
		return errors.New("-q is required")
	}
	data, err := os.ReadFile(*index)
	if err != nil {
		return err
	}
	var p persisted
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	m := store.NewMemory()
	if err := m.Upsert(context.Background(), p.Chunks, p.Vectors); err != nil {
		return err
	}
	pl := pipeline.NewRAGPipeline(deterministicEmbedder{}, m)
	res, err := pl.Query(context.Background(), *q, retrieval.QueryOptions{TopK: *topk})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"query": *q, "results": res})
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	index := fs.String("index", ".goragkit/index.json", "index path")
	addr := fs.String("addr", ":8080", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := os.ReadFile(*index)
	if err != nil {
		return err
	}
	var p persisted
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	m := store.NewMemory()
	if err := m.Upsert(context.Background(), p.Chunks, p.Vectors); err != nil {
		return err
	}
	pl := pipeline.NewRAGPipeline(deterministicEmbedder{}, m)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Query string `json:"query"`
			TopK  int    `json:"top_k"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		res, err := pl.Query(r.Context(), req.Query, retrieval.QueryOptions{TopK: req.TopK})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": res})
	})
	fmt.Println("listening on", *addr)
	return http.ListenAndServe(*addr, h)
}

func chunkTexts(chunks []document.Chunk) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.Text
	}
	return out
}
