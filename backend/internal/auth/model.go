package auth

import (
	"errors"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID string
}

var ErrNotFound  = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")
