// Package otel provides a lightweight adapter shape compatible with OpenTelemetry tracers.
package otel

import (
	"context"

	"github.com/njchilds90/goragkit/telemetry"
)

// OTelSpan models the methods used from an OpenTelemetry span.
type OTelSpan interface {
	End()
	RecordError(error)
}

// OTelTracer models the methods used from an OpenTelemetry tracer.
type OTelTracer interface {
	Start(context.Context, string) (context.Context, OTelSpan)
}

// Tracer wraps an OpenTelemetry-compatible tracer.
type Tracer struct{ inner OTelTracer }

// New returns an OTel-compatible tracer wrapper.
func New(inner OTelTracer) *Tracer { return &Tracer{inner: inner} }

// Start starts a traced span.
func (t *Tracer) Start(ctx context.Context, name string) (context.Context, telemetry.Span) {
	ctx, span := t.inner.Start(ctx, name)
	return ctx, span
}
