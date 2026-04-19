package category_test

import (
	"context"
	"errors"
	"testing"

	"github.com/task-planner/server/internal/category"
)

type stubStorage struct {
	list     func(ctx context.Context, userID string) ([]category.Category, error)
	create   func(ctx context.Context, userID, name, color string) (category.Category, error)
	getByID  func(ctx context.Context, id, userID string) (category.Category, error)
	update   func(ctx context.Context, id, userID, name, color string) (category.Category, error)
	delete   func(ctx context.Context, id, userID string) error
}

func (s *stubStorage) List(ctx context.Context, userID string) ([]category.Category, error) {
	return s.list(ctx, userID)
}
func (s *stubStorage) Create(ctx context.Context, userID, name, color string) (category.Category, error) {
	return s.create(ctx, userID, name, color)
}
func (s *stubStorage) GetByID(ctx context.Context, id, userID string) (category.Category, error) {
	return s.getByID(ctx, id, userID)
}
func (s *stubStorage) Update(ctx context.Context, id, userID, name, color string) (category.Category, error) {
	return s.update(ctx, id, userID, name, color)
}
func (s *stubStorage) Delete(ctx context.Context, id, userID string) error {
	return s.delete(ctx, id, userID)
}

func TestService_Create_emptyName(t *testing.T) {
	svc := category.NewService(&stubStorage{})
	_, err := svc.Create(context.Background(), "u1", category.CreateRequest{Name: "", Color: "#FF0000"})
	if !errors.Is(err, category.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_invalidColor(t *testing.T) {
	svc := category.NewService(&stubStorage{})
	_, err := svc.Create(context.Background(), "u1", category.CreateRequest{Name: "Work", Color: "red"})
	if !errors.Is(err, category.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestService_Create_emptyColor(t *testing.T) {
	var gotColor string
	st := &stubStorage{
		create: func(_ context.Context, _, _, color string) (category.Category, error) {
			gotColor = color
			return category.Category{Color: color}, nil
		},
	}
	svc := category.NewService(st)
	_, err := svc.Create(context.Background(), "u1", category.CreateRequest{Name: "Work", Color: ""})
	if err != nil {
		t.Fatal(err)
	}
	if gotColor != "#3B82F6" {
		t.Errorf("default color = %q, want %q", gotColor, "#3B82F6")
	}
}

func TestService_Update_notFound(t *testing.T) {
	st := &stubStorage{
		update: func(_ context.Context, _, _, _, _ string) (category.Category, error) {
			return category.Category{}, category.ErrNotFound
		},
	}
	svc := category.NewService(st)
	_, err := svc.Update(context.Background(), "id1", "u1", category.UpdateRequest{Name: "X", Color: "#000000"})
	if !errors.Is(err, category.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Delete_notFound(t *testing.T) {
	st := &stubStorage{
		delete: func(_ context.Context, _, _ string) error {
			return category.ErrNotFound
		},
	}
	svc := category.NewService(st)
	err := svc.Delete(context.Background(), "id1", "u1")
	if !errors.Is(err, category.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
