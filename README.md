# goragkit

> Production-ready Retrieval-Augmented Generation (RAG) toolkit for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/goragkit.svg)](https://pkg.go.dev/github.com/njchilds90/goragkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/goragkit)](https://goreportcard.com/report/github.com/njchilds90/goragkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![](https://img.shields.io/badge/version-v0.1.0-blue)

Go has no cohesive RAG library. Python has LangChain, LlamaIndex, Haystack. `goragkit` fills that gap with idiomatic, production-ready primitives for building RAG pipelines in Go — suitable for both developers and AI agents calling Go services.

## Features

- **Chunking** — fixed-size, sentence-aware, and sliding-window strategies
- **Embeddings** — provider-agnostic interface with OpenAI and Ollama adapters
- **Vector Store** — in-memory store + interface for Pinecone, Weaviate, pgvector adapters
- **Retrieval Pipeline** — compose chunker → embedder → store → retriever in a few lines
- **Reranking** — cosine similarity reranker included, BM25 planned
- **Streaming** — `context.Context` throughout, safe for concurrent agents

## Install
```bash
go get github.com/njchilds90/goragkit
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"

    "github.com/njchilds90/goragkit/chunker"
    "github.com/njchilds90/goragkit/embedder"
    "github.com/njchilds90/goragkit/pipeline"
    "github.com/njchilds90/goragkit/store"
)

func main() {
    ctx := context.Background()

    // 1. Chunk your document
    chunks := chunker.NewFixed(512, 64).Chunk("Your long document text here...")

    // 2. Embed with OpenAI
    emb := embedder.NewOpenAI("YOUR_API_KEY", "text-embedding-3-small")

    // 3. Store in memory
    vs := store.NewMemory()

    // 4. Build and run pipeline
    p := pipeline.New(emb, vs)
    if err := p.Index(ctx, chunks); err != nil {
        panic(err)
    }

    results, err := p.Query(ctx, "What does the document say about X?", 5)
    if err != nil {
        panic(err)
    }
    for _, r := range results {
        fmt.Println(r.Score, r.Chunk.Text)
    }
}
```

## Package Overview

| Package | Description |
|---|---|
| `chunker` | Text splitting strategies |
| `embedder` | Embedding provider interface + adapters |
| `store` | Vector store interface + in-memory impl |
| `pipeline` | High-level index/query orchestration |
| `reranker` | Post-retrieval scoring |
| `document` | Document and Chunk types |

## Roadmap

- [ ] pgvector adapter
- [ ] Weaviate adapter
- [ ] Pinecone adapter
- [ ] Anthropic embeddings adapter
- [ ] BM25 reranker
- [ ] Metadata filtering
- [ ] CLI tool for indexing local files

## Contributing

PRs welcome. Please open an issue first for large changes.

## License

MIT © njchilds90