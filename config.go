package golearn

import (
	"os"
)

// Config ...
type Config struct {
	Env             string   `json:"env"`
	Database        Database `json:"database"`
	DefaultLanguage string   `json:"default_language"`
}

// Database ...
type Database struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// ConfigFromEnv returns config based on environment variables
func ConfigFromEnv() *Config {
	cfg := &Config{}

	cfg.Env = os.Getenv("ENV")
	cfg.Database.Host = os.Getenv("DB_HOST")
	cfg.Database.Port = os.Getenv("DB_PORT")
	cfg.Database.Name = os.Getenv("DB_NAME")
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")

	cfg.DefaultLanguage = os.Getenv("DEFAULT_LANGUAGE")

	return cfg
}
