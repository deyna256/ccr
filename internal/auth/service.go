package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
	bcryptCost = 12
)

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
}

type Service struct {
	storage          Storage
	jwtSecret        []byte
	jwtRefreshSecret []byte
	log              *slog.Logger
}

func NewService(storage Storage, jwtSecret, jwtRefreshSecret string, log *slog.Logger) *Service {
	return &Service{
		storage:          storage,
		jwtSecret:        []byte(jwtSecret),
		jwtRefreshSecret: []byte(jwtRefreshSecret),
		log:              log,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	s.log.InfoContext(ctx, "registering user", slog.String("email", req.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Register: %w", err)
	}
	user, err := s.storage.CreateUser(ctx, req.Email, string(hash))
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Register: %w", err)
	}
	resp, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Register: %w", err)
	}
	return resp, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	s.log.InfoContext(ctx, "user login", slog.String("email", req.Email))
	user, err := s.storage.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResponse{}, ErrNotFound
		}
		return AuthResponse{}, fmt.Errorf("auth.service.Login: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, ErrNotFound
	}
	resp, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Login: %w", err)
	}
	return resp, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.storage.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("auth.service.Logout: %w", err)
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (AuthResponse, error) {
	s.log.InfoContext(ctx, "refreshing token")
	rt, err := s.storage.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Refresh: %w", err)
	}
	if err := s.storage.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Refresh: %w", err)
	}
	resp, err := s.issueTokenPair(ctx, rt.UserID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth.service.Refresh: %w", err)
	}
	return resp, nil
}

func (s *Service) issueTokenPair(ctx context.Context, userID string) (AuthResponse, error) {
	now := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		},
		UserID: userID,
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return AuthResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.storage.CreateRefreshToken(ctx, userID, refreshToken, now.Add(refreshTTL)); err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func ValidateAccessToken(secret []byte, tokenStr string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("auth: invalid token: %w", err)
	}
	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("auth: invalid token claims")
	}
	return Claims{UserID: c.UserID}, nil
}
