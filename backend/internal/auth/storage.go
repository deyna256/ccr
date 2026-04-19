package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Storage interface {
	CreateUser(ctx context.Context, email, hash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateUser(ctx context.Context, email, hash string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at, updated_at`,
		email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return User{}, ErrDuplicate
		}
		return User{}, fmt.Errorf("auth.storage.CreateUser: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth.storage.GetUserByEmail: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth.storage.GetUserByID: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("auth.storage.CreateRefreshToken: %w", err)
	}
	return nil
}

func (s *PostgresStorage) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	var rt RefreshToken
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, expires_at, revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()`,
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return RefreshToken{}, ErrNotFound
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("auth.storage.GetRefreshToken: %w", err)
	}
	return rt, nil
}

func (s *PostgresStorage) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("auth.storage.RevokeRefreshToken: %w", err)
	}
	return nil
}
