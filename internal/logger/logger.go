package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap logger with reasonable defaults
// Production logger is configured for JSON output, while development logger is more human-readable
func NewLogger(isProduction bool) (*zap.Logger, error) {
	var config zap.Config

	if isProduction {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return config.Build()
}

// NewNoOpLogger creates a logger that doesn't output anything
// Useful for testing where you want to suppress logging
func NewNoOpLogger() *zap.Logger {
	return zap.NewNop()
}
