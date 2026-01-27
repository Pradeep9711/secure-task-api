package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"secure-task-api/internal/logger"
	"secure-task-api/internal/middleware"
	"secure-task-api/internal/models"
	"secure-task-api/internal/repository"
	"secure-task-api/pkg/utils"
)

type TaskHandler struct {
	repo *repository.Repository
	log  *logger.Logger
}

func NewTaskHandler(repo *repository.Repository, log *logger.Logger) *TaskHandler {
	return &TaskHandler{
		repo: repo,
		log:  log,
	}
}

func (h *TaskHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListTasks)
	r.Post("/", h.CreateTask)
	r.Get("/{id}", h.GetTask)
	r.Put("/{id}", h.UpdateTask)
	r.Delete("/{id}", h.DeleteTask)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("Invalid user ID in context")
		utils.InternalServerError(w, "Invalid user context")
		return
	}

	page, limit := utils.GetPaginationParams(r)

	tasks, total, err := h.repo.Task.GetAll(r.Context(), userID, page, limit)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch tasks")
		utils.InternalServerError(w, "Failed to get tasks")
		return
	}

	totalPages := (total + limit - 1) / limit

	utils.JSONSuccess(w, http.StatusOK, models.TaskListResponse{
		Tasks: tasks,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("Invalid user ID in context")
		utils.InternalServerError(w, "Invalid user context")
		return
	}

	var req models.CreateTaskRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Title == "" {
		utils.BadRequest(w, "Title is required")
		return
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatusPending,
		DueDate:     req.DueDate,
		UserID:      userID,
	}

	if err := h.repo.Task.Create(r.Context(), task); err != nil {
		h.log.WithError(err).Error("Failed to create task")
		utils.InternalServerError(w, "Failed to create task")
		return
	}

	utils.JSONSuccess(w, http.StatusCreated, map[string]interface{}{
		"task": task,
	})
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("Invalid user ID in context")
		utils.InternalServerError(w, "Invalid user context")
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid task ID")
		return
	}

	task, err := h.repo.Task.GetByID(r.Context(), taskID, userID)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch task")
		utils.InternalServerError(w, "Failed to get task")
		return
	}

	if task == nil {
		utils.NotFound(w, "Task not found")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, map[string]interface{}{
		"task": task,
	})
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("Invalid user ID in context")
		utils.InternalServerError(w, "Invalid user context")
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid task ID")
		return
	}

	var req models.UpdateTaskRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	task, err := h.repo.Task.GetByID(r.Context(), taskID, userID)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch task")
		utils.InternalServerError(w, "Failed to update task")
		return
	}

	if task == nil {
		utils.NotFound(w, "Task not found")
		return
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status.IsValid() {
		task.Status = req.Status
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}

	if err := h.repo.Task.Update(r.Context(), task); err != nil {
		h.log.WithError(err).Error("Failed to update task")
		utils.InternalServerError(w, "Failed to update task")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, map[string]interface{}{
		"task": task,
	})
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		utils.Unauthorized(w, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("Invalid user ID in context")
		utils.InternalServerError(w, "Invalid user context")
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid task ID")
		return
	}

	if err := h.repo.Task.Delete(r.Context(), taskID, userID); err != nil {
		h.log.WithError(err).Error("Failed to delete task")
		utils.InternalServerError(w, "Failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
