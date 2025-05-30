package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Config struct {
	Env         string         `mapstructure:"env"`
	Port        string         `mapstructure:"port"`
	DB          DatabaseConfig `mapstructure:"db"`
	JWT         JWTConfig      `yaml:"jwt"`
	LogFilePath string
}

type JWTConfig struct {
	Secret    string        `yaml:"secret"`
	ExpiresIn time.Duration `yaml:"expires_in"`
	MaxAge    int           `yaml:"max_age"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	// 获取项目根目录
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	configPath := fmt.Sprintf("%s/config/config.yaml", dir)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}

	return &cfg, nil
}
