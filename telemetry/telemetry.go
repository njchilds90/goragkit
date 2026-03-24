// Package telemetry provides tracing abstractions.
package telemetry

import "context"

// Span abstracts a tracing span.
type Span interface {
	End()
	RecordError(error)
}

// Tracer abstracts trace instrumentation.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

type noopTracer struct{}
type noopSpan struct{}

// NewNoop returns a tracer that does nothing.
func NewNoop() Tracer { return noopTracer{} }

func (noopTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, noopSpan{}
}
func (noopSpan) End()              {}
func (noopSpan) RecordError(error) {}
