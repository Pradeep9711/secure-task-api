package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Task represents a task in the system
type Task struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	DueDate     time.Time  `json:"due_date" db:"due_date"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response payload for authentication
type AuthResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=255"`
	Description string    `json:"description" validate:"required,min=1"`
	DueDate     time.Time `json:"due_date" validate:"required"`
}

// UpdateTaskRequest represents the request payload for updating a task
type UpdateTaskRequest struct {
	Title       string     `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description string     `json:"description,omitempty" validate:"omitempty,min=1"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      TaskStatus `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed"`
}

// TaskListResponse represents the response payload for listing tasks
type TaskListResponse struct {
	Tasks      []Task     `json:"tasks"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error      string    `json:"error"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
}

// IsValid checks if a TaskStatus is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted:
		return true
	}
	return false
}

// String returns the string representation of TaskStatus
func (s TaskStatus) String() string {
	return string(s)
}
