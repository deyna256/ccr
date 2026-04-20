package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Storage interface {
	List(ctx context.Context, userID string, f ListFilter) ([]Task, error)
	GetByID(ctx context.Context, id, userID string) (Task, error)
	Create(ctx context.Context, t Task) (Task, error)
	Update(ctx context.Context, t Task) (Task, error)
	UpdateStatus(ctx context.Context, id, userID, status string, completedAt, archivedAt *time.Time) (Task, error)
	Delete(ctx context.Context, id, userID string) error

	ListAttachments(ctx context.Context, taskID, userID string) ([]Attachment, error)
	GetAttachment(ctx context.Context, attachmentID, userID string) (Attachment, error)
	CreateAttachment(ctx context.Context, a Attachment) (Attachment, error)
	DeleteAttachment(ctx context.Context, attachmentID, userID string) error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func scanTask(row interface {
	Scan(dest ...any) error
}) (Task, error) {
	var t Task
	var categoryID sql.NullString
	var description sql.NullString
	var startTime sql.NullTime
	var endTime sql.NullTime
	var color sql.NullString
	var completedAt sql.NullTime
	var archivedAt sql.NullTime

	err := row.Scan(
		&t.ID,
		&t.UserID,
		&categoryID,
		&t.Type,
		&t.Title,
		&description,
		&startTime,
		&endTime,
		&t.Status,
		&color,
		&completedAt,
		&archivedAt,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}

	if categoryID.Valid {
		t.CategoryID = &categoryID.String
	}
	if description.Valid {
		t.Description = &description.String
	}
	if startTime.Valid {
		t.StartTime = &startTime.Time
	}
	if endTime.Valid {
		t.EndTime = &endTime.Time
	}
	if color.Valid {
		t.Color = &color.String
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if archivedAt.Valid {
		t.ArchivedAt = &archivedAt.Time
	}
	if t.StartTime != nil && t.EndTime != nil {
		d := int(t.EndTime.Sub(*t.StartTime).Minutes())
		t.DurationMinutes = &d
	}

	return t, nil
}

const taskColumns = `id, user_id, category_id, type, title, description, start_time, end_time,
	status, color, completed_at, archived_at, created_at, updated_at`

func (s *PostgresStorage) List(ctx context.Context, userID string, f ListFilter) ([]Task, error) {
	var rows *sql.Rows
	var err error

	switch {
	case f.Status != nil:
		rows, err = s.db.QueryContext(ctx,
			`SELECT `+taskColumns+` FROM tasks
			 WHERE user_id=$1 AND is_recurring=false AND status=$2
			 ORDER BY updated_at DESC`,
			userID, *f.Status,
		)
	case f.From != nil && f.To != nil:
		rows, err = s.db.QueryContext(ctx,
			`SELECT `+taskColumns+` FROM tasks
			 WHERE user_id=$1 AND is_recurring=false AND status='pending' AND start_time >= $2 AND start_time <= $3
			 ORDER BY start_time`,
			userID, f.From, f.To,
		)
	default:
		rows, err = s.db.QueryContext(ctx,
			`SELECT `+taskColumns+` FROM tasks
			 WHERE user_id=$1 AND is_recurring=false AND status='pending'
			 ORDER BY created_at`,
			userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("task.storage.List: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("task.storage.List: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task.storage.List: %w", err)
	}
	return tasks, nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, id, userID string) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+taskColumns+` FROM tasks WHERE id=$1 AND user_id=$2`,
		id, userID,
	)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("task.storage.GetByID: %w", err)
	}
	return t, nil
}

func (s *PostgresStorage) Create(ctx context.Context, t Task) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO tasks (user_id, type, title, description, start_time, end_time, status, color)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING `+taskColumns,
		t.UserID, t.Type, t.Title, t.Description, t.StartTime, t.EndTime, t.Status, t.Color,
	)
	created, err := scanTask(row)
	if err != nil {
		return Task{}, fmt.Errorf("task.storage.Create: %w", err)
	}
	return created, nil
}

func (s *PostgresStorage) Update(ctx context.Context, t Task) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`UPDATE tasks SET
		  type=$1, title=$2, description=$3, start_time=$4, end_time=$5, color=$6, updated_at=NOW()
		 WHERE id=$7 AND user_id=$8
		 RETURNING `+taskColumns,
		t.Type, t.Title, t.Description, t.StartTime, t.EndTime, t.Color, t.ID, t.UserID,
	)
	updated, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("task.storage.Update: %w", err)
	}
	return updated, nil
}

func (s *PostgresStorage) UpdateStatus(ctx context.Context, id, userID, status string, completedAt, archivedAt *time.Time) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`UPDATE tasks SET status=$1, completed_at=$2, archived_at=$3, updated_at=NOW()
		 WHERE id=$4 AND user_id=$5
		 RETURNING `+taskColumns,
		status, completedAt, archivedAt, id, userID,
	)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("task.storage.UpdateStatus: %w", err)
	}
	return t, nil
}

func (s *PostgresStorage) Delete(ctx context.Context, id, userID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE id=$1 AND user_id=$2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("task.storage.Delete: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("task.storage.Delete: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStorage) ListAttachments(ctx context.Context, taskID, userID string) ([]Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT a.id, a.task_id, a.name, a.file_path, a.file_size, a.mime_type, a.created_at
		 FROM attachments a
		 JOIN tasks t ON t.id = a.task_id
		 WHERE a.task_id=$1 AND t.user_id=$2
		 ORDER BY a.created_at`,
		taskID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("task.storage.ListAttachments: %w", err)
	}
	defer rows.Close()

	var atts []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.Name, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("task.storage.ListAttachments: %w", err)
		}
		atts = append(atts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task.storage.ListAttachments: %w", err)
	}
	return atts, nil
}

func (s *PostgresStorage) GetAttachment(ctx context.Context, attachmentID, userID string) (Attachment, error) {
	var a Attachment
	err := s.db.QueryRowContext(ctx,
		`SELECT a.id, a.task_id, a.name, a.file_path, a.file_size, a.mime_type, a.created_at
		 FROM attachments a
		 JOIN tasks t ON t.id = a.task_id
		 WHERE a.id=$1 AND t.user_id=$2`,
		attachmentID, userID,
	).Scan(&a.ID, &a.TaskID, &a.Name, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Attachment{}, ErrNotFound
	}
	if err != nil {
		return Attachment{}, fmt.Errorf("task.storage.GetAttachment: %w", err)
	}
	return a, nil
}

func (s *PostgresStorage) CreateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	var created Attachment
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO attachments (task_id, name, file_path, file_size, mime_type)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, task_id, name, file_path, file_size, mime_type, created_at`,
		a.TaskID, a.Name, a.FilePath, a.FileSize, a.MimeType,
	).Scan(&created.ID, &created.TaskID, &created.Name, &created.FilePath, &created.FileSize, &created.MimeType, &created.CreatedAt)
	if err != nil {
		return Attachment{}, fmt.Errorf("task.storage.CreateAttachment: %w", err)
	}
	return created, nil
}

func (s *PostgresStorage) DeleteAttachment(ctx context.Context, attachmentID, userID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM attachments a
		 USING tasks t
		 WHERE a.task_id = t.id AND a.id=$1 AND t.user_id=$2`,
		attachmentID, userID,
	)
	if err != nil {
		return fmt.Errorf("task.storage.DeleteAttachment: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("task.storage.DeleteAttachment: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
