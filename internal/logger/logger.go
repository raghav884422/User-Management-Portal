package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new Uber Zap structured logger.
// In production it uses JSON encoding; otherwise uses a human-friendly console encoder.
func NewLogger(env string) *zap.Logger {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	log, err := cfg.Build()
	if err != nil {
		// If logger construction fails, fall back to a no-op logger.
		return zap.NewNop()
	}

	return log
}
