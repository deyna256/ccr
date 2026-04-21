package auth_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/task-planner/server/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

func issueTestToken(t *testing.T, secret string) string {
	t.Helper()
	st := &stubStorage{
		createUser: func(_ context.Context, _, _ string) (auth.User, error) {
			return auth.User{ID: "user-123"}, nil
		},
		createRefreshToken: func(_ context.Context, _, _ string, _ time.Time) error { return nil },
	}
	svc := auth.NewService(st, secret, "refresh", slog.Default())
	resp, err := svc.Register(context.Background(), auth.RegisterRequest{Email: "a@b.com", Password: "pass"})
	if err != nil {
		t.Fatalf("issueTestToken: %v", err)
	}
	return resp.AccessToken
}

func issueExpiredToken(t *testing.T, secret string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": "user-123",
		"uid": "user-123",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("issueExpiredToken: %v", err)
	}
	return tokenStr
}

func testMiddleware(secret string) func(http.Handler) http.Handler {
	return auth.Middleware(secret, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestMiddleware_missingHeader(t *testing.T) {
	h := testMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestMiddleware_malformedBearer(t *testing.T) {
	h := testMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestMiddleware_validToken(t *testing.T) {
	token := issueTestToken(t, "secret")
	var gotID string
	h := testMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Error("user ID not in context")
		}
		gotID = id
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if gotID != "user-123" {
		t.Errorf("user_id = %q, want %q", gotID, "user-123")
	}
}

func TestMiddleware_invalidToken(t *testing.T) {
	h := testMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestMiddleware_expiredToken(t *testing.T) {
	token := issueExpiredToken(t, "secret")
	h := testMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
