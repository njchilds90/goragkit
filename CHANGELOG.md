# Changelog

## v0.2.0 - 2026-03-24

### Added
- High-level `NewRAGPipeline` quickstart API.
- CLI binary `cmd/goragkit` with `index`, `query`, and `serve` commands.
- Hybrid retrieval package with BM25 fusion and metadata filters.
- `loader` package for recursive filesystem ingestion.
- `cache` package with generic LRU implementation.
- OpenTelemetry bridge package (`telemetry/otel`).
- Vector store adapter scaffolds for Pinecone, Weaviate, and pgvector.
- Cohere embedder adapter.
- Structured `rerrors` package for agent-friendly error handling.
- CI workflow with lint, tests, race, and coverage.

### Changed
- Pipeline API moved to config-driven constructor and richer query options.
- Memory vector store now supports metadata filtering and query options.
- README rewritten with architecture, usage, and contribution guidance.
- Go version requirement updated to 1.22.

### Quality
- Added unit tests and benchmark coverage for cache, retrieval, pipeline, and store.
