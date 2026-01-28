package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"secure-task-api/internal/auth"
	"secure-task-api/internal/config"
	"secure-task-api/internal/handlers"
	"secure-task-api/internal/logger"
	"secure-task-api/internal/middleware"
	"secure-task-api/internal/repository"
)

func main() {
	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Secure Task Management API",
		zap.String("app", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize Sentry if DSN is provided
	if cfg.Sentry.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.Sentry.DSN,
			Environment: cfg.Sentry.Environment,
			SampleRate:  cfg.Sentry.SampleRate,
		}); err != nil {
			log.Error("Failed to initialize Sentry", zap.Error(err))
		} else {
			defer sentry.Flush(2 * time.Second)
		}
	}

	// Connect to database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	log.Info("Database connection established")

	// Setup repository
	repo := repository.NewRepository(db)

	// Setup JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	// Setup routes
	router := handlers.NewRouter(cfg, repo, jwtManager, log).SetupRoutes()

	// Wrap router with middleware
	handler := middleware.StripTrailingSlash(
		middleware.RequestLoggingMiddleware(log)(router),
	)

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	// Start server
	go func() {
		log.Info("Server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown forced", zap.Error(err))
	}

	log.Info("Server exited cleanly")
}

// initDatabase connects to PostgreSQL and configures the pool
func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	var db *sql.DB
	var err error
	maxRetries := 5
	retryDelay := 2 * time.Second

	dsn := cfg.GetDSN()

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = db.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
		}

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("database connection failed after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	return db, nil
}
