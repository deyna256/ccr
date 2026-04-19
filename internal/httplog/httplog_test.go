package httplog_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/task-planner/server/internal/httplog"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(h).With("service", "test-svc")
}

func parseLogs(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var records []map[string]any
	dec := json.NewDecoder(buf)
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("invalid JSON log output: %v\nraw: %s", err, buf.String())
		}
		records = append(records, m)
	}
	return records
}

func TestMiddleware_logsArrival(t *testing.T) {
	var buf bytes.Buffer
	h := httplog.Middleware(newTestLogger(&buf))(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/chat", nil))

	records := parseLogs(t, &buf)
	if len(records) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(records))
	}
	arrival := records[0]
	if arrival["msg"] != "request" {
		t.Errorf("msg = %v, want %q", arrival["msg"], "request")
	}
	for _, field := range []string{"method", "path"} {
		if _, ok := arrival[field]; !ok {
			t.Errorf("arrival log missing field %q", field)
		}
	}
	if _, ok := arrival["status"]; ok {
		t.Error("arrival log must not contain status")
	}
}

func TestMiddleware_normalRequest(t *testing.T) {
	var buf bytes.Buffer
	h := httplog.Middleware(newTestLogger(&buf))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/chat", nil))

	records := parseLogs(t, &buf)
	if len(records) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(records))
	}
	completion := records[1]
	if completion["msg"] != "response" {
		t.Errorf("msg = %v, want %q", completion["msg"], "response")
	}
	for _, field := range []string{"method", "path", "status", "duration"} {
		if _, ok := completion[field]; !ok {
			t.Errorf("completion log missing field %q", field)
		}
	}
	if _, ok := completion["cached"]; ok {
		t.Error("cached must be absent for normal requests")
	}
	if completion["status"] != float64(http.StatusCreated) {
		t.Errorf("status = %v, want %d", completion["status"], http.StatusCreated)
	}
}

func TestMiddleware_cachedRequest(t *testing.T) {
	var buf bytes.Buffer
	h := httplog.Middleware(newTestLogger(&buf))(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		httplog.MarkCached(r.Context())
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/article/1", nil))

	records := parseLogs(t, &buf)
	if len(records) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(records))
	}
	completion := records[1]
	if completion["cached"] != true {
		t.Errorf("cached = %v, want true", completion["cached"])
	}
	if _, ok := completion["duration"]; ok {
		t.Error("duration must be absent for cached requests")
	}
}

func TestMiddleware_defaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	h := httplog.Middleware(newTestLogger(&buf))(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	records := parseLogs(t, &buf)
	if len(records) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(records))
	}
	if records[1]["status"] != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", records[1]["status"])
	}
}
