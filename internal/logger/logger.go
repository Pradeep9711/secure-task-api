package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"secure-task-api/internal/config"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(cfg config.LoggingConfig) (*Logger, error) {
	var zapConfig zap.Config

	if cfg.Level == "development" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Set encoding
	zapConfig.Encoding = cfg.Encoding

	// Set output paths
	if len(cfg.OutputPaths) > 0 {
		zapConfig.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	// Build logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.Logger.With(zap.Error(err))}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{l.Logger.With(zap.String("request_id", requestID))}
}

func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{l.Logger.With(zap.String("user_id", userID))}
}

// RequestLogger logs HTTP requests
func (l *Logger) RequestLogger(method, path, remoteAddr, userAgent string, status int, duration float64) {
	l.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("remote_addr", remoteAddr),
		zap.String("user_agent", userAgent),
		zap.Int("status", status),
		zap.Float64("duration_ms", duration),
	)
}
