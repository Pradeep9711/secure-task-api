package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Sentry   SentryConfig
	Logging  LoggingConfig
}

type AppConfig struct {
	Name         string
	Version      string
	Port         int
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	DSN             string // Full connection string override
}

func (d DatabaseConfig) GetDSN() string {
	// If full DATABASE_URL is provided, use it
	if d.DSN != "" {
		return d.DSN
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.DBName,
		d.SSLMode,
	)
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type SentryConfig struct {
	DSN         string
	Environment string
	SampleRate  float64
}

type LoggingConfig struct {
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Use environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_ENVIRONMENT", "development")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_SSLMODE", "require") // Render requires SSL
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_ENCODING", "json")

	// Helper to get env var with fallback
	getEnv := func(key, fallback string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return fallback
	}

	// Parse durations safely
	parseDuration := func(val string, defaultVal time.Duration) time.Duration {
		if val == "" {
			return defaultVal
		}
		d, err := time.ParseDuration(val)
		if err != nil {
			return defaultVal
		}
		return d
	}

	// Build config explicitly from environment variables
	cfg := &Config{
		App: AppConfig{
			Name:         getEnv("APP_NAME", "Secure Task Management API"),
			Version:      getEnv("APP_VERSION", "1.0.0"),
			Port:         v.GetInt("APP_PORT"),
			Environment:  getEnv("APP_ENVIRONMENT", "development"),
			ReadTimeout:  parseDuration(os.Getenv("APP_READ_TIMEOUT"), 15*time.Second),
			WriteTimeout: parseDuration(os.Getenv("APP_WRITE_TIMEOUT"), 15*time.Second),
			IdleTimeout:  parseDuration(os.Getenv("APP_IDLE_TIMEOUT"), 60*time.Second),
		},
		Database: DatabaseConfig{
			// Check for DATABASE_URL first (Render provides this)
			DSN:             getEnv("DATABASE_URL", ""),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            v.GetInt("DB_PORT"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSLMODE", "require"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: parseDuration(os.Getenv("DB_CONN_MAX_LIFETIME"), 30*time.Minute),
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""),
			AccessTokenDuration:  parseDuration(os.Getenv("JWT_ACCESS_DURATION"), 15*time.Minute),
			RefreshTokenDuration: parseDuration(os.Getenv("JWT_REFRESH_DURATION"), 7*24*time.Hour),
		},
		Sentry: SentryConfig{
			DSN:         getEnv("SENTRY_DSN", ""),
			Environment: getEnv("SENTRY_ENVIRONMENT", getEnv("APP_ENVIRONMENT", "development")),
			SampleRate:  v.GetFloat64("SENTRY_SAMPLE_RATE"),
		},
		Logging: LoggingConfig{
			Level:            getEnv("LOG_LEVEL", "info"),
			Encoding:         getEnv("LOG_ENCODING", "json"),
			OutputPaths:      strings.Split(getEnv("LOG_OUTPUT_PATHS", "stdout"), ","),
			ErrorOutputPaths: strings.Split(getEnv("LOG_ERROR_OUTPUT_PATHS", "stderr"), ","),
		},
	}

	// Validate required fields
	if cfg.Database.DSN == "" && (cfg.Database.Host == "" || cfg.Database.User == "") {
		return nil, fmt.Errorf("database configuration missing: set DATABASE_URL or DB_HOST/DB_USER/DB_PASSWORD/DB_NAME")
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
