package task

import (
	"errors"
	"time"
)

type Task struct {
	ID              string     `json:"id"`
	UserID          string     `json:"-"`
	Type            string     `json:"type"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	StartTime       *time.Time `json:"start_time,omitempty"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Status          string     `json:"status"`
	Color           *string    `json:"color,omitempty"`
	CompletedAt     *time.Time `json:"-"`
	ArchivedAt      *time.Time `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Attachment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Name      string    `json:"name"`
	FilePath  string    `json:"-"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

type ListFilter struct {
	From   *time.Time
	To     *time.Time
	Status *string
}

type WriteRequest struct {
	Type            string     `json:"type"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	DurationMinutes *int       `json:"duration_minutes"`
	Color           *string    `json:"color"`
}

type StatusRequest struct {
	Status string `json:"status"`
}

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
)
