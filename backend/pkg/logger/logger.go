package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ordermonitor/internal/config"
)

type Logger = *zap.Logger

type Field = zap.Field

func New(cfg config.LoggerConfig) (*zap.Logger, error) {
	var zapCfg zap.Config
	mode := strings.ToLower(cfg.Mode)

	if mode == "console" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	if err := zapCfg.Level.UnmarshalText([]byte(cfg.Level)); err != nil {
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapCfg.Build(zap.AddCaller())
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

