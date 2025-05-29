package config

import (
	"github.com/spf13/viper"
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
	viper.AutomaticEnv()
	viper.SetConfigFile("./config/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic("读取配置文件失败: " + err.Error())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("配置解析失败: " + err.Error())
	}

	return &cfg
}
