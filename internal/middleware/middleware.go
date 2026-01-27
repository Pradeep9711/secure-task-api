package middleware

import (
	"context"
	"net/http"
	"strings"

	"secure-task-api/internal/auth"
	"secure-task-api/internal/logger"
)

// private context keys to avoid collisions with other packages
type contextKey string

const (
	userIDKey contextKey = "user_id"
	emailKey  contextKey = "email"
)

// AuthMiddleware validates the JWT and attaches user data to the request context
func AuthMiddleware(jwtManager *auth.JWTManager, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorized(w, "authorization header missing")
				return
			}

			token := bearerToken(authHeader)
			if token == "" {
				unauthorized(w, "invalid authorization header")
				return
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				log.WithError(err).Warn("token validation failed")
				unauthorized(w, "invalid or expired token")
				return
			}

			// store authenticated user data in context for downstream handlers
			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, emailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// StripTrailingSlash normalizes URLs like `/tasks/` to `/tasks`
func StripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			r.URL.Path = path[:len(path)-1]
		}
		next.ServeHTTP(w, r)
	})
}

// extracts the token part from `Authorization: Bearer <token>`
func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// sends a consistent unauthorized response
func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","message":"` + msg + `"}`))
}

// helper used by handlers to read user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// helper used by handlers to read email from context
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailKey).(string)
	return email, ok
}
