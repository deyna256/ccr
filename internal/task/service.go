package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/xid"
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

type FileStorage interface {
	Open(taskID string) (File, error)
	Create(taskID, filename string) (File, error)
	Remove(path string) error
	MkdirAll(path string, perms uint32) error
}

type File interface {
	io.Writer
	io.Closer
	Path() string
}

type Service struct {
	storage       Storage
	fileStorage   FileStorage
	log           *slog.Logger
}

func NewService(storage Storage, fileStorage FileStorage, log *slog.Logger) *Service {
	return &Service{storage: storage, fileStorage: fileStorage, log: log}
}

func (s *Service) List(ctx context.Context, userID string, f ListFilter) ([]Task, error) {
	tasks, err := s.storage.List(ctx, userID, f)
	if err != nil {
		return nil, fmt.Errorf("task.service.List: %w", err)
	}
	return tasks, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID string) (Task, error) {
	t, err := s.storage.GetByID(ctx, id, userID)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.GetByID: %w", err)
	}
	return t, nil
}

func (s *Service) Create(ctx context.Context, userID string, req WriteRequest) (Task, error) {
	s.log.InfoContext(ctx, "creating task")
	if req.Title == "" {
		return Task{}, fmt.Errorf("task.service.Create: %w", ErrInvalid)
	}
	endTime := req.EndTime
	if endTime == nil && req.DurationMinutes != nil && req.StartTime != nil {
		end := req.StartTime.Add(time.Duration(*req.DurationMinutes) * time.Minute)
		endTime = &end
	}
	t := Task{
		UserID:      userID,
		Type:        "task",
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     endTime,
		Status:      "pending",
		Color:       req.Color,
	}
	created, err := s.storage.Create(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.Create: %w", err)
	}
	s.log.InfoContext(ctx, "task created", slog.String("task_id", created.ID))
	return created, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req WriteRequest) (Task, error) {
	s.log.InfoContext(ctx, "updating task", slog.String("task_id", id))
	if req.Title == "" {
		return Task{}, fmt.Errorf("task.service.Update: %w", ErrInvalid)
	}
	endTime := req.EndTime
	if endTime == nil && req.DurationMinutes != nil && req.StartTime != nil {
		end := req.StartTime.Add(time.Duration(*req.DurationMinutes) * time.Minute)
		endTime = &end
	}
	t := Task{
		ID:          id,
		UserID:      userID,
		Type:        "task",
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     endTime,
		Color:       req.Color,
	}
	updated, err := s.storage.Update(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.Update: %w", err)
	}
	return updated, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id, userID, status string) (Task, error) {
	s.log.InfoContext(ctx, "updating task status", slog.String("task_id", id), slog.String("status", status))
	var completedAt, archivedAt *time.Time
	now := time.Now()
	switch status {
	case "done":
		completedAt = &now
	case "archived":
		archivedAt = &now
	case "pending":
	default:
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", ErrInvalid)
	}
	t, err := s.storage.UpdateStatus(ctx, id, userID, status, completedAt, archivedAt)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	s.log.InfoContext(ctx, "deleting task", slog.String("task_id", id))
	if err := s.storage.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("task.service.Delete: %w", err)
	}
	return nil
}

func (s *Service) ListAttachments(ctx context.Context, taskID, userID string) ([]Attachment, error) {
	att, err := s.storage.ListAttachments(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task.service.ListAttachments: %w", err)
	}
	return att, nil
}

func (s *Service) GetAttachment(ctx context.Context, attachmentID, userID string) (Attachment, error) {
	a, err := s.storage.GetAttachment(ctx, attachmentID, userID)
	if err != nil {
		return Attachment{}, fmt.Errorf("task.service.GetAttachment: %w", err)
	}
	return a, nil
}

func (s *Service) UploadAttachment(ctx context.Context, taskID, userID, name, mimeType string, size int64, r io.Reader) (Attachment, error) {
	s.log.InfoContext(ctx, "uploading attachment", slog.String("task_id", taskID), slog.String("filename", name))
	if _, err := s.storage.GetByID(ctx, taskID, userID); err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: %w", err)
	}
	dir := taskID
	if err := s.fileStorage.MkdirAll(dir, 0750); err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: mkdir: %w", err)
	}
	ext := filepath.Ext(name)
	filePath := xid.New().String() + ext
	f, err := s.fileStorage.Create(dir, filePath)
	if err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: create file: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = s.fileStorage.Remove(f.Path())
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: write file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = s.fileStorage.Remove(f.Path())
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: close file: %w", err)
	}
	a, err := s.storage.CreateAttachment(ctx, Attachment{
		TaskID:   taskID,
		Name:     name,
		FilePath: f.Path(),
		FileSize: size,
		MimeType: mimeType,
	})
	if err != nil {
		if rmErr := s.fileStorage.Remove(f.Path()); rmErr != nil {
			return Attachment{}, fmt.Errorf("task.service.UploadAttachment: %w (file %s orphaned: %v)", err, f.Path(), rmErr)
		}
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: %w (file %s orphaned)", err, f.Path())
	}
	return a, nil
}

func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, userID string) error {
	s.log.InfoContext(ctx, "deleting attachment", slog.String("attachment_id", attachmentID))
	a, err := s.storage.GetAttachment(ctx, attachmentID, userID)
	if err != nil {
		return fmt.Errorf("task.service.DeleteAttachment: %w", err)
	}
	if err := s.storage.DeleteAttachment(ctx, attachmentID, userID); err != nil {
		return fmt.Errorf("task.service.DeleteAttachment: %w", err)
	}
	if err := s.fileStorage.Remove(a.FilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.log.WarnContext(ctx, "failed to remove attachment file", slog.String("path", a.FilePath), slog.String("error", err.Error()))
	}
	return nil
}
