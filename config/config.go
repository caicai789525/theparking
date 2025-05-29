package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Config struct {
	Env  string         `mapstructure:"env"`
	Port string         `mapstructure:"8080"`
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
	viper.AutomaticEnv()
	// 获取项目根目录
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("获取当前工作目录失败: %v", err))
	}
	configPath := fmt.Sprintf("%s/config/config.yaml", dir)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		panic("读取配置文件失败: " + err.Error())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("配置解析失败: " + err.Error())
	}

	return &cfg
}
