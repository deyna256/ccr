package category

import (
	"errors"
	"time"
)

type Category struct {
	ID        string
	UserID    string
	Name      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

var ErrNotFound = errors.New("not found")
var ErrInvalid  = errors.New("invalid input")
