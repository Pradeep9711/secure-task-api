package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"secure-task-api/internal/models"
)

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new TaskRepository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task into the database
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, due_date, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	task.ID = uuid.New()
	now := time.Now()

	err := r.db.QueryRowContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.DueDate, task.UserID, now, now,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	return err
}

// GetByID retrieves a single task by ID and user
func (r *TaskRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, due_date, user_id, created_at, updated_at, deleted_at
		FROM tasks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	var task models.Task
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.DueDate,
		&task.UserID, &task.CreatedAt, &task.UpdatedAt, &task.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// GetAll retrieves all tasks for a user with pagination
func (r *TaskRepository) GetAll(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Task, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, title, description, status, due_date, user_id, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status,
			&task.DueDate, &task.UserID, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Update modifies an existing task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, due_date = $4, updated_at = NOW()
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate, task.ID, task.UserID,
	).Scan(&task.UpdatedAt)
}

// Delete marks a task as deleted
func (r *TaskRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `
		UPDATE tasks
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// HealthCheck verifies the database connection
func (r *TaskRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.db.PingContext(ctx)
}
