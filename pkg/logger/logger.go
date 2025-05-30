package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"modules/config"
	"os"
)

var Log *zap.Logger

func InitLogger(cfg *config.Config) {
	var zapCfg zap.Config
	switch cfg.Env {
	case "production":
		zapCfg = zap.NewProductionConfig()
	default:
		zapCfg = zap.NewDevelopmentConfig()
	}

	// 打开日志文件
	file, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// 创建一个将日志同时输出到控制台和文件的核心
	consoleEncoder := zapcore.NewConsoleEncoder(zapCfg.EncoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(zapCfg.EncoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapCfg.Level),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapCfg.Level),
	)

	Log = zap.New(core)
}
