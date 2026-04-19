package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func newTestLogger(buf *bytes.Buffer, extractors ...ContextExtractor) *slog.Logger {
	inner := slog.NewJSONHandler(buf, nil)
	return slog.New(&contextHandler{inner: inner, extractors: extractors})
}

func parseRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}
	return m
}

func TestNew_notNil(t *testing.T) {
	if New("ccr", slog.LevelInfo) == nil {
		t.Fatal("New returned nil")
	}
}

func TestContextHandler_noExtractors(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.InfoContext(context.Background(), "test")

	record := parseRecord(t, &buf)
	if _, ok := record["request_id"]; ok {
		t.Error("unexpected request_id field when no extractor is registered")
	}
}

func TestContextHandler_injectsAttrs(t *testing.T) {
	var buf bytes.Buffer
	ext := func(_ context.Context) []slog.Attr {
		return []slog.Attr{slog.String("request_id", "abc-123")}
	}
	logger := newTestLogger(&buf, ext)

	logger.InfoContext(context.Background(), "test")

	record := parseRecord(t, &buf)
	if record["request_id"] != "abc-123" {
		t.Errorf("request_id = %v, want %q", record["request_id"], "abc-123")
	}
}

func TestContextHandler_multipleExtractors(t *testing.T) {
	var buf bytes.Buffer
	ext1 := func(_ context.Context) []slog.Attr {
		return []slog.Attr{slog.String("request_id", "abc-123")}
	}
	ext2 := func(_ context.Context) []slog.Attr {
		return []slog.Attr{slog.String("user_id", "user-456")}
	}
	logger := newTestLogger(&buf, ext1, ext2)

	logger.InfoContext(context.Background(), "test")

	record := parseRecord(t, &buf)
	if record["request_id"] != "abc-123" {
		t.Errorf("request_id = %v, want %q", record["request_id"], "abc-123")
	}
	if record["user_id"] != "user-456" {
		t.Errorf("user_id = %v, want %q", record["user_id"], "user-456")
	}
}

func TestContextHandler_extractorReturnsNil(t *testing.T) {
	var buf bytes.Buffer
	ext := func(_ context.Context) []slog.Attr { return nil }
	logger := newTestLogger(&buf, ext)

	logger.InfoContext(context.Background(), "test")

	parseRecord(t, &buf)
}

func TestContextHandler_WithAttrs_preservesExtractors(t *testing.T) {
	var buf bytes.Buffer
	ext := func(_ context.Context) []slog.Attr {
		return []slog.Attr{slog.String("request_id", "abc-123")}
	}
	logger := newTestLogger(&buf, ext).With("service", "test")

	logger.InfoContext(context.Background(), "test")

	record := parseRecord(t, &buf)
	if record["request_id"] != "abc-123" {
		t.Errorf("request_id = %v, want %q after WithAttrs", record["request_id"], "abc-123")
	}
}
