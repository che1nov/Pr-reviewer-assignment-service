package config

import "os"

type Config struct {
	LogLevel    string
	HTTPPort    string
	AdminToken  string
	UserToken   string
	DatabaseURL string
}

func Load() Config {
	return Config{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		HTTPPort:    fallback(os.Getenv("HTTP_PORT"), "8080"),
		AdminToken:  os.Getenv("ADMIN_TOKEN"),
		UserToken:   os.Getenv("USER_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}
