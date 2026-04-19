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

type Service struct {
	storage         Storage
	fileStoragePath string
	log             *slog.Logger
}

func NewService(storage Storage, fileStoragePath string, log *slog.Logger) *Service {
	return &Service{storage: storage, fileStoragePath: fileStoragePath, log: log}
}

func (s *Service) List(ctx context.Context, userID string, f ListFilter) ([]Task, error) {
	regular, err := s.storage.List(ctx, userID, f)
	if err != nil {
		return nil, fmt.Errorf("task.service.List: %w", err)
	}
	if f.From == nil || f.To == nil {
		return regular, nil
	}
	templates, err := s.storage.ListRecurring(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("task.service.List: %w", err)
	}
	result := append([]Task{}, regular...)
	for _, tmpl := range templates {
		occurrences, err := expand(tmpl, *f.From, *f.To)
		if err != nil {
			return nil, fmt.Errorf("task.service.List: %w", err)
		}
		result = append(result, occurrences...)
	}
	return result, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID string) (Task, error) {
	t, err := s.storage.GetByID(ctx, id, userID)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.GetByID: %w", err)
	}
	return t, nil
}

func (s *Service) Create(ctx context.Context, userID string, req WriteRequest) (Task, error) {
	if err := validateWriteRequest(req); err != nil {
		return Task{}, fmt.Errorf("task.service.Create: %w", err)
	}
	taskType := req.Type
	if taskType == "" {
		taskType = "task"
	}
	endTime := req.EndTime
	if endTime == nil && req.DurationMinutes != nil && req.StartTime != nil {
		end := req.StartTime.Add(time.Duration(*req.DurationMinutes) * time.Minute)
		endTime = &end
	}
	t := Task{
		UserID:         userID,
		CategoryID:     req.CategoryID,
		Type:           taskType,
		Title:          req.Title,
		Description:    req.Description,
		StartTime:      req.StartTime,
		EndTime:        endTime,
		Status:         "pending",
		Color:          req.Color,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
	}
	created, err := s.storage.Create(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.Create: %w", err)
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req WriteRequest) (Task, error) {
	if err := validateWriteRequest(req); err != nil {
		return Task{}, fmt.Errorf("task.service.Update: %w", err)
	}
	endTime := req.EndTime
	if endTime == nil && req.DurationMinutes != nil && req.StartTime != nil {
		end := req.StartTime.Add(time.Duration(*req.DurationMinutes) * time.Minute)
		endTime = &end
	}
	t := Task{
		ID:             id,
		UserID:         userID,
		CategoryID:     req.CategoryID,
		Type:           req.Type,
		Title:          req.Title,
		Description:    req.Description,
		StartTime:      req.StartTime,
		EndTime:        endTime,
		Color:          req.Color,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
	}
	updated, err := s.storage.Update(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.Update: %w", err)
	}
	return updated, nil
}

func validateWriteRequest(req WriteRequest) error {
	if req.Title == "" {
		return ErrInvalid
	}
	if req.Type != "" && req.Type != "task" && req.Type != "event" {
		return ErrInvalid
	}
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, id, userID, status string) (Task, error) {
	var completedAt, archivedAt *time.Time
	now := time.Now()
	switch status {
	case "done":
		completedAt = &now
	case "archived":
		archivedAt = &now
	case "pending":
		// both nil
	default:
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", ErrInvalid)
	}
	existing, err := s.storage.GetByID(ctx, id, userID)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", err)
	}
	if status == "done" && existing.Type == "event" {
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", ErrInvalid)
	}
	t, err := s.storage.UpdateStatus(ctx, id, userID, status, completedAt, archivedAt)
	if err != nil {
		return Task{}, fmt.Errorf("task.service.UpdateStatus: %w", err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
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
	if _, err := s.storage.GetByID(ctx, taskID, userID); err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: %w", err)
	}
	dir := filepath.Join(s.fileStoragePath, taskID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: mkdir: %w", err)
	}
	ext := filepath.Ext(name)
	filePath := filepath.Join(dir, xid.New().String()+ext)
	f, err := os.Create(filePath)
	if err != nil {
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: create file: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = os.Remove(filePath)
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: write file: %w", err)
	}
	_ = f.Close()
	a, err := s.storage.CreateAttachment(ctx, Attachment{
		TaskID:   taskID,
		Name:     name,
		FilePath: filePath,
		FileSize: size,
		MimeType: mimeType,
	})
	if err != nil {
		_ = os.Remove(filePath)
		return Attachment{}, fmt.Errorf("task.service.UploadAttachment: %w", err)
	}
	return a, nil
}

func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, userID string) error {
	a, err := s.storage.GetAttachment(ctx, attachmentID, userID)
	if err != nil {
		return fmt.Errorf("task.service.DeleteAttachment: %w", err)
	}
	if err := s.storage.DeleteAttachment(ctx, attachmentID, userID); err != nil {
		return fmt.Errorf("task.service.DeleteAttachment: %w", err)
	}
	if err := os.Remove(a.FilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.log.WarnContext(ctx, "failed to remove attachment file", slog.String("path", a.FilePath), slog.String("error", err.Error()))
	}
	return nil
}
