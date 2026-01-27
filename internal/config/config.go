package config

import (
	"fmt"
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

	dsn string
}

func (d DatabaseConfig) DSN() string {
	if d.dsn != "" {
		return d.dsn
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

	// Load .env from deployments for local run
	v.SetConfigFile("deployments/.env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // ignore if not found

	// Load YAML config
	v.AddConfigPath("configs")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	_ = v.ReadInConfig() // ignore if not found

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("app.port", 8080)
	v.SetDefault("app.environment", "development")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.encoding", "json")

	// Bind environment variables
	_ = v.BindEnv("APP_PORT")
	_ = v.BindEnv("APP_ENV")
	_ = v.BindEnv("DB_HOST")
	_ = v.BindEnv("DB_PORT")
	_ = v.BindEnv("DB_USER")
	_ = v.BindEnv("DB_PASSWORD")
	_ = v.BindEnv("DB_NAME")
	_ = v.BindEnv("JWT_SECRET")
	_ = v.BindEnv("SENTRY_DSN")
	_ = v.BindEnv("LOG_LEVEL")
	_ = v.BindEnv("DATABASE_URL") // important

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config parsing failed: %w", err)
	}

	// If DATABASE_URL is set, use it as DSN
	if databaseURL := v.GetString("DATABASE_URL"); databaseURL != "" {
		cfg.Database.dsn = databaseURL
	}

	cfg.App.ReadTimeout = time.Duration(v.GetInt("app.read_timeout")) * time.Second
	cfg.App.WriteTimeout = time.Duration(v.GetInt("app.write_timeout")) * time.Second
	cfg.App.IdleTimeout = time.Duration(v.GetInt("app.idle_timeout")) * time.Second
	cfg.Database.ConnMaxLifetime = time.Duration(v.GetInt("database.conn_max_lifetime")) * time.Minute
	cfg.JWT.AccessTokenDuration = v.GetDuration("jwt.access_token_duration")
	cfg.JWT.RefreshTokenDuration = v.GetDuration("jwt.refresh_token_duration")

	return &cfg, nil
}
