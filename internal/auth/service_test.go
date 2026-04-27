package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/task-planner/server/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type stubStorage struct {
	createUser               func(ctx context.Context, email, hash string) (auth.User, error)
	getUserByEmail           func(ctx context.Context, email string) (auth.User, error)
	getUserByID              func(ctx context.Context, id string) (auth.User, error)
	setAdmin                 func(ctx context.Context, userID string, isAdmin bool) error
	updateLastLogin          func(ctx context.Context, userID string) error
	listUserStats            func(ctx context.Context) ([]auth.UserStats, error)
	createRefreshToken       func(ctx context.Context, userID, token string, expiresAt time.Time) error
	createRefreshTokenFamily func(ctx context.Context, userID, token string, expiresAt time.Time) (string, error)
	getRefreshToken          func(ctx context.Context, token string) (auth.RefreshToken, error)
	advanceTokenFamily       func(ctx context.Context, familyID string, oldToken, newToken string, newExpiresAt time.Time) (bool, error)
	revokeAllInFamily        func(ctx context.Context, familyID string) error
	revokeRefreshToken       func(ctx context.Context, token string) error
}

func (s *stubStorage) CreateUser(ctx context.Context, email, hash string) (auth.User, error) {
	return s.createUser(ctx, email, hash)
}
func (s *stubStorage) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	return s.getUserByEmail(ctx, email)
}
func (s *stubStorage) GetUserByID(ctx context.Context, id string) (auth.User, error) {
	return s.getUserByID(ctx, id)
}
func (s *stubStorage) SetAdmin(ctx context.Context, userID string, isAdmin bool) error {
	if s.setAdmin == nil {
		return nil
	}
	return s.setAdmin(ctx, userID, isAdmin)
}
func (s *stubStorage) UpdateLastLogin(ctx context.Context, userID string) error {
	if s.updateLastLogin == nil {
		return nil
	}
	return s.updateLastLogin(ctx, userID)
}
func (s *stubStorage) ListUserStats(ctx context.Context) ([]auth.UserStats, error) {
	if s.listUserStats == nil {
		return nil, nil
	}
	return s.listUserStats(ctx)
}
func (s *stubStorage) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	return s.createRefreshToken(ctx, userID, token, expiresAt)
}
func (s *stubStorage) CreateRefreshTokenFamily(ctx context.Context, userID, token string, expiresAt time.Time) (string, error) {
	if s.createRefreshTokenFamily == nil {
		return "", nil
	}
	return s.createRefreshTokenFamily(ctx, userID, token, expiresAt)
}
func (s *stubStorage) GetRefreshToken(ctx context.Context, token string) (auth.RefreshToken, error) {
	return s.getRefreshToken(ctx, token)
}
func (s *stubStorage) AdvanceTokenFamily(ctx context.Context, familyID string, oldToken, newToken string, newExpiresAt time.Time) (bool, error) {
	if s.advanceTokenFamily == nil {
		return false, nil
	}
	return s.advanceTokenFamily(ctx, familyID, oldToken, newToken, newExpiresAt)
}
func (s *stubStorage) RevokeAllInFamily(ctx context.Context, familyID string) error {
	if s.revokeAllInFamily == nil {
		return nil
	}
	return s.revokeAllInFamily(ctx, familyID)
}
func (s *stubStorage) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.revokeRefreshToken(ctx, token)
}

func newTestService(st auth.Storage) *auth.Service {
	return auth.NewService(st, "test-secret", "test-refresh-secret", slog.Default())
}

func TestService_Register_hashesPassword(t *testing.T) {
	var storedHash string
	st := &stubStorage{
		createUser: func(_ context.Context, _, hash string) (auth.User, error) {
			storedHash = hash
			return auth.User{ID: "u1"}, nil
		},
		createRefreshTokenFamily: func(_ context.Context, _, _ string, _ time.Time) (string, error) { return "", nil },
	}
	_, err := newTestService(st).Register(context.Background(), auth.RegisterRequest{Email: "a@b.com", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if storedHash == "secret" {
		t.Error("password must not be stored as plaintext")
	}
	if storedHash == "" {
		t.Error("stored hash must not be empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("secret")); err != nil {
		t.Errorf("stored hash does not verify with original password: %v", err)
	}
}

func TestService_Register_duplicate(t *testing.T) {
	st := &stubStorage{
		createUser: func(_ context.Context, _, _ string) (auth.User, error) {
			return auth.User{}, auth.ErrDuplicate
		},
	}
	_, err := newTestService(st).Register(context.Background(), auth.RegisterRequest{Email: "a@b.com", Password: "secret"})
	if !errors.Is(err, auth.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestService_Login_wrongPassword(t *testing.T) {
	st := &stubStorage{
		getUserByEmail: func(_ context.Context, _ string) (auth.User, error) {
			return auth.User{ID: "u1", PasswordHash: "$2a$12$invalidhash"}, nil
		},
	}
	_, err := newTestService(st).Login(context.Background(), auth.LoginRequest{Email: "a@b.com", Password: "wrong"})
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("expected ErrNotFound for wrong password, got %v", err)
	}
}

func TestService_Login_unknownEmail(t *testing.T) {
	st := &stubStorage{
		getUserByEmail: func(_ context.Context, _ string) (auth.User, error) {
			return auth.User{}, auth.ErrNotFound
		},
	}
	_, err := newTestService(st).Login(context.Background(), auth.LoginRequest{Email: "x@y.com", Password: "pass"})
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Logout_success(t *testing.T) {
	var revokedToken string
	st := &stubStorage{
		revokeRefreshToken: func(_ context.Context, token string) error {
			revokedToken = token
			return nil
		},
	}
	err := newTestService(st).Logout(context.Background(), "refresh-token-abc")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if revokedToken != "refresh-token-abc" {
		t.Errorf("revoked token = %q, want %q", revokedToken, "refresh-token-abc")
	}
}

func TestService_Refresh_tokenReuse(t *testing.T) {
	st := &stubStorage{
		getRefreshToken: func(_ context.Context, _ string) (auth.RefreshToken, error) {
			return auth.RefreshToken{FamilyID: "family-1", UserID: "u1"}, nil
		},
		advanceTokenFamily: func(_ context.Context, familyID string, _, _ string, _ time.Time) (bool, error) {
			return true, nil
		},
	}
	_, err := newTestService(st).Refresh(context.Background(), "old-token")
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("expected ErrNotFound on token reuse, got %v", err)
	}
}

func TestService_Refresh_notFound(t *testing.T) {
	st := &stubStorage{
		getRefreshToken: func(_ context.Context, _ string) (auth.RefreshToken, error) {
			return auth.RefreshToken{}, auth.ErrNotFound
		},
	}
	_, err := newTestService(st).Refresh(context.Background(), "bad-token")
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestValidateAccessToken_tampered(t *testing.T) {
	_, err := auth.ValidateAccessToken([]byte("secret"), "tampered.token.value")
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
	if !errors.Is(err, auth.ErrNotFound) && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid token error, got %v", err)
	}
}

func TestValidateAccessToken_wrongSecret(t *testing.T) {
	svc := auth.NewService(&stubStorage{
		createRefreshTokenFamily: func(_ context.Context, _, _ string, _ time.Time) (string, error) { return "", nil },
		createUser: func(_ context.Context, _, _ string) (auth.User, error) {
			return auth.User{ID: "u1"}, nil
		},
	}, "secret-a", "refresh-secret", slog.Default())

	resp, err := svc.Register(context.Background(), auth.RegisterRequest{Email: "a@b.com", Password: "pass"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = auth.ValidateAccessToken([]byte("secret-b"), resp.AccessToken)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}
