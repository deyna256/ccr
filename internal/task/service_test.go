package task_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
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

type stubFileStorage struct {
	mkdirErr  error
	createErr error
	removeErr error
	files     map[string]*stubFile
	created   []*stubFile
}

func (s *stubFileStorage) MkdirAll(path string, perms uint32) error {
	return s.mkdirErr
}

func (s *stubFileStorage) Create(taskID, filename string) (task.File, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.files == nil {
		s.files = make(map[string]*stubFile)
	}
	f := &stubFile{path: taskID + "/" + filename}
	s.files[f.path] = f
	s.created = append(s.created, f)
	return f, nil
}

func (s *stubFileStorage) Open(taskID string) (task.File, error) {
	return nil, task.ErrFileNotFound
}

func (s *stubFileStorage) Remove(path string) error {
	return s.removeErr
}

type stubFile struct {
	path string
}

func (f *stubFile) Write(p []byte) (int, error) {
	return len(p), nil
}

func (f *stubFile) Close() error {
	return nil
}

func (f *stubFile) Path() string {
	return f.path
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

	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())
	result, err := svc.UpdateStatus(context.Background(), "task1", "user1", "done")
	if err != nil {
		t.Fatal(err)
	}
	if gotCompletedAt == nil {
		t.Error("expected completedAt to be set")
	} else if time.Since(*gotCompletedAt) > 5*time.Second {
		t.Errorf("expected completedAt to be recent, got %v", *gotCompletedAt)
	}
	if result.Status != "done" {
		t.Errorf("expected status=done, got %q", result.Status)
	}
}

func TestService_UpdateStatus_invalidStatus(t *testing.T) {
	store := &stubStorage{}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	_, err := svc.UpdateStatus(context.Background(), "task1", "user1", "unknown-status")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_UploadAttachment_writesFile(t *testing.T) {
	taskID := "task-abc"

	store := &stubStorage{
		getByID: func(_ context.Context, id, userID string) (task.Task, error) {
			return task.Task{ID: id, UserID: userID, Type: "task", Status: "pending"}, nil
		},
		createAttachment: func(_ context.Context, a task.Attachment) (task.Attachment, error) {
			a.ID = "att1"
			return a, nil
		},
	}

	fs := &stubFileStorage{}
	svc := task.NewService(store, fs, newTestLogger())
	att, err := svc.UploadAttachment(
		context.Background(),
		taskID, "user1",
		"test.txt", "text/plain",
		13,
		strings.NewReader("hello attachment"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if att.ID != "att1" {
		t.Errorf("attachment ID = %q, want %q", att.ID, "att1")
	}
	if len(fs.created) != 1 {
		t.Errorf("expected 1 file created, got %d", len(fs.created))
	}
}

func TestService_Create_emptyTitle(t *testing.T) {
	store := &stubStorage{}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	_, err := svc.Create(context.Background(), "user1", task.WriteRequest{Title: ""})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Update_validData(t *testing.T) {
	existingTask := task.Task{
		ID:     "task1",
		UserID: "user1",
		Type:   "task",
		Title:  "old title",
		Status: "pending",
	}
	store := &stubStorage{
		getByID: func(_ context.Context, _, _ string) (task.Task, error) {
			return existingTask, nil
		},
		update: func(_ context.Context, t task.Task) (task.Task, error) {
			return t, nil
		},
	}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	updated, err := svc.Update(context.Background(), "task1", "user1", task.WriteRequest{Title: "new title"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Title != "new title" {
		t.Errorf("title = %q, want %q", updated.Title, "new title")
	}
}

func TestService_Delete_success(t *testing.T) {
	store := &stubStorage{
		getByID: func(_ context.Context, id, userID string) (task.Task, error) {
			return task.Task{ID: id, UserID: userID}, nil
		},
		delete: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	err := svc.Delete(context.Background(), "task1", "user1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestService_Delete_notFound(t *testing.T) {
	store := &stubStorage{
		getByID: func(_ context.Context, _, _ string) (task.Task, error) {
			return task.Task{}, task.ErrNotFound
		},
		delete: func(_ context.Context, _, _ string) error {
			return task.ErrNotFound
		},
	}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	err := svc.Delete(context.Background(), "task1", "user1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_List_withStatus(t *testing.T) {
	status := "done"
	store := &stubStorage{
		list: func(_ context.Context, _ string, f task.ListFilter) ([]task.Task, error) {
			if f.Status != nil && *f.Status == status {
				return []task.Task{{ID: "task1", Status: status}}, nil
			}
			return nil, nil
		},
	}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	tasks, err := svc.List(context.Background(), "user1", task.ListFilter{Status: &status})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != "done" {
		t.Errorf("status = %q, want %q", tasks[0].Status, "done")
	}
}

func TestService_UploadAttachment_taskNotFound(t *testing.T) {
	store := &stubStorage{
		getByID: func(_ context.Context, _, _ string) (task.Task, error) {
			return task.Task{}, task.ErrNotFound
		},
	}
	svc := task.NewService(store, &stubFileStorage{}, newTestLogger())

	_, err := svc.UploadAttachment(context.Background(), "task1", "user1", "file.txt", "text/plain", 0, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, task.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
