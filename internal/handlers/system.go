package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"secure-task-api/internal/logger"
	"secure-task-api/internal/models"
	"secure-task-api/internal/repository"
	"secure-task-api/pkg/utils"
)

// SystemHandler handles system endpoints like health checks
type SystemHandler struct {
	repo *repository.Repository
	log  *logger.Logger
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(repo *repository.Repository, log *logger.Logger) *SystemHandler {
	return &SystemHandler{
		repo: repo,
		log:  log,
	}
}

// RegisterRoutes registers system routes
func (h *SystemHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.HealthCheck)
	r.Get("/debug/panic", h.TriggerPanic)
}

// HealthCheck checks the health of the service
func (h *SystemHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database connection
	err := h.repo.Task.HealthCheck(ctx)
	if err != nil {
		h.log.WithError(err).Error("Database health check failed")
		utils.JSONResponse(w, http.StatusServiceUnavailable, models.HealthResponse{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Database:  "disconnected",
		})
		return
	}

	utils.JSONResponse(w, http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Database:  "connected",
	})
}

// TriggerPanic triggers a panic for testing Sentry integration
func (h *SystemHandler) TriggerPanic(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Panic endpoint triggered for testing")
	panic("Test panic for Sentry integration")
}
