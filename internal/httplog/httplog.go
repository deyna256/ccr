package httplog

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type contextKey struct{}

type requestMeta struct {
	cached bool
}

// no-op when called outside Middleware context
func MarkCached(ctx context.Context) {
	if m, ok := ctx.Value(contextKey{}).(*requestMeta); ok {
		m.cached = true
	}
}

func Middleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			meta := &requestMeta{}
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			log.LogAttrs(r.Context(), slog.LevelInfo, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			next.ServeHTTP(rw, r.WithContext(
				context.WithValue(r.Context(), contextKey{}, meta),
			))

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
			}
			if meta.cached {
				attrs = append(attrs, slog.Bool("cached", true))
			} else {
				attrs = append(attrs, slog.String("duration", time.Since(start).String()))
			}

			log.LogAttrs(r.Context(), slog.LevelInfo, "response", attrs...)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
