package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"secure-task-api/internal/auth"
	"secure-task-api/internal/config"
	"secure-task-api/internal/logger"
	"secure-task-api/internal/middleware"
	"secure-task-api/internal/repository"
)

type Router struct {
	config     *config.Config
	repo       *repository.Repository
	jwtManager *auth.JWTManager
	log        *logger.Logger
}

func NewRouter(
	config *config.Config,
	repo *repository.Repository,
	jwtManager *auth.JWTManager,
	log *logger.Logger,
) *Router {
	return &Router{
		config:     config,
		repo:       repo,
		jwtManager: jwtManager,
		log:        log,
	}
}

// Sets up all HTTP routes and middleware for the service.
func (r *Router) SetupRoutes() chi.Router {
	router := chi.NewRouter()

	// Request-scoped middleware applied globally.
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(NewStructuredLogger(r.log).Middleware)
	router.Use(chimiddleware.Recoverer)

	// System and diagnostic endpoints.
	systemHandler := NewSystemHandler(r.repo, r.log)
	router.Get("/health", systemHandler.HealthCheck)
	router.Get("/debug/panic", systemHandler.TriggerPanic)

	router.Route("/v1", func(v1 chi.Router) {

		// Public authentication endpoints.
		authHandler := NewAuthHandler(r.repo, r.jwtManager, r.log)
		v1.Route("/auth", authHandler.RegisterRoutes)

		// Routes that require a valid JWT.
		v1.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware(r.jwtManager, r.log))

			taskHandler := NewTaskHandler(r.repo, r.log)
			protected.Route("/tasks", taskHandler.RegisterRoutes)
		})
	})

	return router
}

// StructuredLogger adapts the internal logger to Chi middleware.
type StructuredLogger struct {
	log *logger.Logger
}

func NewStructuredLogger(log *logger.Logger) *StructuredLogger {
	return &StructuredLogger{log: log}
}

func (l *StructuredLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		l.log.RequestLogger(
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			ww.Status(),
			time.Since(start).Seconds()*1000,
		)
	})
}
