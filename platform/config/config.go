package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	App      AppConfig      `toml:"app"`
	Database DatabaseConfig `toml:"database"`
	Server   ServerConfig   `toml:"server"`
	JWT      JWTConfig      `toml:"jwt"`
}

type AppConfig struct {
	Name string `toml:"name"`
	Env  string `toml:"env"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
}

type JWTConfig struct {
	Secret string `toml:"secret"`
}

type ServerConfig struct {
	Port int `toml:"port"`
}

func Load(path string) (*Config, error) {
	env := getEnv("TODO_ENV", "development")
	filename := fmt.Sprintf("config.%s.toml", env)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		filename = "config.toml"
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", path, filename))
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse toml: %w", err)
	}

	return &cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}
