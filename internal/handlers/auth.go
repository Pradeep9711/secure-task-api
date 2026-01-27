package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"secure-task-api/internal/auth"
	"secure-task-api/internal/logger"
	"secure-task-api/internal/models"
	"secure-task-api/internal/repository"
	"secure-task-api/pkg/utils"
)

type AuthHandler struct {
	repo       *repository.Repository
	jwtManager *auth.JWTManager
	log        *logger.Logger
}

// Wires repository, JWT logic, and logger into the auth handler.
func NewAuthHandler(
	repo *repository.Repository,
	jwtManager *auth.JWTManager,
	log *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		repo:       repo,
		jwtManager: jwtManager,
		log:        log,
	}
}

// Registers auth routes under /v1/auth.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
}

// Creates a new user account and returns a token pair on success.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		h.log.WithError(err).Error("invalid register payload")
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		utils.BadRequest(w, "Email, password, and name are required")
		return
	}

	// Prevent duplicate accounts by email.
	existingUser, err := h.repo.User.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.log.WithError(err).Error("failed to check existing user")
		utils.InternalServerError(w, "Failed to register user")
		return
	}
	if existingUser != nil {
		utils.BadRequest(w, "User with this email already exists")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.log.WithError(err).Error("password hashing failed")
		utils.InternalServerError(w, "Failed to register user")
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
	}

	if err := h.repo.User.Create(r.Context(), user); err != nil {
		h.log.WithError(err).Error("failed to persist user")
		utils.InternalServerError(w, "Failed to register user")
		return
	}

	accessToken, refreshToken, err :=
		h.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		h.log.WithError(err).Error("token generation failed")
		utils.InternalServerError(w, "Failed to register user")
		return
	}

	utils.JSONSuccess(w, http.StatusCreated, models.AuthResponse{
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

// Authenticates a user and returns a fresh token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		h.log.WithError(err).Error("invalid login payload")
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "Email and password are required")
		return
	}

	user, err := h.repo.User.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.log.WithError(err).Error("failed to fetch user during login")
		utils.InternalServerError(w, "Failed to login")
		return
	}
	if user == nil {
		utils.Unauthorized(w, "Invalid email or password")
		return
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		h.log.WithError(err).Warn("password verification failed")
		utils.Unauthorized(w, "Invalid email or password")
		return
	}

	accessToken, refreshToken, err :=
		h.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		h.log.WithError(err).Error("token generation failed")
		utils.InternalServerError(w, "Failed to login")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, models.AuthResponse{
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh handles token refresh requests
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := utils.ParseJSON(r, &req); err != nil {
		h.log.WithError(err).Error("invalid refresh payload")
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.BadRequest(w, "refresh_token is required")
		return
	}

	// Validate refresh token
	userIDStr, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.log.WithError(err).Warn("refresh token validation failed")
		utils.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.WithError(err).Error("invalid user ID in refresh token")
		utils.Unauthorized(w, "Invalid token")
		return
	}

	// Fetch user from database
	user, err := h.repo.User.GetByID(r.Context(), userID)
	if err != nil {
		h.log.WithError(err).Error("failed to fetch user during refresh")
		utils.InternalServerError(w, "Failed to refresh token")
		return
	}

	if user == nil {
		utils.Unauthorized(w, "User not found")
		return
	}

	// Generate new token pair
	accessToken, refreshToken, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		h.log.WithError(err).Error("token generation failed during refresh")
		utils.InternalServerError(w, "Failed to refresh token")
		return
	}

	// Return response matching existing AuthResponse format
	utils.JSONSuccess(w, http.StatusOK, models.AuthResponse{
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
