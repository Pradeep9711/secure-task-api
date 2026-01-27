package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"secure-task-api/internal/models"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// TaskRepositoryInterface defines the interface for task repository
type TaskRepositoryInterface interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Task, error)
	GetAll(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Task, int, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	HealthCheck(ctx context.Context) error
}

// Repository aggregates all repository interfaces
type Repository struct {
	User UserRepositoryInterface
	Task TaskRepositoryInterface
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		User: NewUserRepository(db),
		Task: NewTaskRepository(db),
	}
}
