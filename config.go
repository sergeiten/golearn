package golearn

import (
	"os"
	"strconv"
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
	Delay    int    `json:"delay"`
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

	delay, err := strconv.Atoi(os.Getenv("DB_DELAY"))
	if err != nil {
		LogPrint(err, "failed to convert delay env")
	}
	cfg.Database.Delay = delay

	cfg.DefaultLanguage = os.Getenv("DEFAULT_LANGUAGE")

	return cfg
}
