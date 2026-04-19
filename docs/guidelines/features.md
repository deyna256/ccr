# Feature Implementation Guideline

## Principles

- **Testability** — every package must be testable in isolation. Dependencies are hidden behind interfaces defined on the consumer side.
- **Observability** — every meaningful operation is logged. The correlation ID flows from the entry point to the deepest call.
- **Go-idiomatic** — standard project layout, standard error handling, no unnecessary abstractions.
- **Simplicity** — implement exactly what is required. No speculative generality, no over-engineering.

---

## Package Structure

Each service is a self-contained Go module. The typical layout:

```
service-name/
├── main.go          # wiring only: config, logger, dependencies, server start
├── config.go        # configuration struct loaded from environment
├── server.go        # HTTP router and middleware setup
├── handler.go       # HTTP handlers (or handler per domain area)
└── storage.go       # data access — implements a consumer-defined interface
```

For larger services, group by domain area rather than by layer:

```
service-name/
├── main.go
├── config.go
├── chat/
│   ├── handler.go
│   ├── service.go
│   └── storage.go
└── session/
    ├── handler.go
    └── storage.go
```

Rules:
- `main.go` contains only wiring — no business logic.
- Each package has a single clear responsibility.
- Do not create a package for a single type or function.

---

## Configuration

Load all configuration from environment variables at startup. Fail fast if a required value is absent.

```go
type Config struct {
    Addr        string
    DatabaseURL string
    LLMProvider string
}

func configFromEnv() (Config, error) {
    c := Config{
        Addr:        env("ADDR", ":8080"),
        DatabaseURL: mustEnv("DATABASE_URL"),
        LLMProvider: mustEnv("LLM_PROVIDER_URL"),
    }
    return c, nil
}

func mustEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatalf("required env variable %q is not set", key)
    }
    return v
}

func env(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

Never read `os.Getenv` outside of config loading. Pass the `Config` struct to everything that needs it.

---

## Dependency Injection and Interfaces

Define interfaces on the **consumer side** — in the package that uses the dependency, not the package that implements it.

```go
// in package chat — consumer defines the contract
type Storage interface {
    SaveMessage(ctx context.Context, msg Message) error
    History(ctx context.Context, sessionID string, limit int) ([]Message, error)
}

type Service struct {
    storage Storage
    log     *slog.Logger
}

func NewService(storage Storage, log *slog.Logger) *Service {
    return &Service{storage: storage, log: log}
}
```

The real implementation (e.g. `PostgresStorage`) lives in the same package or a sub-package and satisfies the interface without knowing about it.

Rules:
- Interfaces contain only the methods the consumer actually calls — no more.
- Accept interfaces, return concrete types.
- Inject dependencies through constructors, not global variables or `init`.

---

## main.go

`main.go` is the composition root. It wires everything together and starts the server. No logic here.

```go
func main() {
    cfg, err := configFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    logger := logging.New("mentor", slog.LevelInfo, correlation.Extractor())

    db, err := openDB(cfg.DatabaseURL)
    if err != nil {
        logger.Error("failed to connect to database", slog.String("error", err.Error()))
        os.Exit(1)
    }

    storage := chat.NewPostgresStorage(db)
    service := chat.NewService(storage, logger)
    handler := chat.NewHandler(service, logger)

    srv := newServer(cfg.Addr, handler, logger)
    logger.Info("server started", slog.String("addr", cfg.Addr))

    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        logger.Error("server error", slog.String("error", err.Error()))
        os.Exit(1)
    }
}
```

---

## HTTP Handlers

Handlers translate HTTP to domain calls and back. No business logic in handlers.

```go
type Handler struct {
    service *Service
    log     *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
    return &Handler{service: service, log: log}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    http.NewServeMux() // register routes in constructor or dedicated method
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    resp, err := h.service.Chat(r.Context(), req)
    if err != nil {
        h.log.ErrorContext(r.Context(), "chat failed", slog.String("error", err.Error()))
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

Rules:
- Decode input → call service → encode output.
- Log errors at the handler level with `ErrorContext` — pass the request context so `request_id` is attached.
- Return appropriate HTTP status codes. Do not swallow errors silently.

---

## Server Setup

Mount middleware in fixed order on every server:

```go
func newServer(addr string, h *Handler, log *slog.Logger) *http.Server {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/chat", h.handleChat)
    mux.HandleFunc("GET /api/history/{id}", h.handleHistory)

    handler := correlation.Middleware(
        httplog.Middleware(log)(mux),
    )

    return &http.Server{
        Addr:    addr,
        Handler: handler,
    }
}
```

Middleware order: `correlation.Middleware` outermost, `httplog.Middleware` next. See [logging guideline](logging.md).

---

## Observability

Log at the entry point (HTTP layer is covered by `httplog.Middleware`) and at significant domain events. Pass context to every log call.

```go
func (s *Service) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
    s.log.InfoContext(ctx, "llm request", slog.Int("history_len", len(req.History)))

    resp, err := s.llm.Complete(ctx, req)
    if err != nil {
        return ChatResponse{}, fmt.Errorf("llm complete: %w", err)
    }

    s.log.InfoContext(ctx, "llm response", slog.Int("tokens", resp.Tokens))
    return resp, nil
}
```

Rules:
- Always use `*Context` variants (`InfoContext`, `ErrorContext`) — this ensures `request_id` is present in every log line.
- Log the start and result of calls to external services (LLM, database, downstream HTTP).
- Do not log and return an error simultaneously — let the caller decide.
- See [logging guideline](logging.md) for field conventions and log format.

---

## Error Handling

Wrap errors with context at every layer boundary. Use `fmt.Errorf` with `%w`.

```go
// storage layer
func (s *PostgresStorage) SaveMessage(ctx context.Context, msg Message) error {
    _, err := s.db.ExecContext(ctx, `INSERT INTO messages ...`, msg.ID, msg.Text)
    if err != nil {
        return fmt.Errorf("save message: %w", err)
    }
    return nil
}

// service layer
func (s *Service) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
    if err := s.storage.SaveMessage(ctx, msg); err != nil {
        return ChatResponse{}, fmt.Errorf("chat: %w", err)
    }
    // ...
}
```

Rules:
- Each layer wraps with its own context: `"save message: %w"`, `"chat: %w"`.
- Do not discard errors. Do not use `_` for error returns except in tests.
- Sentinel errors (`errors.Is`) for conditions callers need to branch on. Wrapped errors for everything else.

---

## Testability

Structure code so that dependencies can be replaced with test implementations:

```go
// consumer-defined interface in chat package
type LLMClient interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

// test implementation
type stubLLM struct {
    resp CompletionResponse
    err  error
}

func (s *stubLLM) Complete(_ context.Context, _ CompletionRequest) (CompletionResponse, error) {
    return s.resp, s.err
}

func TestService_Chat_returnsLLMResponse(t *testing.T) {
    svc := NewService(&stubLLM{resp: CompletionResponse{Text: "hello"}}, slog.Default())
    resp, err := svc.Chat(context.Background(), ChatRequest{})
    if err != nil {
        t.Fatal(err)
    }
    if resp.Text != "hello" {
        t.Errorf("Text = %q, want %q", resp.Text, "hello")
    }
}
```

See [testing guideline](testing.md) for full testing conventions.
