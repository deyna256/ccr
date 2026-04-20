package task_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/task-planner/server/internal/task"
)

type stubStorage struct {
	list             func(ctx context.Context, userID string, f task.ListFilter) ([]task.Task, error)
	getByID          func(ctx context.Context, id, userID string) (task.Task, error)
	create           func(ctx context.Context, t task.Task) (task.Task, error)
	update           func(ctx context.Context, t task.Task) (task.Task, error)
	updateStatus     func(ctx context.Context, id, userID, status string, completedAt, archivedAt *time.Time) (task.Task, error)
	delete           func(ctx context.Context, id, userID string) error
	listAttachments  func(ctx context.Context, taskID, userID string) ([]task.Attachment, error)
	getAttachment    func(ctx context.Context, attachmentID, userID string) (task.Attachment, error)
	createAttachment func(ctx context.Context, a task.Attachment) (task.Attachment, error)
	deleteAttachment func(ctx context.Context, attachmentID, userID string) error
}

func (s *stubStorage) List(ctx context.Context, userID string, f task.ListFilter) ([]task.Task, error) {
	return s.list(ctx, userID, f)
}
func (s *stubStorage) GetByID(ctx context.Context, id, userID string) (task.Task, error) {
	return s.getByID(ctx, id, userID)
}
func (s *stubStorage) Create(ctx context.Context, t task.Task) (task.Task, error) {
	return s.create(ctx, t)
}
func (s *stubStorage) Update(ctx context.Context, t task.Task) (task.Task, error) {
	return s.update(ctx, t)
}
func (s *stubStorage) UpdateStatus(ctx context.Context, id, userID, status string, completedAt, archivedAt *time.Time) (task.Task, error) {
	return s.updateStatus(ctx, id, userID, status, completedAt, archivedAt)
}
func (s *stubStorage) Delete(ctx context.Context, id, userID string) error {
	return s.delete(ctx, id, userID)
}
func (s *stubStorage) ListAttachments(ctx context.Context, taskID, userID string) ([]task.Attachment, error) {
	return s.listAttachments(ctx, taskID, userID)
}
func (s *stubStorage) GetAttachment(ctx context.Context, attachmentID, userID string) (task.Attachment, error) {
	return s.getAttachment(ctx, attachmentID, userID)
}
func (s *stubStorage) CreateAttachment(ctx context.Context, a task.Attachment) (task.Attachment, error) {
	return s.createAttachment(ctx, a)
}
func (s *stubStorage) DeleteAttachment(ctx context.Context, attachmentID, userID string) error {
	return s.deleteAttachment(ctx, attachmentID, userID)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestService_UpdateStatus_done(t *testing.T) {
	regularTask := task.Task{
		ID:     "task1",
		UserID: "user1",
		Type:   "task",
		Status: "pending",
	}

	var gotCompletedAt *time.Time
	store := &stubStorage{
		getByID: func(_ context.Context, _, _ string) (task.Task, error) {
			return regularTask, nil
		},
		updateStatus: func(_ context.Context, _, _, status string, completedAt, _ *time.Time) (task.Task, error) {
			gotCompletedAt = completedAt
			updated := regularTask
			updated.Status = status
			updated.CompletedAt = completedAt
			return updated, nil
		},
	}

	svc := task.NewService(store, t.TempDir(), newTestLogger())
	result, err := svc.UpdateStatus(context.Background(), "task1", "user1", "done")
	if err != nil {
		t.Fatal(err)
	}
	if gotCompletedAt == nil {
		t.Error("expected completedAt to be set")
	}
	if result.Status != "done" {
		t.Errorf("expected status=done, got %q", result.Status)
	}
}

func TestService_UpdateStatus_invalidStatus(t *testing.T) {
	store := &stubStorage{}
	svc := task.NewService(store, t.TempDir(), newTestLogger())

	_, err := svc.UpdateStatus(context.Background(), "task1", "user1", "unknown-status")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_UploadAttachment_writesFile(t *testing.T) {
	dir := t.TempDir()
	taskID := "task-abc"
	content := "hello attachment content"

	store := &stubStorage{
		getByID: func(_ context.Context, id, userID string) (task.Task, error) {
			return task.Task{ID: id, UserID: userID, Type: "task", Status: "pending"}, nil
		},
		createAttachment: func(_ context.Context, a task.Attachment) (task.Attachment, error) {
			a.ID = "att1"
			return a, nil
		},
	}

	svc := task.NewService(store, dir, newTestLogger())
	att, err := svc.UploadAttachment(
		context.Background(),
		taskID, "user1",
		"test.txt", "text/plain",
		int64(len(content)),
		strings.NewReader(content),
	)
	if err != nil {
		t.Fatal(err)
	}
	if att.FilePath == "" {
		t.Fatal("expected FilePath to be set")
	}
	if _, err := os.Stat(att.FilePath); err != nil {
		t.Errorf("expected file to exist at %q: %v", att.FilePath, err)
	}
	data, err := os.ReadFile(att.FilePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("expected file content %q, got %q", content, string(data))
	}
}

func TestService_Create_emptyTitle(t *testing.T) {
	store := &stubStorage{}
	svc := task.NewService(store, t.TempDir(), newTestLogger())

	_, err := svc.Create(context.Background(), "user1", task.WriteRequest{Title: ""})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}
