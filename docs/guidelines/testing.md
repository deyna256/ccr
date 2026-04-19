# Testing Guideline

## Philosophy

- Tests are the primary tool for verifying behaviour, not implementation details.
- Every package must be independently testable without real external services.
- No mocks for external infrastructure (HTTP servers, databases, brokers) — use real in-process replacements (`httptest`, in-memory structs). A mock tests how you *think* a service behaves; if the contract is misunderstood or the service changes, the mock keeps passing while production breaks.
- In-process replacements are not mocks: `httptest.NewServer` runs a real HTTP stack, `bytes.Buffer` is a real `io.Writer`. They test real behaviour without network calls.
- Test doubles for *your own* interfaces are fine. If a package defines a `Storage` interface on the consumer side, a test implementation that satisfies it is not a mock of an external service — it is a real implementation of your own contract.
- Test only observable outcomes: returned values, written output, HTTP responses, side effects visible through the public API.

---

## Package Layout

Use the external test package (`package foo_test`) for all tests.  
This enforces the public API boundary and prevents tests from reaching into unexported internals.

```
correlation/
├── correlation.go       // package correlation
└── correlation_test.go  // package correlation_test
```

Exception: when a test helper needs access to unexported types (e.g. building a test logger that reuses an internal handler), place it in the same package (`package foo`) and name the file `export_test.go`.

---

## Test Function Naming

```
Test<Subject>_<scenario>
```

`Subject` is the type or function under test. `scenario` describes the condition or input, not the expected result.

```go
// good
func TestMiddleware_generatesIDWhenAbsent(t *testing.T)
func TestMiddleware_propagatesIncomingID(t *testing.T)
func TestExtractor_withoutID(t *testing.T)
func TestContextHandler_WithAttrs_preservesExtractors(t *testing.T)

// bad
func TestMiddleware(t *testing.T)           // no scenario
func TestMiddleware_returnsID(t *testing.T) // describes result, not condition
```

---

## Structure

Each test follows the same three-phase structure without section comments:

```go
func TestTransport_injectsIDFromContext(t *testing.T) {
    // arrange
    const wantID = "outgoing-id-456"
    var gotHeader string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        gotHeader = r.Header.Get(correlation.Header)
    }))
    defer server.Close()
    client := &http.Client{Transport: &correlation.Transport{Base: http.DefaultTransport}}

    // act
    ctx := correlation.WithContext(context.Background(), wantID)
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
    _, err := client.Do(req)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }

    // assert
    if gotHeader != wantID {
        t.Fatalf("server got header %q = %q, want %q", correlation.Header, gotHeader, wantID)
    }
}
```

---

## Assertions

Use `t.Fatal` / `t.Fatalf` when the test cannot continue after a failure.  
Use `t.Error` / `t.Errorf` to accumulate multiple independent failures.

```go
// stop immediately — further checks are meaningless
if len(records) != 2 {
    t.Fatalf("expected 2 log lines, got %d", len(records))
}

// accumulate — both fields are independent
if _, ok := record["method"]; !ok {
    t.Errorf("missing field %q", "method")
}
if _, ok := record["path"]; !ok {
    t.Errorf("missing field %q", "path")
}
```

Do not use third-party assertion libraries (`testify`, `gomega`). Stdlib is sufficient.

---

## HTTP Testing

Use `net/http/httptest` — no real ports, no network.

**Handler under test:**

```go
h := httplog.Middleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusCreated)
}))

rr := httptest.NewRecorder()
h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/chat", nil))
```

**Integration test with a real server** (when testing an outgoing client):

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    gotHeader = r.Header.Get(correlation.Header)
}))
defer server.Close()

client := &http.Client{Transport: &correlation.Transport{Base: http.DefaultTransport}}
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
client.Do(req)
```

Use `httptest.NewServer` only when testing outgoing HTTP behaviour (transport, client middleware).  
For handler testing prefer `httptest.NewRecorder` + direct `ServeHTTP` call — it's faster and has no cleanup.

---

## Logging in Tests

When a test needs to inspect log output, build a logger backed by `bytes.Buffer` that mirrors the shape of `logging.New` but writes to the buffer:

```go
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
```

Parse log output as JSON — never match raw strings:

```go
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
```

---

## Test Helpers

Mark helpers with `t.Helper()` — error lines point to the call site, not inside the helper:

```go
func parseRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
    t.Helper()
    var m map[string]any
    if err := json.NewDecoder(buf).Decode(&m); err != nil {
        t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
    }
    return m
}
```

---

## What Not to Test

- Internal implementation details (unexported fields, private functions).
- Framework or stdlib behaviour (e.g. that `http.DefaultTransport` works correctly).
- Error paths that cannot happen given valid inputs from internal callers.
- Logging output format beyond what is required to verify correct behaviour (field presence is fine; exact string matching is fragile).
