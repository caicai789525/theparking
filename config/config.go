package config

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"modules/pkg/logger"
	"os"
	"time"
)

type Config struct {
	Env  string         `mapstructure:"env"`
	Port string         `mapstructure:"port"`
	DB   DatabaseConfig `mapstructure:"db"`
	JWT  JWTConfig      `mapstructure:"jwt"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type JWTConfig struct {
	Secret    string        `mapstructure:"secret"`
	ExpiresIn time.Duration `mapstructure:"expires_in"`
	MaxAge    int           `mapstructure:"max_age"`
}

func LoadConfig() *Config {
	// 假设配置文件名为 config.yaml，路径为项目根目录下
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		// 使用 zap.Error 函数将 error 转换为 zap.Field 类型
		logger.Log.Error("无法读取配置文件", zap.Error(err))
		panic(fmt.Sprintf("无法读取配置文件: %v", err))
	}

	var cfg Config
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		// 使用 zap.Error 函数将 error 转换为 zap.Field 类型
		logger.Log.Error("解析配置文件失败", zap.Error(err))
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	logger.Log.Info("读取到的端口配置", zap.String("port", cfg.Port))

	// 确保端口配置正确
	if cfg.Port == "" {
		// 可以设置默认端口
		cfg.Port = "8080"
		logger.Log.Info("端口配置为空，使用默认端口", zap.String("port", cfg.Port))
	}

	return &cfg
}
