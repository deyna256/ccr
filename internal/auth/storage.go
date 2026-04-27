package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

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
		 RETURNING id, email, password_hash, is_admin, created_at, updated_at`,
		email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
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
		`SELECT id, email, password_hash, is_admin, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
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
		`SELECT id, email, password_hash, is_admin, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth.storage.GetUserByID: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) SetAdmin(ctx context.Context, userID string, isAdmin bool) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET is_admin = $1, updated_at = NOW() WHERE id = $2`,
		isAdmin, userID,
	)
	if err != nil {
		return fmt.Errorf("auth.storage.SetAdmin: %w", err)
	}
	return nil
}

func (s *PostgresStorage) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET last_login = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("auth.storage.UpdateLastLogin: %w", err)
	}
	return nil
}

func (s *PostgresStorage) ListUserStats(ctx context.Context) ([]UserStats, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.is_admin, u.created_at, u.last_login,
		        COUNT(t.id) as task_count
		 FROM users u
		 LEFT JOIN tasks t ON t.user_id = u.id
		 GROUP BY u.id
		 ORDER BY u.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list user stats: %w", err)
	}
	defer rows.Close()

	var users []UserStats
	for rows.Next() {
		var u UserStats
		var lastLogin sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.IsAdmin, &u.CreatedAt, &lastLogin, &u.TaskCount); err != nil {
			return nil, fmt.Errorf("scan user stats: %w", err)
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *PostgresStorage) CreateRefreshTokenFamily(ctx context.Context, userID, token string, expiresAt time.Time) (string, error) {
	var familyID string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at, family_id, sequence)
		 VALUES ($1, $2, $3, gen_random_uuid(), 0)
		 RETURNING family_id`,
		userID, token, expiresAt,
	).Scan(&familyID)
	if err != nil {
		return "", fmt.Errorf("auth.storage.CreateRefreshTokenFamily: %w", err)
	}
	return familyID, nil
}

func (s *PostgresStorage) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	var rt RefreshToken
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, family_id, sequence, expires_at, revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()`,
		token,
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.FamilyID, &rt.Sequence, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return RefreshToken{}, ErrNotFound
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("auth.storage.GetRefreshToken: %w", err)
	}
	return rt, nil
}

func (s *PostgresStorage) AdvanceTokenFamily(ctx context.Context, familyID string, oldToken, newToken string, newExpiresAt time.Time) (reused bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}

	var currentSeq int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sequence), -1) FROM refresh_tokens WHERE family_id = $1 AND token = $2`,
		familyID, oldToken,
	).Scan(&currentSeq)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return false, fmt.Errorf("get current sequence: %w", err)
	}

	var maxSeq int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sequence), -1) FROM refresh_tokens WHERE family_id = $1`,
		familyID,
	).Scan(&maxSeq)
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("get max sequence: %w", err)
	}

	if currentSeq < maxSeq {
		_, err = tx.ExecContext(ctx,
			`UPDATE refresh_tokens SET revoked_at = NOW() WHERE family_id = $1`,
			familyID,
		)
		if err != nil {
			_ = tx.Rollback()
			return false, fmt.Errorf("revoke family on reuse: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit revoke family: %w", err)
		}
		return true, nil
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1`,
		oldToken,
	)
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("revoke old token: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at, family_id, sequence)
		 SELECT user_id, $2, $3, family_id, sequence + 1
		 FROM refresh_tokens WHERE token = $1`,
		oldToken, newToken, newExpiresAt,
	)
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("create new token in family: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit new token: %w", err)
	}
	return false, nil
}

func (s *PostgresStorage) RevokeAllInFamily(ctx context.Context, familyID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE family_id = $1`,
		familyID,
	)
	if err != nil {
		return fmt.Errorf("auth.storage.RevokeAllInFamily: %w", err)
	}
	return nil
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
