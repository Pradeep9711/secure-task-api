package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"secure-task-api/internal/models"
)

// JSONResponse sends a JSON response
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// JSONError sends an error response
func JSONError(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, models.ErrorResponse{
		Error:      http.StatusText(status),
		Message:    message,
		Timestamp:  time.Now(),
		StatusCode: status,
	})
}

// JSONSuccess sends a success response
func JSONSuccess(w http.ResponseWriter, status int, data interface{}) {
	JSONResponse(w, status, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ValidationError sends a validation error response
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Validation Error",
		"message": "Invalid input data",
		"errors":  errors,
	})
}

// Unauthorized sends an unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusUnauthorized, message)
}

// NotFound sends a not found response
func NotFound(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusNotFound, message)
}

// InternalServerError sends an internal server error response
func InternalServerError(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusInternalServerError, message)
}

// BadRequest sends a bad request response
func BadRequest(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusBadRequest, message)
}
