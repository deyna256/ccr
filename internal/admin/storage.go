package admin

import (
	"context"
	"database/sql"
	"log/slog"
)

type AuthStorageAdapter struct {
	db  *sql.DB
	log *slog.Logger
}

func NewAuthStorageAdapter(db *sql.DB, log *slog.Logger) *AuthStorageAdapter {
	return &AuthStorageAdapter{db: db, log: log}
}

func (s *AuthStorageAdapter) SetAdmin(ctx context.Context, userID string, isAdmin bool) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET is_admin = $1, updated_at = NOW() WHERE id = $2`,
		isAdmin, userID,
	)
	return err
}

func (s *AuthStorageAdapter) ListUserStats(ctx context.Context) ([]UserStats, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.is_admin, u.created_at, u.last_login,
		        COUNT(t.id) as task_count
		 FROM users u
		 LEFT JOIN tasks t ON t.user_id = u.id
		 GROUP BY u.id
		 ORDER BY u.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserStats
	for rows.Next() {
		var u UserStats
		var lastLogin sql.NullTime
		if err := rows.Scan(&u.ID, &u.Email, &u.IsAdmin, &u.CreatedAt, &lastLogin, &u.TaskCount); err != nil {
			return nil, err
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}
		users = append(users, u)
	}
	return users, rows.Err()
}