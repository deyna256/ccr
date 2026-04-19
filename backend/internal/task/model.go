package task

import (
	"errors"
	"time"
)

type Task struct {
	ID             string
	UserID         string
	CategoryID     *string
	Type           string
	Title          string
	Description    *string
	StartTime      *time.Time
	EndTime        *time.Time
	Status         string
	CompletedAt    *time.Time
	ArchivedAt     *time.Time
	IsRecurring    bool
	RecurrenceRule *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RecurrenceID   *string
}

type Attachment struct {
	ID        string
	TaskID    string
	Name      string
	FilePath  string
	FileSize  int64
	MimeType  string
	CreatedAt time.Time
}

type RecurrenceRule struct {
	Freq     string   `json:"freq"`
	Interval int      `json:"interval"`
	Days     []string `json:"days,omitempty"`
	Until    string   `json:"until,omitempty"`
}

type ListFilter struct {
	From *time.Time
	To   *time.Time
}

type WriteRequest struct {
	CategoryID     *string    `json:"category_id"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	IsRecurring    bool       `json:"is_recurring"`
	RecurrenceRule *string    `json:"recurrence_rule"`
}

type StatusRequest struct {
	Status string `json:"status"`
}

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
)
