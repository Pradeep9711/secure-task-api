package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"secure-task-api/internal/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func RequestLoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &statusRecorder{
				ResponseWriter: w,
			}

			next.ServeHTTP(rec, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status", rec.status),
				zap.Duration("duration", time.Since(start)),
			}

			if userID, ok := GetUserIDFromContext(r.Context()); ok {
				fields = append(fields, zap.String("user_id", userID))
			}

			log.Info("http request completed", fields...)
		})
	}
}
