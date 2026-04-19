package category

import (
	"context"
	"fmt"
	"regexp"
)

var colorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

const defaultColor = "#3B82F6"

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) List(ctx context.Context, userID string) ([]Category, error) {
	cats, err := s.storage.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("category.service.List: %w", err)
	}
	return cats, nil
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Category, error) {
	if req.Name == "" {
		return Category{}, fmt.Errorf("category.service.Create: %w", ErrInvalid)
	}
	color := req.Color
	if color == "" {
		color = defaultColor
	} else if !colorRe.MatchString(color) {
		return Category{}, fmt.Errorf("category.service.Create: %w", ErrInvalid)
	}
	cat, err := s.storage.Create(ctx, userID, req.Name, color)
	if err != nil {
		return Category{}, fmt.Errorf("category.service.Create: %w", err)
	}
	return cat, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (Category, error) {
	if req.Name == "" {
		return Category{}, fmt.Errorf("category.service.Update: %w", ErrInvalid)
	}
	color := req.Color
	if color == "" {
		color = defaultColor
	} else if !colorRe.MatchString(color) {
		return Category{}, fmt.Errorf("category.service.Update: %w", ErrInvalid)
	}
	cat, err := s.storage.Update(ctx, id, userID, req.Name, color)
	if err != nil {
		return Category{}, fmt.Errorf("category.service.Update: %w", err)
	}
	return cat, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	if err := s.storage.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("category.service.Delete: %w", err)
	}
	return nil
}
