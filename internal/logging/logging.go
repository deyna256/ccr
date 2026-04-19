package logging

import (
	"context"
	"log/slog"
	"os"
)

type ContextExtractor func(context.Context) []slog.Attr

func New(service string, level slog.Level, extractors ...ContextExtractor) *slog.Logger {
	inner := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // Promtail records ingestion time
			}
			return a
		},
	})
	return slog.New(&contextHandler{inner: inner, extractors: extractors}).With("service", service)
}

type contextHandler struct {
	inner      slog.Handler
	extractors []ContextExtractor
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, ext := range h.extractors {
		r.AddAttrs(ext(ctx)...)
	}
	return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{inner: h.inner.WithAttrs(attrs), extractors: h.extractors}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{inner: h.inner.WithGroup(name), extractors: h.extractors}
}
