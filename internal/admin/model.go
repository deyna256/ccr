package admin

import (
	"errors"
	"time"
)

type UserStats struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	IsAdmin   bool       `json:"is_admin"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	TaskCount int        `json:"task_count"`
}

var ErrNotFound = errors.New("not found")