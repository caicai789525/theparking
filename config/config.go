package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type JWTConfig struct {
	Secret    string `yaml:"secret"`
	ExpiresIn string `yaml:"expires_in"`
	MaxAge    int    `yaml:"max_age"`
}

type Config struct {
	Env  string `yaml:"env"`
	Port string `yaml:"port"`
	DB   struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"db"`
	JWT         JWTConfig `yaml:"jwt"`
	LogFilePath string    `yaml:"log_file_path"` // 添加 LogFilePath 字段
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
