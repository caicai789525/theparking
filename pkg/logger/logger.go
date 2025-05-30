package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"modules/config"
	"os"
)

var Log *zap.Logger

func InitLogger(cfg *config.Config) {
	// 检查日志文件路径是否为空
	if cfg.LogFilePath == "" {
		panic("日志文件路径为空")
	}
	file, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(file),
			zap.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),

			zap.DebugLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(file), zapcore.AddSync(os.Stdout)),
			zap.InfoLevel,
		),
	)

	Log = zap.New(core)

	Log = zap.New(core, zap.AddCaller())
}
