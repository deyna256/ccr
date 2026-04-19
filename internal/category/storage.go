package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Storage interface {
	List(ctx context.Context, userID string) ([]Category, error)
	Create(ctx context.Context, userID, name, color string) (Category, error)
	GetByID(ctx context.Context, id, userID string) (Category, error)
	Update(ctx context.Context, id, userID, name, color string) (Category, error)
	Delete(ctx context.Context, id, userID string) error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) List(ctx context.Context, userID string) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, color, created_at, updated_at
		 FROM categories WHERE user_id = $1 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("category.storage.List: %w", err)
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("category.storage.List: %w", err)
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category.storage.List: %w", err)
	}
	return cats, nil
}

func (s *PostgresStorage) Create(ctx context.Context, userID, name, color string) (Category, error) {
	var c Category
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO categories (user_id, name, color)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, name, color, created_at, updated_at`,
		userID, name, color,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Category{}, fmt.Errorf("category.storage.Create: %w", err)
	}
	return c, nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, id, userID string) (Category, error) {
	var c Category
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, color, created_at, updated_at
		 FROM categories WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	if err != nil {
		return Category{}, fmt.Errorf("category.storage.GetByID: %w", err)
	}
	return c, nil
}

func (s *PostgresStorage) Update(ctx context.Context, id, userID, name, color string) (Category, error) {
	var c Category
	err := s.db.QueryRowContext(ctx,
		`UPDATE categories SET name = $1, color = $2, updated_at = NOW()
		 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, name, color, created_at, updated_at`,
		name, color, id, userID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	if err != nil {
		return Category{}, fmt.Errorf("category.storage.Update: %w", err)
	}
	return c, nil
}

func (s *PostgresStorage) Delete(ctx context.Context, id, userID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM categories WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("category.storage.Delete: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("category.storage.Delete: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
