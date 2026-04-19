package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/task-planner/server/internal/auth"
)

func issueTestToken(t *testing.T, secret string) string {
	t.Helper()
	st := &stubStorage{
		createUser: func(_ context.Context, _, _ string) (auth.User, error) {
			return auth.User{ID: "user-123"}, nil
		},
		createRefreshToken: func(_ context.Context, _, _ string, _ time.Time) error { return nil },
	}
	svc := auth.NewService(st, secret, "refresh")
	resp, err := svc.Register(context.Background(), auth.RegisterRequest{Email: "a@b.com", Password: "pass"})
	if err != nil {
		t.Fatalf("issueTestToken: %v", err)
	}
	return resp.AccessToken
}

func TestMiddleware_missingHeader(t *testing.T) {
	h := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestMiddleware_malformedBearer(t *testing.T) {
	h := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestMiddleware_validToken_injectsUserID(t *testing.T) {
	token := issueTestToken(t, "secret")
	var gotID string
	h := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	h := auth.Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
