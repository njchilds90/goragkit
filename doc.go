// Package goragkit provides production-ready RAG (Retrieval-Augmented Generation)
// primitives for Go.
//
// # Overview
//
// goragkit is organised into focused sub-packages:
//
//   - [github.com/njchilds90/goragkit/document] — core types (Document, Chunk, ScoredChunk)
//   - [github.com/njchilds90/goragkit/chunker]  — text splitting strategies
//   - [github.com/njchilds90/goragkit/embedder] — embedding provider interface + adapters
//   - [github.com/njchilds90/goragkit/store]    — vector store interface + in-memory impl
//   - [github.com/njchilds90/goragkit/pipeline] — high-level index/query orchestration
//   - [github.com/njchilds90/goragkit/reranker] — post-retrieval scoring
//
// # Quick start
//
//	p := pipeline.New(
//	    embedder.NewOpenAI(apiKey, "text-embedding-3-small"),
//	    store.NewMemory(),
//	)
//	p.Index(ctx, chunker.NewFixed(512, 64).Chunk(docText))
//	results, _ := p.Query(ctx, "your question", 5)
//
// See the README for a full example.
package goragkit