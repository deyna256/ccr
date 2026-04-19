# Logging Guideline

## Overview

All services use structured JSON logging via `log/slog` (stdlib).
The shared library lives at `github.com/tfcp-site/httpx`.

Relevant packages:

| Package                                  | Responsibility                                      |
|------------------------------------------|-----------------------------------------------------|
| `github.com/tfcp-site/httpx/logging`     | Logger construction, context extractor pattern      |
| `github.com/tfcp-site/httpx/httplog`     | HTTP middleware: request/response log lines         |
| `github.com/tfcp-site/httpx/correlation` | Correlation ID propagation across service calls     |

---

## Log Format

### Standard fields

Field order is fixed across all log lines:

| # | Field     | Type   | Example                    |
|---|-----------|--------|----------------------------|
| 1 | `level`   | string | `INFO`, `WARN`, `ERROR`    |
| 2 | `service` | string | `"mentor"`                 |
| 3 | `msg`     | string | `"request"`                |
| 4 | _additional fields_ | — | —                    |

`time` is intentionally absent — Promtail records ingestion time automatically.

### Correlation

All log lines produced during a client request carry `request_id` — a correlation identifier that flows through the entire chain: middleware → handler → services → outgoing calls.

`request_id` is always the **first additional field** (position 4).

### Log types

**HTTP** — two lines per incoming request emitted by `httplog.Middleware`.

Arrival (`msg:"request"`):

```json
{"level":"INFO","service":"mentor","msg":"request","request_id":"550e8400-...","method":"POST","path":"/api/chat"}
```

Completion (`msg:"response"`), normal:

```json
{"level":"INFO","service":"mentor","msg":"response","request_id":"550e8400-...","method":"POST","path":"/api/chat","status":200,"duration":"1.240s"}
```

Completion, cache hit — `cached:true`, no `duration`:

```json
{"level":"INFO","service":"mentor","msg":"response","request_id":"550e8400-...","method":"GET","path":"/article/1","status":200,"cached":true}
```

`duration` and `cached` are mutually exclusive.

**App** — arbitrary application logs with additional fields defined per log site.

Within a client request — carry `request_id`:

```json
{"level":"INFO","service":"mentor","msg":"request","request_id":"550e8400-...","method":"POST","path":"/api/chat"}
{"level":"INFO","service":"mentor","msg":"cache miss","request_id":"550e8400-...","key":"chat:abc"}
{"level":"INFO","service":"mentor","msg":"llm response","request_id":"550e8400-...","tokens":312}
{"level":"INFO","service":"mentor","msg":"response","request_id":"550e8400-...","method":"POST","path":"/api/chat","status":200,"duration":"1.240s"}
```

Outside a client request (startup, background jobs) — no `request_id`:

```json
{"level":"INFO","service":"mentor","msg":"server started","addr":":8080"}
{"level":"INFO","service":"mentor","msg":"cleanup done","job":"cleanup","deleted":42}
```

### Field conventions

Use `snake_case` for all field names. Keep records flat — no nested objects.

| Field        | Type   | Source                      | Notes                                    |
|--------------|--------|-----------------------------|------------------------------------------|
| `level`      | string | slog                        | Always first                             |
| `service`    | string | `logging.New` arg           | Always second                            |
| `msg`        | string | call site                   | Always third                             |
| `request_id` | string | `correlation.Extractor`     | First additional field; present within a client request |
| `method`     | string | `httplog.Middleware`        | HTTP method                              |
| `path`       | string | `httplog.Middleware`        | Request path                             |
| `status`     | int    | `httplog.Middleware`        | Response status code                     |
| `duration`   | string | `httplog.Middleware`        | Absent for cached responses              |
| `cached`     | bool   | `httplog.MarkCached`        | Present only when `true`; mutually exclusive with `duration` |
| `error`      | string | call site                   | `slog.String("error", err.Error())`      |

---

## Logger Construction

Every service creates one logger at startup via `logging.New`:

```go
log := logging.New("mentor", slog.LevelInfo, correlation.Extractor())
```

Arguments:
- `service` — service name; attached as field #2 to every log line
- `level` — minimum log level (`slog.LevelInfo` in production)
- `extractors` — zero or more `ContextExtractor` functions called on every log record

Always pass context to log calls (`InfoContext`, `ErrorContext`, etc.) — otherwise `request_id` and other extractor-injected fields will not appear.

---

## HTTP Middleware Stack

Mount in this order on every HTTP server:

```go
handler := correlation.Middleware(
    httplog.Middleware(log)(mux),
)
```

`correlation.Middleware` must be outermost so the ID is in context before `httplog` reads it.

To mark a response as a cache hit, call `httplog.MarkCached` inside the handler:

```go
if hit := h.cache.Get(r.URL.Path); hit != nil {
    httplog.MarkCached(r.Context())
    w.Write(hit)
    return
}
```

---

## Correlation ID Propagation

**Inbound** — `correlation.Middleware` reads `X-Request-ID` from the request header. If absent, a new ID is generated via `xid`. The ID is stored in context and echoed back in the response header.

**Outbound** — use `correlation.Transport` to forward the ID to downstream services:

```go
client := &http.Client{
    Transport: &correlation.Transport{Base: http.DefaultTransport},
}
```

**Log injection** — pass `correlation.Extractor()` to `logging.New`. It reads the ID from context and injects it as `request_id` at position 4 on every log record:

```go
log := logging.New("examinator", slog.LevelInfo, correlation.Extractor())
log.InfoContext(ctx, "fetching user profile", slog.String("user_id", userID))
// → {"level":"INFO","service":"examinator","msg":"fetching user profile","request_id":"550e8400-...","user_id":"u-42"}
```

---

## Context Extractors

`ContextExtractor` is a function that reads attributes from context and returns them as `[]slog.Attr`. It is called on every log record.

Use it to attach identifiers that are always present for a given request scope (session ID, user ID, job ID):

```go
func sessionExtractor(ctx context.Context) []slog.Attr {
    if id, ok := ctx.Value(sessionKey{}).(string); ok && id != "" {
        return []slog.Attr{slog.String("session_id", id)}
    }
    return nil
}

log := logging.New("mock-interview", slog.LevelInfo,
    correlation.Extractor(),
    sessionExtractor,
)
```

Rules:
- Return `nil` when the value is absent — never return zero-value string attrs.
- Keep extractors pure — no I/O, no side effects.
- Register only fields that are always present for the service's main request type. One-off fields belong on individual log calls.
- `correlation.Extractor()` must always be first — it sets `request_id` at position 4.

---

## Log Levels

| Level   | When to use                                                  |
|---------|--------------------------------------------------------------|
| `DEBUG` | Developer diagnostics; disabled in production                |
| `INFO`  | Normal operation events (startup, request handled, job done) |
| `WARN`  | Recoverable anomalies (retry, fallback, degraded path)       |
| `ERROR` | Failures that affect correctness and require attention       |

Do not log and return an error at the same time — the caller decides whether to log.

---

## Observability Pipeline

```
Service (stdout, JSON)
    │
    │ Docker json-file driver (max 10 MB × 7 files per container)
    ▼
Promtail
    (Docker socket autodiscovery, labels: service / compose_service)
    │
    │ HTTP push
    ▼
Loki  (retention: 72 h)
    │
    ▼
Grafana
```

Because Promtail records ingestion time, services must **not** include a timestamp — `logging.New` strips it via `ReplaceAttr`.
