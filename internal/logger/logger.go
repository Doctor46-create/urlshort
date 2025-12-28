package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Doctor46-create/urlshort/internal/config"
)

func NewLogger(config *config.Config) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(config.LogLevel))
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	cfg.EncoderConfig.CallerKey = "logLine"

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger config: %w", err)
	}

	return logger, nil
}
