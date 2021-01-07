package trace

import (
	"context"
	"time"
)

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, *Span)
}

type SpanType int

const (
	SpanTypeRequestInbound SpanType = iota
	SpanTypeRequestOutbound
)

type Span struct {
	// Id of the trace
	Trace string
	// name of the span
	Name string
	// id of the span
	Id string
	// parent span id
	Parent string
	// Start time
	Started time.Time
	// Duration in nano seconds
	Duration time.Duration
	// associated data
	Metadata map[string]string
	// Type
	Type SpanType
}
