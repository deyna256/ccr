package correlation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/task-planner/server/internal/correlation"
)

func TestMiddleware_generatesIDWhenAbsent(t *testing.T) {
	var gotID string
	h := correlation.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotID = correlation.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if gotID == "" {
		t.Fatal("expected a generated ID, got empty string")
	}
	if got := rr.Header().Get(correlation.Header); got != gotID {
		t.Fatalf("response header %q = %q, want %q", correlation.Header, got, gotID)
	}
}

func TestMiddleware_propagatesIncomingID(t *testing.T) {
	const wantID = "incoming-id-123"
	var gotID string
	h := correlation.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotID = correlation.FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(correlation.Header, wantID)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if gotID != wantID {
		t.Fatalf("FromContext = %q, want %q", gotID, wantID)
	}
}

func TestTransport_injectsIDFromContext(t *testing.T) {
	const wantID = "outgoing-id-456"
	var gotHeader string

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(correlation.Header)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &correlation.Transport{Base: http.DefaultTransport},
	}

	ctx := correlation.WithContext(context.Background(), wantID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if gotHeader != wantID {
		t.Fatalf("server got %q = %q, want %q", correlation.Header, gotHeader, wantID)
	}
}

func TestExtractor_withID(t *testing.T) {
	ctx := correlation.WithContext(context.Background(), "abc-123")
	attrs := correlation.Extractor()(ctx)
	if len(attrs) != 1 {
		t.Fatalf("len(attrs) = %d, want 1", len(attrs))
	}
	if attrs[0].Key != "request_id" || attrs[0].Value.String() != "abc-123" {
		t.Errorf("attr = %v, want request_id=abc-123", attrs[0])
	}
}

func TestExtractor_withoutID(t *testing.T) {
	attrs := correlation.Extractor()(context.Background())
	if len(attrs) != 0 {
		t.Errorf("len(attrs) = %d, want 0", len(attrs))
	}
}

func TestTransport_skipsWhenContextEmpty(t *testing.T) {
	var gotHeader string

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(correlation.Header)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &correlation.Transport{Base: http.DefaultTransport},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if gotHeader != "" {
		t.Fatalf("expected no header, got %q", gotHeader)
	}
}
