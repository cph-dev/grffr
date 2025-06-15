package logging

import (
	"context"
	"log/slog"
)

// Code in this file is inspired by the article written by Ayooluwa Isaiah:
// https://betterstack.com/community/guides/logging/logging-in-go/
// License unknown.

type ctxKey string

const (
	SlogFields ctxKey = "slog_fields"
)

type ContextHandler struct {
	slog.Handler
}

// Handle adds contextual attributes to the Record before calling the underlying
// handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(SlogFields).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}

	// Call the underlying handler
	return h.Handler.Handle(ctx, r)
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithAttrs returns a new [ContextHandler] whose attributes consists
// of h's attributes followed by attrs.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}

// AppendCtx adds an slog attribute to the provided context so that it will be
// included in any Record created with such context.
func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(SlogFields).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, SlogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)

	return context.WithValue(parent, SlogFields, v)
}
