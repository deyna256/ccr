package correlation

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/rs/xid"
)

const Header = "X-Request-ID"

type contextKey struct{}

func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(Header)
		if id == "" {
			id = xid.New().String()
		}
		w.Header().Set(Header, id)
		next.ServeHTTP(w, r.WithContext(WithContext(r.Context(), id)))
	})
}

func Extractor() func(context.Context) []slog.Attr {
	return func(ctx context.Context) []slog.Attr {
		if id := FromContext(ctx); id != "" {
			return []slog.Attr{slog.String("request_id", id)}
		}
		return nil
	}
}

type Transport struct {
	Base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	if id := FromContext(req.Context()); id != "" {
		req = req.Clone(req.Context())
		req.Header.Set(Header, id)
	}
	return base.RoundTrip(req)
}
